package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		if m.palette != nil {
			w := paletteInputWidth(m.contentWidth())
			m.palette.input.Width = w
		}
		return m, m.syncSelectionPreview()

	case tea.KeyMsg:
		if m.state == stateInstalling && m.installFlow != nil {
			return m.handleInstallingUpdate(msg)
		}
		if m.prompt != nil {
			return m.handlePromptKeys(msg)
		}
		if m.state == stateCommandPalette && m.palette != nil {
			return m.handlePaletteKeys(msg)
		}
		if m.state == stateHelpOverlay {
			return m.handleHelpOverlayKeys(msg)
		}
		if m.state == stateConfirming {
			return m.handleConfirmKeys(msg)
		}
		if m.state == stateBindingAgent {
			return m.handleBindingKeys(msg)
		}
		if m.state == stateFilteringAgent {
			return m.handleAgentFilterUpdate(msg)
		}
		if m.state == stateInspecting {
			return m.handleInspectingKeys(msg)
		}
		if m.listFilterActive() {
			return m.handleListFilterKeys(msg)
		}
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		if m.state == stateHelpOverlay {
			return m.handleHelpOverlayMouse(msg)
		}
		return m.handleMouseMsg(msg)

	case panel.SkillsScannedMsg:
		if msg.Err != nil {
			m.reportError(msg.Err)
			return m, nil
		}
		m.panels.Get(panel.TabSkills).ApplyScan(msg)
		m.status = "ready"
		m.clearError()
		if m.state == stateInstalling && m.installFlow != nil {
			m.status = "ready"
			return m, nil
		}
		if m.activeTab == panel.TabSkills && (m.state == stateHome || m.state == stateListing || m.state == stateSearching) {
			m.refreshActiveList()
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case panel.MCPScannedMsg:
		if msg.Err != nil {
			m.reportError(msg.Err)
			return m, nil
		}
		m.panels.Get(panel.TabMCP).ApplyScan(msg)
		m.status = "ready"
		m.clearError()
		if m.activeTab == panel.TabMCP && (m.state == stateHome || m.state == stateListing || m.state == stateSearching) {
			m.refreshActiveList()
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case panel.PreviewLoadedMsg:
		if msg.Gen != m.previewGen || msg.Tab != m.activeTab {
			return m, nil
		}
		if msg.Err != nil {
			m.preview.SetContent("Preview failed:\n\n" + msg.Err.Error())
			return m, nil
		}
		m.previewBody = msg.Content
		m.preview.SetContent(msg.Content)
		return m, nil

	case mutationCompletedMsg:
		if msg.err != nil {
			m.reportError(msg.err)
			return m, m.scanAllCmd()
		}
		m.clearError()
		m.status = "ready"
		var flashCmd tea.Cmd
		if msg.message != "" {
			flashCmd = m.flashFooter(msg.message)
		}
		if msg.selectName != "" {
			if msg.targetTab == panel.TabMCP {
				return m, tea.Batch(flashCmd, tea.Sequence(
					m.scanAllCmd(),
					func() tea.Msg { return reselectMCPMsg{name: msg.selectName} },
				))
			}
			return m, tea.Batch(flashCmd, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: msg.selectName} },
			))
		}
		return m, tea.Batch(flashCmd, m.scanAllCmd())

	case reselectMCPMsg:
		if m.selectMCPByName(msg.name) {
			return m, tea.Batch(m.flashFooter(fmt.Sprintf("selected MCP %s", msg.name)), m.syncSelectionPreview())
		}
		return m, nil

	case reselectSkillMsg:
		if m.selectSkillByName(msg.name) {
			return m, tea.Batch(m.flashFooter(fmt.Sprintf("selected %s", msg.name)), m.syncSelectionPreview())
		}
		return m, nil

	case footerFlashTimeoutMsg:
		return m.handleFooterFlashTimeout(msg)

	case installSearchCompletedMsg:
		return m.handleInstallingUpdate(msg)

	case installProgressTickMsg:
		return m.handleInstallProgressTick()

	case themeDetectedMsg:
		return m.handleThemeDetected(msg)

	case installCompletedMsg:
		if m.installCancel != nil {
			m.installCancel = nil
		}
		if errors.Is(msg.err, context.Canceled) {
			m.clearInstallFlow()
			m.state = m.lastState
			m.status = "ready"
			m.setFooterContext("Install cancelled")
			return m, nil
		}
		m.clearInstallFlow()
		m.state = stateListing
		m.lastState = stateListing
		if msg.err != nil {
			m.reportError(msg.err)
			return m, m.scanAllCmd()
		}
		m.clearError()
		m.status = "ready"
		return m, tea.Batch(
			m.flashFooter(fmt.Sprintf("installed %s", msg.name)),
			tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: msg.name} },
			),
		)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.state == stateInstalling && m.installFlow != nil && m.installFlow.searching {
			return m, cmd
		}
		return m, cmd

	case progress.FrameMsg:
		if m.state == stateInstalling && m.installFlow != nil && m.installFlow.installing {
			next, cmd := m.installFlow.progress.Update(msg)
			m.installFlow.progress = next.(progress.Model)
			return m, cmd
		}
		return m, nil
	}

	if m.state == stateInstalling && m.installFlow != nil {
		model, cmd := m.handleInstallingUpdate(msg)
		m.syncInstallHint()
		return model, cmd
	}

	var (
		listCmd    tea.Cmd
		previewCmd tea.Cmd
	)
	m.list, listCmd = m.list.Update(msg)
	m.preview, previewCmd = m.preview.Update(msg)

	return m, tea.Batch(listCmd, previewCmd)
}

