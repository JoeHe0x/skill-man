package installui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// panelInnerWidth is the content width inside Modal border + padding.
func (m Model) panelInnerWidth(outerWidth int) int {
	frameX, _ := m.styles().Modal.GetFrameSize()
	return max(16, outerWidth-frameX)
}

func (m Model) panelStyle(outerWidth int, accent bool) lipgloss.Style {
	st := m.styles().Modal.Width(outerWidth)
	if accent {
		return st.BorderForeground(lipgloss.Color("69"))
	}
	return st
}

func (m Model) renderPanel(outerWidth int, title string, accent bool, blocks ...string) string {
	s := m.styles()
	var lines []string
	if title != "" {
		lines = append(lines, s.PanelTitle.Render(title))
	}
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, block)
	}
	return m.panelStyle(outerWidth, accent).Render(strings.Join(lines, "\n"))
}

func (m Model) searchFieldStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Width(width)
}

func (m Model) renderSearchField(innerWidth int) string {
	contentW := max(8, innerWidth-4)
	var line string
	if m.searching {
		spin := m.spinner.View()
		spinW := lipgloss.Width(spin)
		m.searchInput.Width = max(4, contentW-spinW-1)
		line = lipgloss.JoinHorizontal(lipgloss.Center, m.searchInput.View(), spin)
	} else {
		m.searchInput.Width = contentW
		line = m.searchInput.View()
	}
	return m.searchFieldStyle(innerWidth).Render(line)
}

func (m Model) skillSummaryLines(innerWidth int) []string {
	if m.selected.Name == "" {
		return nil
	}
	s := m.styles()
	lines := []string{s.ItemSelected.Render(truncate(m.selected.Name, innerWidth))}
	if m.selected.Source != "" {
		lines = append(lines, s.Hint.Render("from "+truncate(m.selected.Source, innerWidth-5)))
	}
	return lines
}

func (m Model) renderBrowse(outerWidth, listHeight int) string {
	inner := m.panelInnerWidth(outerWidth)
	s := m.styles()
	var blocks []string

	blocks = append(blocks, m.renderSearchField(inner))

	if err := m.hostErrMsg(); err != "" && len(m.results) == 0 && !m.searching {
		blocks = append(blocks, s.StatusError.Render(truncate(err, inner)))
	}

	if len(m.results) > 0 || m.searching {
		m.resultList.SetSize(inner, listHeight)
		blocks = append(blocks, m.resultList.View())
	}

	blocks = append(blocks, s.Hint.Render("Enter · search   / · focus search   Esc · close"))
	return m.renderPanel(outerWidth, "Install skill", false, blocks...)
}

func (m Model) renderPaths(outerWidth, listHeight int) string {
	inner := m.panelInnerWidth(outerWidth)
	s := m.styles()
	m.pathsList.SetSize(inner, listHeight)

	blocks := []string{
		strings.Join(m.skillSummaryLines(inner), "\n"),
		m.pathsList.View(),
		s.Hint.Render("Space · toggle path   Enter · review   Esc · back"),
	}
	return m.renderPanel(outerWidth, "Install paths", false, blocks...)
}

func (m Model) renderInstalling(outerWidth int) string {
	inner := m.panelInnerWidth(outerWidth)
	s := m.styles()
	m.progress.Width = max(20, inner-2)

	blocks := []string{
		strings.Join(m.skillSummaryLines(inner), "\n"),
		m.progress.View(),
		s.Hint.Render("Running skills CLI — please wait   Esc twice · cancel"),
	}
	return m.renderPanel(outerWidth, "Installing…", true, blocks...)
}
