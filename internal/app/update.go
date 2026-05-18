package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
		return m, m.syncSelectionPreview()

	case tea.KeyMsg:
		if m.prompt != nil {
			return m.handlePromptKeys(msg)
		}
		if m.state == stateConfirming {
			return m.handleConfirmKeys(msg)
		}
		if m.state == stateBindingAgent {
			return m.handleBindingKeys(msg)
		}
		if m.state == stateInspecting {
			return m.handleInspectingKeys(msg)
		}
		return m.handleKeyMsg(msg)

	case panel.SkillsScannedMsg:
		if msg.Err != nil {
			m.reportError(msg.Err)
			return m, nil
		}
		m.panels.Get(panel.TabSkills).ApplyScan(msg)
		m.status = "ready"
		m.clearError()
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
		if msg.message != "" {
			m.hint = msg.message
		}
		if msg.selectName != "" {
			if msg.targetTab == panel.TabMCP {
				return m, tea.Sequence(
					m.scanAllCmd(),
					func() tea.Msg { return reselectMCPMsg{name: msg.selectName} },
				)
			}
			return m, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: msg.selectName} },
			)
		}
		return m, m.scanAllCmd()

	case reselectMCPMsg:
		if m.selectMCPByName(msg.name) {
			m.hint = fmt.Sprintf("selected MCP %s", msg.name)
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case reselectSkillMsg:
		if m.selectSkillByName(msg.name) {
			m.hint = fmt.Sprintf("selected %s", msg.name)
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var (
		listCmd    tea.Cmd
		previewCmd tea.Cmd
	)
	m.list, listCmd = m.list.Update(msg)
	m.preview, previewCmd = m.preview.Update(msg)

	m.updateHint()

	return m, tea.Batch(listCmd, previewCmd)
}

// --- key dispatch -----------------------------------------------------------

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Home):
		m.clearError()
		m.state = stateHome
		m.lastState = stateHome
		m.refreshActiveList()
		if preview := m.activePanel().StaticPreview(); preview != "" {
			m.preview.SetContent(preview)
			return m, nil
		}
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.Help):
		return m.handleHelp()

	case key.Matches(msg, keys.Tab):
		return m, m.switchExtensionTab(false)

	case key.Matches(msg, keys.ShiftTab):
		return m, m.switchExtensionTab(true)

	case key.Matches(msg, keys.Down):
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.Up):
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	case key.Matches(msg, keys.List):
		return m.handleList()

	case key.Matches(msg, keys.Find):
		return m.showFindPrompt()

	case key.Matches(msg, keys.Agent):
		return m.handleCycleAgent()

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
	}

	return m, nil
}

// --- command handlers -------------------------------------------------------

func (m *Model) handleHelp() (tea.Model, tea.Cmd) {
	m.state = stateViewingHelp
	m.lastState = stateViewingHelp
	m.setCommandItems()
	m.preview.SetContent(helpPreview)
	m.hint = "Ctrl+A:cycle agent  Ctrl+L:list  Ctrl+R:reload  Esc:back"
	return m, nil
}

func (m *Model) handleList() (tea.Model, tea.Cmd) {
	m.state = stateListing
	m.lastState = stateListing
	m.hint = fmt.Sprintf("%d %s | agents: %s | %s", m.activePanel().Count(), m.activePanel().CountLabel(), m.agentDisplay(), m.cwd)
	m.refreshActiveList()
	return m, m.syncSelectionPreview()
}

func (m *Model) handleReload() (tea.Model, tea.Cmd) {
	m.status = "loading"
	m.hint = m.activePanel().ReloadHint()
	return m, m.scanAllCmd()
}

func (m *Model) handleInspectSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Inspect {
		m.hint = "Inspect is not available for this tab"
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		m.lastState = m.state
		m.state = stateInspecting
		m.tree.setRoot(selected.skill.Path)
		m.hint = "Space/Enter: toggle folder | Esc: back to skills"

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
		m.hint = "Disable is not available for this tab"
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.hint = "Select an item first, then press 'x' to toggle disable"
		return m, nil
	}
	if selected.kind == itemKindMCP && selected.mcp != nil {
		srv := selected.mcp
		m.status = "loading"
		action := "Disabling"
		if srv.AggregatedDisabled() {
			action = "Enabling"
		}
		m.hint = fmt.Sprintf("%s MCP %s...", action, srv.GetName())
		return m, m.toggleDisableMCPCmd(srv)
	}
	if selected.kind != itemKindSkill {
		m.hint = "Select a skill or MCP server first"
		return m, nil
	}
	skill := selected.skill
	m.status = "loading"
	action := "Disabling"
	if skill.IsDisabled() {
		action = "Enabling"
	}
	m.hint = fmt.Sprintf("%s %s...", action, skill.GetName())
	return m, m.toggleDisableSkillCmd(skill)
}

func (m *Model) handleRemoveSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Remove {
		m.hint = "Remove is not available for this tab"
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.hint = "Select an item first, then press Delete to remove"
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
		m.hint = "Select a skill or MCP server first"
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
		m.hint = "Update is not available for this tab"
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		skill := selected.skill
		m.status = "loading"
		m.hint = fmt.Sprintf("Updating %s...", skill.GetName())
		return m, m.updateSkillCmd(skill)
	}
	m.status = "loading"
	m.hint = "Updating all managed local skills..."
	return m, m.updateAllSkillsCmd()
}