// --- key dispatch -----------------------------------------------------------

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		if m.state == stateInstalling && m.installFlow != nil && m.installFlow.installing {
			return m.handleInstallQuitAttempt()
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Home):
		m.clearError()
		if m.list.FilterState() != list.Unfiltered {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		m.state = stateHome
		m.lastState = stateHome
		m.refreshActiveList()
		if preview := m.activePanel().StaticPreview(); preview != "" {
			m.preview.SetContent(preview)
			return m, nil
		}
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.HelpToggle):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil

	case key.Matches(msg, keys.HelpScreen):
		return m.handleHelp()

	case key.Matches(msg, keys.Palette):
		return m.openCommandPalette()

	case key.Matches(msg, keys.Tab):
		m.focusedPane = focusPaneList
		return m, m.switchExtensionTab(false)

	case key.Matches(msg, keys.ShiftTab):
		m.focusedPane = focusPaneList
		return m, m.switchExtensionTab(true)

	case key.Matches(msg, keys.Down):
		m.focusedPane = focusPaneList
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.Up):
		m.focusedPane = focusPaneList
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		m.focusedPane = focusPanePreview
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	case key.Matches(msg, keys.List):
		return m.handleList()

	case key.Matches(msg, keys.Find), key.Matches(msg, keys.Filter):
		return m.startListFilter()

	case key.Matches(msg, keys.Agent):
		return m.handleOpenAgentFilter()

	case key.Matches(msg, keys.Reload):
		return m.handleReload()

	case key.Matches(msg, keys.Update):
		return m.handleUpdate()

	case key.Matches(msg, keys.Enter):
		return m.handleInspectSelected()

	case key.Matches(msg, keys.Bind):
		return m.handleBindSelected()

	case key.Matches(msg, keys.Disable):
		return m.handleDisableSelected()

	case key.Matches(msg, keys.Delete):
		return m.handleRemoveSelected()

	case key.Matches(msg, keys.Add):
		return m.startInstallFlow()

	case key.Matches(msg, keys.Init):
		if m.activeTab == panel.TabSkills && m.activePanel().Capabilities().Init {
			return m.showInitPrompt()
		}
		m.setFooterContext("Init is only available on the Skills tab")
		return m, nil
	}

	return m, nil
}

