package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func (m *Model) renderHeader() string {
	if m.width >= 80 && m.height >= 24 {
		return m.renderFullHeader()
	}
	return m.renderCompactHeader()
}

func (m *Model) renderFullHeader() string {
	cwd := m.cwd
	if len(cwd) > 48 {
		cwd = "…" + cwd[len(cwd)-47:]
	}

	topLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.appTitle.Render(" skill-man "),
		m.styles.appVersion.Render("v0.1"),
		m.styles.appPath.Render(cwd),
	)

	statsLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.statusBarDim.Render("scope: project"),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("agents: %s", m.agentDisplay())),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("skills: %d", len(m.panels.Skills()))),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("mcp: %d", len(m.panels.MCPServers()))),
		m.styles.statusBarSep.Render(" │ "),
		m.statusView(),
	)

	inner := lipgloss.JoinVertical(lipgloss.Left, topLine, "", statsLine)
	banner := m.styles.headerBanner.Width(max(20, m.width-4)).Render(inner)

	return lipgloss.JoinVertical(lipgloss.Left, banner, m.renderExtensionTabs())
}

func (m *Model) renderCompactHeader() string {
	line1 := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.appTitleCompact.Render("skill-man"),
		m.styles.appVersion.Render("v0.1"),
		m.styles.statusBarDim.Render(fmt.Sprintf("agents: %s", m.agentDisplay())),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("skills: %d", len(m.panels.Skills()))),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("mcp: %d", len(m.panels.MCPServers()))),
		m.styles.statusBarSep.Render(" │ "),
		m.statusView(),
	)

	return lipgloss.JoinVertical(lipgloss.Left, line1, m.renderExtensionTabs())
}

func (m *Model) renderExtensionTabs() string {
	skillTab := m.styles.tabInactive.Render("Skills")
	if m.activeTab == panel.TabSkills {
		skillTab = m.styles.tabActive.Render("Skills")
	}
	mcpTab := m.styles.tabInactive.Render("MCP")
	if m.activeTab == panel.TabMCP {
		mcpTab = m.styles.tabActive.Render("MCP")
	}
	tabs := lipgloss.JoinHorizontal(
		lipgloss.Left,
		skillTab,
		m.styles.tabSep.Render(" │ "),
		mcpTab,
	)
	return m.styles.tabBar.Render(tabs)
}
