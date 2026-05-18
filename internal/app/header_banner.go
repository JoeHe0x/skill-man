package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func (m *Model) renderHeader() string {
	if m.width >= 80 && m.height >= 24 {
		return m.renderFullHeader()
	}
	return m.renderCompactHeader()
}

func (m *Model) renderFullHeader() string {
	w := m.contentWidth()
	innerW := bannerInnerWidth(w)

	titleBlock := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.appTitle.Render(" skill-man "),
		m.styles.appVersion.Render("v0.1"),
	)

	cwdMax := innerW - lipgloss.Width(titleBlock) - 1
	if cwdMax < 8 {
		cwdMax = 8
	}
	cwdStyled := m.styles.appPath.Render(truncateRunes(m.cwd, cwdMax))

	statsLeft := m.styles.statusBarDim.Render(fmt.Sprintf(
		"project · agents: %s · %d skills · %d mcp",
		m.agentDisplay(),
		len(m.panels.Skills()),
		len(m.panels.MCPServers()),
	))

	inner := lipgloss.JoinVertical(lipgloss.Left,
		joinHeaderRow(innerW, titleBlock, cwdStyled),
		joinHeaderRow(innerW, statsLeft, m.statusView()),
	)
	banner := m.styles.headerBanner.Width(w).Render(inner)

	return lipgloss.JoinVertical(lipgloss.Left, banner, m.renderExtensionTabs())
}

func (m *Model) renderCompactHeader() string {
	w := m.contentWidth()

	left := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.appTitleCompact.Render("skill-man"),
		m.styles.appVersion.Render("v0.1"),
		m.styles.statusBarDim.Render(fmt.Sprintf(
			"· agents: %s · %d skills · %d mcp",
			m.agentDisplay(),
			len(m.panels.Skills()),
			len(m.panels.MCPServers()),
		)),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		joinHeaderRow(w, left, m.statusView()),
		m.renderExtensionTabs(),
	)
}

// bannerInnerWidth is the usable text width inside headerBanner (border + horizontal padding).
func bannerInnerWidth(outer int) int {
	return max(20, outer-4)
}

func joinHeaderRow(width int, left, right string) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	if leftW+rightW+1 > width {
		right = truncateStyled(right, max(1, width-leftW-1))
		rightW = lipgloss.Width(right)
	}
	gap := width - leftW - rightW
	if gap < 1 {
		return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", right)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, left, strings.Repeat(" ", gap), right)
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return "…" + string(runes[len(runes)-(max-1):])
}

func truncateStyled(s string, max int) string {
	if max <= 0 {
		return ""
	}
	return ansi.Truncate(s, max, "…")
}

func (m *Model) renderExtensionTabs() string {
	skillTab := m.renderTabItem("Skills", m.activeTab == panel.TabSkills)
	mcpTab := m.renderTabItem("MCP", m.activeTab == panel.TabMCP)
	tabs := lipgloss.JoinHorizontal(
		lipgloss.Left,
		skillTab,
		"  ",
		mcpTab,
	)
	return m.styles.tabBar.Width(m.contentWidth()).Render(tabs)
}

func (m *Model) renderTabItem(name string, active bool) string {
	var label string
	if active {
		label = m.styles.tabActive.Render(name)
	} else {
		label = m.styles.tabInactive.Render(name)
	}
	w := lipgloss.Width(label)
	if active {
		underline := m.styles.tabUnderline.Render(strings.Repeat("─", w))
		return lipgloss.JoinVertical(lipgloss.Left, label, underline)
	}
	return lipgloss.JoinVertical(lipgloss.Left, label, strings.Repeat(" ", w))
}
