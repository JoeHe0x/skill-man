package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

// --- listing / home state key handler ---

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Home):
		m.clearError()
		if m.list.FilterState() != list.Unfiltered {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		m.transitionTo(stateHome)
		if preview := m.activePanel().StaticPreview(); preview != "" {
			m.preview.SetContent(preview)
			return m, nil
		}
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.HelpToggle):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil

	case key.Matches(msg, keys.HelpScreen):
		return m.openHelpOverlay()

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
		m.transitionTo(stateListing)
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.Find), key.Matches(msg, keys.Filter):
		return m.startListFilter()

	case key.Matches(msg, keys.Agent):
		return m.handleOpenAgentFilter()

	case key.Matches(msg, keys.Reload):
		m.status = "loading"
		m.setFooterContext(m.activePanel().ReloadHint())
		return m, m.scanAllCmd()

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

// --- command handlers ---

func (m *Model) handleInspectSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Inspect {
		m.setFooterContext("Inspect is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(actionable)
	if !ok || !selected.CanInspect() {
		return m, nil
	}
	target := selected.InspectTarget()
	switch target.Kind {
	case "skill":
		m.transitionTo(stateInspecting)
		m.tree.setRoot(target.SkillPath)
		m.setFooterContext("Inspecting skill files")
		sel := m.tree.SelectedItem()
		if sel.path != "" && !sel.isDir {
			m.status = "loading"
			return m, m.previewFileCmd(sel.path)
		}
	case "mcp":
		width := m.preview.Width
		if width == 0 {
			width = max(40, m.width/2)
		}
		pi := panel.Item{
			Kind:       panel.ItemMCP,
			MCPKey:     target.MCPKey,
			MCPMembers: target.MCPMembers,
		}
		return m, m.activePanel().SyncPreview(pi, width, &m.previewGen)
	}
	return m, nil
}

func (m *Model) handleDisableSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.selectedListItem()
	if !ok {
		m.setFooterContext("Select a skill or MCP server first")
		return m, nil
	}
	act, ok := interface{}(selected).(actionable)
	if !ok || !act.CanDisable() {
		return m, nil
	}
	target := act.DisableTarget()
	switch target.Kind {
	case "skill":
		m.status = "loading"
		action := "Disabling"
		if target.Skill.IsDisabled() {
			action = "Enabling"
		}
		m.setFooterContext(fmt.Sprintf("%s %s...", action, target.Skill.GetName()))
		return m, runCommand(&command.ToggleDisableSkill{Skill: target.Skill, Manager: m.skillManager})
	case "mcp":
		m.status = "loading"
		action := "Disabling"
		if mcpKeyDisabled(target.MCPMembers) {
			action = "Enabling"
		}
		key := mcpKeyFromListItem(selected)
		m.setFooterContext(fmt.Sprintf("%s MCP `%s`...", action, key))
		return m, runCommand(&command.ToggleDisableMCPKey{Members: target.MCPMembers, Manager: m.mcpManager})
	}
	return m, nil
}

func (m *Model) handleRemoveSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.selectedListItem()
	if !ok {
		m.setFooterContext("Select an item first")
		return m, nil
	}
	act, ok := interface{}(selected).(actionable)
	if !ok || !act.CanRemove() {
		return m, nil
	}
	target := act.RemoveTarget()
	switch target.Kind {
	case "skill":
		m.pending = &pendingAction{name: "remove", skillName: target.Skill.GetName(), skill: target.Skill}
	case "mcp":
		m.pending = &pendingAction{
			name:       "remove",
			mcpName:    target.MCPName,
			mcpMembers: target.MCPMembers,
		}
	}
	m.transitionTo(stateConfirming)
	return m, nil
}

func (m *Model) handleUpdate() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Update {
		m.setFooterContext("Update is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(actionable)
	if ok && selected.CanUpdate() {
		target := selected.UpdateTarget()
		m.status = "loading"
		m.setFooterContext(fmt.Sprintf("Updating %s...", target.Skill.GetName()))
		return m, runCommand(&command.UpdateSkill{Skill: target.Skill})
	}
	m.status = "loading"
	m.setFooterContext("Updating all managed local skills...")
	return m, runCommand(&command.UpdateAllSkills{Skills: m.panels.Skills()})
}

// --- prompt launchers ---

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Find {
		m.setFooterContext("Find is not available for this tab")
		return m, nil
	}
	return m, m.showPrompt("Find", "search query...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		text = strings.TrimSpace(text)
		m.transitionTo(stateSearching)
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
		return runCommand(&command.AddSkill{Source: source, CWD: m.cwd, Agents: m.activeAgents()})
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
		return runCommand(&command.InitSkill{Root: m.cwd, Name: name})
	})
}

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