// --- command handlers -------------------------------------------------------

func (m *Model) handleHelp() (tea.Model, tea.Cmd) {
	return m.openHelpOverlay()
}

func (m *Model) handleList() (tea.Model, tea.Cmd) {
	m.state = stateListing
	m.lastState = stateListing
	m.setFooterContext(fmt.Sprintf("%d %s · agents: %s", m.activePanel().Count(), m.activePanel().CountLabel(), m.agentDisplay()))
	m.refreshActiveList()
	return m, m.syncSelectionPreview()
}

func (m *Model) handleReload() (tea.Model, tea.Cmd) {
	m.status = "loading"
	m.setFooterContext(m.activePanel().ReloadHint())
	return m, m.scanAllCmd()
}

func (m *Model) handleInspectSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Inspect {
		m.setFooterContext("Inspect is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		m.lastState = m.state
		m.state = stateInspecting
		m.tree.setRoot(selected.skill.Path)
		m.setFooterContext("Inspecting skill files")

		// load preview for the selected tree item if it's a file
		sel := m.tree.SelectedItem()
		if sel.path != "" && !sel.isDir {
			m.status = "loading"
			return m, m.previewFileCmd(sel.path)
		}
	}
	if ok && selected.kind == itemKindMCP && selected.mcp != nil {
		width := m.preview.Width
		if width == 0 {
			width = max(40, m.width/2)
		}
		return m, m.activePanel().SyncPreview(listItemToPanel(selected), width, &m.previewGen)
	}
	return m, nil
}

func (m *Model) handleDisableSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Disable {
		m.setFooterContext("Disable is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.setFooterContext("Select an item first, then press 'x' to toggle disable")
		return m, nil
	}
	if selected.kind == itemKindMCP && selected.mcp != nil {
		srv := selected.mcp
		m.status = "loading"
		action := "Disabling"
		if srv.AggregatedDisabled() {
			action = "Enabling"
		}
		m.setFooterContext(fmt.Sprintf("%s MCP %s...", action, srv.GetName()))
		return m, m.toggleDisableMCPCmd(srv)
	}
	if selected.kind != itemKindSkill {
		m.setFooterContext("Select a skill or MCP server first")
		return m, nil
	}
	skill := selected.skill
	m.status = "loading"
	action := "Disabling"
	if skill.IsDisabled() {
		action = "Enabling"
	}
	m.setFooterContext(fmt.Sprintf("%s %s...", action, skill.GetName()))
	return m, m.toggleDisableSkillCmd(skill)
}

func (m *Model) handleRemoveSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Remove {
		m.setFooterContext("Remove is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.setFooterContext("Select an item first, then press Delete to remove")
		return m, nil
	}
	if selected.kind == itemKindMCP && selected.mcp != nil {
		srv := selected.mcp
		m.pending = &pendingAction{name: "remove", mcpName: srv.GetName(), mcp: srv}
		m.lastState = m.state
		m.state = stateConfirming
		return m, nil
	}
	if selected.kind != itemKindSkill {
		m.setFooterContext("Select a skill or MCP server first")
		return m, nil
	}
	skill := selected.skill
	m.pending = &pendingAction{name: "remove", skillName: skill.GetName(), skill: skill}
	m.lastState = m.state
	m.state = stateConfirming
	return m, nil
}

func (m *Model) handleUpdate() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Update {
		m.setFooterContext("Update is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		skill := selected.skill
		m.status = "loading"
		m.setFooterContext(fmt.Sprintf("Updating %s...", skill.GetName()))
		return m, m.updateSkillCmd(skill)
	}
	m.status = "loading"
	m.setFooterContext("Updating all managed local skills...")
	return m, m.updateAllSkillsCmd()
}

