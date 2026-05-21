package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/command"
)

func (m *Model) beginRemoveConfirm() {
	m.setFooterContext("y confirm · n/Esc cancel")
}

func (m *Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		return m.executeRemoveConfirm()
	case key.Matches(msg, keys.Cancel):
		return m.cancelRemoveConfirm()
	default:
		return m, nil
	}
}

func (m *Model) cancelRemoveConfirm() (tea.Model, tea.Cmd) {
	m.pending = nil
	m.transitionTo(stateListing)
	m.setFooterContext("Cancelled")
	return m, nil
}

func (m *Model) executeRemoveConfirm() (tea.Model, tea.Cmd) {
	if m.pending == nil || m.pending.name != "remove" {
		m.pending = nil
		m.transitionTo(stateListing)
		return m, nil
	}
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

func (m *Model) renderRemoveConfirmArea() string {
	leftWidth, mainHeight, _, _ := m.paneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, m.renderRemoveConfirmDialog())
}

func (m *Model) renderRemoveConfirmDialog() string {
	if m.pending == nil {
		return ""
	}
	leftWidth, _, _, _ := m.paneSizes()
	dialogWidth := min(max(36, leftWidth-4), 52)
	if dialogWidth > leftWidth-2 {
		dialogWidth = max(24, leftWidth-2)
	}

	target := m.pending.skillName
	if m.pending.mcpName != "" {
		target = "MCP " + m.pending.mcpName
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.PanelTitle.Render("Remove "+truncate(target, dialogWidth-8)+"?"),
		m.styles.Hint.Render("[y/N]"),
	)

	return m.styles.ModalDanger.
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		Render(body)
}
