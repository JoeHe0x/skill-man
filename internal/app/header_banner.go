package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) headerOverviewText() string {
	cwd := m.cwd
	if len(cwd) > 48 {
		cwd = "…" + cwd[len(cwd)-47:]
	}
	return strings.TrimSpace(fmt.Sprintf(`skill-man

Skills  %d    MCP  %d
Agents  %s
Status  %s

scope   project
%s`,
		len(m.panels.Skills()),
		len(m.panels.MCPServers()),
		m.agentDisplay(),
		m.statusLabel(),
		cwd,
	))
}

func (m *Model) statusLabel() string {
	if m.errMsg != "" {
		return "error"
	}
	return m.status
}

func joinHeaderColumns(left, right string, sep lipgloss.Style) string {
	leftLines := strings.Split(strings.TrimSuffix(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimSuffix(right, "\n"), "\n")
	height := max(len(leftLines), len(rightLines))

	logoWidth := 0
	for _, line := range leftLines {
		if w := lipgloss.Width(line); w > logoWidth {
			logoWidth = w
		}
	}

	sepCell := sep.Render("│")

	var rows []string
	for i := 0; i < height; i++ {
		l := padLine(leftLines, i, logoWidth)
		r := padLine(rightLines, i, 0)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, l, " ", sepCell, " ", r))
	}
	return strings.Join(rows, "\n")
}

func padLine(lines []string, idx, width int) string {
	if idx < len(lines) {
		line := lines[idx]
		if width > 0 {
			pad := width - lipgloss.Width(line)
			if pad > 0 {
				return line + strings.Repeat(" ", pad)
			}
		}
		return line
	}
	if width > 0 {
		return strings.Repeat(" ", width)
	}
	return ""
}