// --- prompt handlers --------------------------------------------------------

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Find {
		m.setFooterContext("Find is not available for this tab")
		return m, nil
	}
	return m, m.showPrompt("Find", "search query...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		text = strings.TrimSpace(text)
		m.state = stateSearching
		m.lastState = stateSearching
		if text == "" {
			m.refreshActiveList()
			return tea.Batch(m.flashFooter("Search cancelled"), m.syncSelectionPreview())
		}
		items := m.activePanel().SearchItems(text, m.agentIDs)
		m.setFooterContext(fmt.Sprintf("find: %q → %d result(s)", text, len(items)))
		m.setMainListItems(panelToListItems(items))
		return m.syncSelectionPreview()
	})
}

func (m *Model) showAddPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Add source", "path or SKILL.md ...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		source := strings.TrimSpace(text)
		if source == "" {
			return m.flashFooter("Add cancelled")
		}
		m.status = "loading"
		m.setFooterContext(fmt.Sprintf("Installing from %s...", source))
		return m.addSkillCmd(source)
	})
}

func (m *Model) showInitPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Init name", "new-skill (enter for default)", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		name := strings.TrimSpace(text)
		if name == "" {
			name = "new-skill"
		}
		m.status = "loading"
		m.setFooterContext(fmt.Sprintf("Creating skill template: %s", name))
		return m.initSkillCmd(name)
	})
}

func (m *Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.hidePrompt()
		if m.state == stateInstalling {
			m.cancelInstallFlow("Install cancelled")
			return m, nil
		}
		m.setFooterContext("Cancelled")
		return m, nil
	case key.Matches(msg, keys.Enter):
		text := strings.TrimSpace(m.prompt.input.Value())
		cmd := m.prompt.action(m, text)
		return m, cmd
	}
	var cmd tea.Cmd
	m.prompt.input, cmd = m.prompt.input.Update(msg)
	return m, cmd
}

// --- confirmation -----------------------------------------------------------

func (m *Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		if m.pending != nil && m.pending.name == "remove" {
			if m.pending.mcp != nil {
				srv := m.pending.mcp
				m.pending = nil
				m.state = m.lastState
				m.status = "loading"
				m.setFooterContext(fmt.Sprintf("Removing MCP %s...", srv.GetName()))
				return m, m.removeMCPCmd(srv)
			}
			skill := m.pending.skill
			m.pending = nil
			m.state = m.lastState
			m.status = "loading"
			m.setFooterContext(fmt.Sprintf("Removing %s...", skill.GetName()))
			return m, m.removeSkillCmd(skill)
		}
		m.pending = nil
		m.state = m.lastState
		return m, nil
	case key.Matches(msg, keys.Cancel):
		m.pending = nil
		m.state = m.lastState
		m.setFooterContext("Cancelled")
		return m, nil
	default:
		return m, nil
	}
}

// --- helpers ----------------------------------------------------------------

func (m *Model) setAgentFilter(id string) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" || id == "all" {
		m.agentIDs = []string{"all"}
		return
	}
	if _, ok := agent.AgentByID(id); ok {
		m.agentIDs = []string{id}
		return
	}

}

const helpPreview = `# skill-man

Keybindings:

- Tab / Shift+Tab: switch Skills and MCP tabs
- ?: expand footer keys · F1: command list
- Enter: inspect skill (open file tree)
- x: toggle disable/enable for selected skill
- b: bind/unbind skill to specific agents
- Delete: remove selected skill (with confirmation)
- Ctrl+L: list skills
- Ctrl+F: find skills (prompt)
- Ctrl+A: open agent filter dialog
- Ctrl+D: open Search & Install dialog (skills.sh registry)
- Ctrl+N: create new skill template (prompt)
- Ctrl+R: reload/rescan skills
- Ctrl+U: update skill (selected or all)
- Ctrl+J / Down, Ctrl+K / Up: navigate list
- PgUp / PgDn: scroll preview
- Esc: home / cancel
- Ctrl+C: quit

Prompts appear at the bottom for commands that need text input.
Press Enter to confirm, Esc to cancel.`

