package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
)

func (m *Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		if m.pending != nil && m.pending.name == "remove" {
			if len(m.pending.mcpMembers) > 0 {
				members := m.pending.mcpMembers
				name := m.pending.mcpName
				m.pending = nil
				m.transitionTo(stateListing)
				m.status = "loading"
				m.setFooterContext(fmt.Sprintf("Removing MCP `%s`...", name))
				return m, runCommand(&command.RemoveMCPKey{Members: members, Manager: m.mcpManager})
			}
			skill := m.pending.skill
			m.pending = nil
			m.transitionTo(stateListing)
			m.status = "loading"
			m.setFooterContext(fmt.Sprintf("Removing %s...", skill.GetName()))
			return m, runCommand(&command.RemoveSkill{Skill: skill, Manager: m.skillManager, ProjectRoot: m.cwd, Home: m.home})
		}
		m.pending = nil
		m.transitionTo(stateListing)
		return m, nil
	case key.Matches(msg, keys.Cancel):
		m.pending = nil
		m.transitionTo(stateListing)
		m.setFooterContext("Cancelled")
		return m, nil
	default:
		return m, nil
	}
}
