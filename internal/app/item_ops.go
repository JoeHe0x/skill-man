package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func (m *Model) selectedPanelItem() (panel.Item, bool) {
	return m.selectedListItem()
}

func (m *Model) inspectItem(item panel.Item) (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Inspect || !item.CanInspect() {
		if !m.activePanel().Capabilities().Inspect {
			m.setFooterContext("Inspect is not available for this tab")
		}
		return m, nil
	}
	eff, ok := item.InspectEffect()
	if !ok {
		return m, nil
	}
	if eff.SkillPath != "" {
		m.transitionTo(stateInspecting)
		m.tree.setRoot(eff.SkillPath)
		m.setFooterContext("Inspecting skill files")
		sel := m.tree.SelectedItem()
		if sel.path != "" && !sel.isDir {
			return m, m.previewFileCmd(sel.path)
		}
		return m, nil
	}
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	pi := panel.Item{
		Kind:       panel.ItemMCP,
		MCPKey:     eff.MCPKey,
		MCPMembers: eff.MCPMembers,
	}
	return m, m.activePanel().SyncPreview(pi, width, &m.previewGen)
}

func (m *Model) disableItem(item panel.Item) (tea.Model, tea.Cmd) {
	if m.status == "loading" {
		return m, nil
	}
	if !item.CanDisable() {
		m.setFooterContext("Select a skill or MCP server first")
		return m, nil
	}
	eff, ok := item.DisableEffect()
	if !ok {
		return m, nil
	}
	if eff.Skill != nil {
		m.status = "loading"
		action := "Disabling"
		if eff.Skill.IsDisabled() {
			action = "Enabling"
		}
		m.setFooterContext(fmt.Sprintf("%s %s...", action, eff.Skill.GetName()))
		return m, runCommand(&command.ToggleDisableSkill{Skill: eff.Skill, Manager: m.skillManager})
	}
	if len(eff.MCPMembers) > 0 {
		m.status = "loading"
		action := "Disabling"
		if mcpKeyDisabled(eff.MCPMembers) {
			action = "Enabling"
		}
		key := item.MCPConfigKey()
		m.setFooterContext(fmt.Sprintf("%s MCP `%s`...", action, key))
		return m, runCommand(&command.ToggleDisableMCPKey{Members: eff.MCPMembers, Manager: m.mcpManager})
	}
	return m, nil
}

func (m *Model) removeItem(item panel.Item) (tea.Model, tea.Cmd) {
	if !item.CanRemove() {
		m.setFooterContext("Select an item first")
		return m, nil
	}
	eff, ok := item.RemoveEffect()
	if !ok {
		return m, nil
	}
	return m.confirm.requestRemove(eff)
}

func (m *Model) updateItem(item panel.Item) (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Update {
		m.setFooterContext("Update is not available for this tab")
		return m, nil
	}
	if eff, ok := item.UpdateEffect(); ok {
		m.status = "loading"
		m.setFooterContext(fmt.Sprintf("Updating %s...", eff.Skill.GetName()))
		return m, runCommand(&command.UpdateSkill{Skill: eff.Skill})
	}
	m.status = "loading"
	m.setFooterContext("Updating all managed local skills...")
	return m, runCommand(&command.UpdateAllSkills{Skills: m.panels.Skills()})
}
