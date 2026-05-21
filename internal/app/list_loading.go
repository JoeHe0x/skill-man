package app

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func (m *Model) shouldShowListLoading() bool {
	if m.status != "loading" {
		return false
	}
	switch m.state {
	case stateInspecting, stateBindingAgent, stateFilteringAgent,
		stateConfirming, stateInstalling, stateCommandPalette, stateHelpOverlay:
		return false
	default:
		return true
	}
}

func (m *Model) scanLoadingLabel() string {
	switch m.activeTab {
	case panel.TabMCP:
		return "Scanning MCP configs…"
	default:
		return "Scanning skills…"
	}
}

func (m *Model) renderListLoading(width, height int) string {
	spin := m.spinner.View()
	label := m.styles.Hint.Render(m.scanLoadingLabel())
	block := lipgloss.JoinVertical(lipgloss.Center, spin, label)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}

func (m *Model) renderStartupSplash() string {
	spin := m.spinner.View()
	label := m.styles.Hint.Render("Loading skills…")
	return lipgloss.JoinVertical(lipgloss.Center, spin, label)
}
