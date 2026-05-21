package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) contentWidth() int {
	return m.Core.contentWidth()
}

func (m *Model) chromeHeights() (header, footer int) {
	return lipgloss.Height(m.renderHeader()), lipgloss.Height(m.renderFooter())
}

func (m *Model) shouldStack() bool {
	return m.Core.shouldStack()
}

func (m *Model) mainAreaSize() (int, int) {
	contentWidth := m.contentWidth()
	headerH, footerH := m.chromeHeights()

	mainHeight := m.height - headerH - footerH
	if mainHeight < 6 {
		mainHeight = 6
	}

	return contentWidth, mainHeight
}

func (m *Model) paneSizes() (int, int, int, int) {
	_, mainHeight := m.mainAreaSize()
	return m.paneSizesFor(mainHeight)
}

func (m *Model) paneSizesFor(mainHeight int) (int, int, int, int) {
	contentWidth := m.contentWidth()
	if m.shouldStack() {
		topHeight := mainHeight / 2
		bottomHeight := mainHeight - topHeight
		return contentWidth, topHeight, contentWidth, bottomHeight
	}

	leftWidth := (contentWidth * 35) / 100
	if contentWidth < 120 {
		leftWidth = (contentWidth * 40) / 100
	}
	rightWidth := contentWidth - leftWidth
	return leftWidth, mainHeight, rightWidth, mainHeight
}

// panelInnerSize returns list/preview dimensions inside a bordered panel with a title row.
// Chrome: border (2) + padding (2) + title row (1) = 5 lines.
func panelInnerSize(outerWidth, outerHeight int) (int, int) {
	return max(8, outerWidth-4), max(3, outerHeight-5)
}

func clipLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}
