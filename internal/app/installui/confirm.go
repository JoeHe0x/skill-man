package installui

import (
	"strings"
)

func (m *Model) renderConfirm(outerWidth int) string {
	inner := m.panelInnerWidth(outerWidth)
	return m.renderPanel(outerWidth, "Confirm install", false, m.confirmBody(inner))
}

func (m *Model) confirmBody(innerWidth int) string {
	s := m.styles()
	var lines []string
	lines = append(lines, m.skillSummaryLines(innerWidth)...)
	lines = append(lines, "", s.HintBold.Render("Will install to:"))

	for _, t := range m.targets {
		if !t.desired {
			continue
		}
		lines = append(lines, s.ItemDesc.Render("  • "+truncate(t.skillDir, innerWidth-4)))
		if names := formatDirAgents(t.agents); names != "" {
			lines = append(lines, s.Hint.Render("      "+truncate(names, innerWidth-6)))
		}
	}

	lines = append(lines,
		"",
		s.StatusWarn.Render("Not installed yet — press Enter to run install"),
		s.Hint.Render("Esc · go back to paths"),
	)
	return strings.Join(lines, "\n")
}