func (m *Model) handleBindSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Bind {
		m.setFooterContext("Bind is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.setFooterContext("Select an item first to manage agent bindings")
		return m, nil
	}

	m.lastState = m.state
	m.state = stateBindingAgent
	m.syncBindHint()

	if selected.kind == itemKindMCP && selected.mcp != nil {
		m.bindingSkill = nil
		m.bindingMCP = selected.mcp
		m.bindingAgents = newMCPBindChoices(selected.mcp, m.cwd, m.home)
		m.setAgentListItems(bindChoicesToListItems(m.bindingAgents, m.cwd, m.home))
		m.agentList.Select(0)
		return m, nil
	}

	if selected.kind != itemKindSkill {
		m.state = m.lastState
		return m, m.flashFooter("Select a skill or MCP server first")
	}

	m.bindingMCP = nil
	m.bindingSkill = selected.skill
	m.bindingAgents = newSkillBindChoices(selected.skill)
	m.setAgentListItems(bindChoicesToListItems(m.bindingAgents, m.cwd, m.home))
	m.agentList.Select(0)
	return m, nil
}

func (m *Model) handleBindingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.bindingMCP != nil {
			srv := m.bindingMCP
			if err := applyMCPBindChoices(m.mcpManager, srv, m.bindingAgents, m.cwd, m.home); err != nil {
				m.reportError(err)
			}
			m.clearBindingSession()
			m.state = m.lastState
			var cmds []tea.Cmd
			if m.errMsg == "" {
				cmds = append(cmds, m.flashFooter(fmt.Sprintf("Updated MCP bindings for %s", srv.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectMCPMsg{name: srv.GetName()} },
			))
			return m, tea.Batch(cmds...)
		}
		if m.bindingSkill != nil {
			skill := m.bindingSkill
			if err := applySkillBindChoices(context.Background(), m.skillManager, skill, m.bindingAgents, m.cwd, m.home); err != nil {
				m.reportError(err)
			}
			m.clearBindingSession()
			m.state = m.lastState
			var cmds []tea.Cmd
			if m.errMsg == "" {
				cmds = append(cmds, m.flashFooter(fmt.Sprintf("Updated agent bindings for %s", skill.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: skill.GetName()} },
			))
			return m, tea.Batch(cmds...)
		}
		m.state = m.lastState
		return m, nil

	case key.Matches(msg, keys.Cancel):
		m.clearBindingSession()
		m.state = m.lastState
		return m, m.flashFooter("Agent binding cancelled")

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		if idx >= 0 && idx < len(m.bindingAgents) {
			m.bindingAgents[idx].desired = !m.bindingAgents[idx].desired
			m.setAgentListItems(bindChoicesToListItems(m.bindingAgents, m.cwd, m.home))
			m.agentList.Select(idx)
			m.syncBindHint()
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) handleInspectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.state = m.lastState
		return m, tea.Batch(m.flashFooter("Returned to skill list"), m.syncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd
	}

	oldSelected := m.tree.SelectedItem()
	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	newSelected := m.tree.SelectedItem()

	if newSelected.path != "" && newSelected.path != oldSelected.path && !newSelected.isDir {
		// load preview for the file
		m.status = "loading"
		return m, tea.Batch(cmd, m.previewFileCmd(newSelected.path))
	}

	return m, cmd
}

func (m *Model) previewFileCmd(path string) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	m.previewGen++
	gen := m.previewGen
	return func() tea.Msg {
		// Reuse RenderSkillPreview but with a dummy skill that points to this file
		dummy := skilldomain.Skill{
			BaseExtension: extension.BaseExtension{
				Name:       filepath.Base(path),
				ConfigPath: path,
			},
		}
		content, err := service.RenderSkillPreview(dummy, width)
		return panel.PreviewLoadedMsg{Tab: m.activeTab, Content: content, Err: err, Gen: gen}
	}
}