func (m *Model) handleCycleAgent() (tea.Model, tea.Cmd) {
	allIDs := []string{"all"}
	for _, a := range m.allAgents {
		allIDs = append(allIDs, a.ID)
	}
	current := "all"
	if len(m.agentIDs) > 0 {
		current = m.agentIDs[0]
	}
	for i, id := range allIDs {
		if strings.EqualFold(id, current) {
			next := allIDs[(i+1)%len(allIDs)]
			m.setAgentFilter(next)
			break
		}
	}
	m.hint = fmt.Sprintf("Agent filter: %s", m.agentDisplay())
	if m.activeTab != panel.TabSkills {
		return m, nil
	}
	m.refreshActiveList()
	return m, m.syncSelectionPreview()
}

// --- prompt handlers --------------------------------------------------------

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Find {
		m.hint = "Find is not available for this tab"
		return m, nil
	}
	return m, m.showPrompt("Find", "search query...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		text = strings.TrimSpace(text)
		m.state = stateSearching
		m.lastState = stateSearching
		if text == "" {
			m.hint = "Search cancelled"
			m.refreshActiveList()
			return m.syncSelectionPreview()
		}
		items := m.activePanel().SearchItems(text, m.agentIDs)
		m.hint = fmt.Sprintf("find: %q -> %d result(s)", text, len(items))
		m.setMainListItems(panelToListItems(items))
		return m.syncSelectionPreview()
	})
}

func (m *Model) showAddPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Add source", "path or SKILL.md ...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		source := strings.TrimSpace(text)
		if source == "" {
			m.hint = "Add cancelled"
			return nil
		}
		m.status = "loading"
		m.hint = fmt.Sprintf("Installing from %s...", source)
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
		m.hint = fmt.Sprintf("Creating skill template: %s", name)
		return m.initSkillCmd(name)
	})
}

func (m *Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.hidePrompt()
		m.hint = "Cancelled"
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
				m.hint = fmt.Sprintf("Removing MCP %s...", srv.GetName())
				return m, m.removeMCPCmd(srv)
			}
			skill := m.pending.skill
			m.pending = nil
			m.state = m.lastState
			m.status = "loading"
			m.hint = fmt.Sprintf("Removing %s...", skill.GetName())
			return m, m.removeSkillCmd(skill)
		}
		m.pending = nil
		m.state = m.lastState
		return m, nil
	case key.Matches(msg, keys.Cancel):
		m.pending = nil
		m.state = m.lastState
		m.hint = "Cancelled"
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
	m.hint = fmt.Sprintf("Unknown agent: %s", id)
}

const helpPreview = `# skill-man

Keybindings:

- Tab / Shift+Tab: switch Skills and MCP tabs
- ? / F1: show this help
- Enter: inspect skill (open file tree)
- x: toggle disable/enable for selected skill
- b: bind/unbind skill to specific agents
- Delete: remove selected skill (with confirmation)
- Ctrl+L: list skills
- Ctrl+F: find skills (prompt)
- Ctrl+A: cycle agent filter
- Ctrl+D: add/install skill (prompt)
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
		m.hint = "Bind is not available for this tab"
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.hint = "Select an item first to manage agent bindings"
		return m, nil
	}

	m.lastState = m.state
	m.state = stateBindingAgent
	m.hint = "Space: toggle each agent (multi-select) | Enter: apply all | Esc: cancel"

	if selected.kind == itemKindMCP && selected.mcp != nil {
		m.bindingSkill = nil
		m.bindingMCP = selected.mcp
		m.bindingAgents = newMCPBindChoices(selected.mcp)
		m.agentList.SetItems(bindChoicesToListItems(m.bindingAgents))
		m.agentList.Select(0)
		return m, nil
	}

	if selected.kind != itemKindSkill {
		m.hint = "Select a skill or MCP server first"
		m.state = m.lastState
		return m, nil
	}

	m.bindingMCP = nil
	m.bindingSkill = selected.skill
	m.bindingAgents = newSkillBindChoices(selected.skill)
	m.agentList.SetItems(bindChoicesToListItems(m.bindingAgents))
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
			if m.errMsg == "" {
				m.hint = fmt.Sprintf("Updated MCP bindings for %s", srv.GetName())
			}
			return m, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectMCPMsg{name: srv.GetName()} },
			)
		}
		if m.bindingSkill != nil {
			skill := m.bindingSkill
			if err := applySkillBindChoices(context.Background(), m.skillManager, skill, m.bindingAgents, m.cwd, m.home); err != nil {
				m.reportError(err)
			}
			m.clearBindingSession()
			m.state = m.lastState
			if m.errMsg == "" {
				m.hint = fmt.Sprintf("Updated agent bindings for %s", skill.GetName())
			}
			return m, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: skill.GetName()} },
			)
		}
		m.state = m.lastState
		return m, nil

	case key.Matches(msg, keys.Cancel):
		m.clearBindingSession()
		m.state = m.lastState
		m.hint = "Agent binding cancelled"
		return m, nil

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		if idx >= 0 && idx < len(m.bindingAgents) {
			m.bindingAgents[idx].desired = !m.bindingAgents[idx].desired
			cmd := m.agentList.SetItems(bindChoicesToListItems(m.bindingAgents))
			m.agentList.Select(idx)
			return m, cmd
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
		m.hint = "Returned to skill list"
		return m, m.syncSelectionPreview()

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
