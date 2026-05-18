package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func TestLayoutChromeFitsTerminal(t *testing.T) {
	sizes := [][2]int{{120, 40}, {80, 24}, {100, 30}, {90, 20}}
	for _, sz := range sizes {
		m := New("/mnt/c/Code/skill-man", "/home/joe")
		m.width = sz[0]
		m.height = sz[1]

		header := m.renderHeader()
		footer := m.renderFooter()
		_, wantMainH := m.mainAreaSize()
		main := m.renderMainAreaSized(wantMainH)

		headerH := lipgloss.Height(header)
		footerH := lipgloss.Height(footer)
		mainLines := lipgloss.Height(main)

		if headerH != lineCount(header) {
			t.Errorf("%dx%d header lipgloss.Height=%d lines=%d", sz[0], sz[1], headerH, lineCount(header))
		}
		if wantMainH != m.height-headerH-footerH && wantMainH != m.height-lipgloss.Height(m.renderHeader())-lipgloss.Height(m.renderFooter()) {
			// wantMainH clamped at 6
		}
		if headerH+wantMainH+footerH > m.height {
			t.Errorf("%dx%d chrome exceeds terminal: header=%d main=%d footer=%d sum=%d term=%d",
				sz[0], sz[1], headerH, wantMainH, footerH, headerH+wantMainH+footerH, m.height)
		}
		if mainLines > wantMainH {
			t.Errorf("%dx%d main rendered taller than budget: got=%d want=%d", sz[0], sz[1], mainLines, wantMainH)
		}
	}
}

func TestViewTotalHeight(t *testing.T) {
	m := New("/mnt/c/Code/skill-man", "/home/joe")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := updated.(*Model)
	view := m2.View()
	got := lineCount(view)
	if got > m2.height {
		t.Fatalf("view lines %d exceed terminal height %d", got, m2.height)
	}
}
