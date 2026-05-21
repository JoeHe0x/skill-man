package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/render"
)

func (m *Model) applyTheme(dark bool) tea.Cmd {
	if m.darkTheme == dark && m.themeReady {
		return nil
	}
	m.darkTheme = dark
	m.themeReady = true
	m.styles = theme.NewStyles(dark)

	m.listPane.applyTheme(m.styles)

	initHelpStyles(&m.help, m.styles)
	render.SetDarkTheme(dark)

	if m.install.flow != nil {
		m.install.flow.ApplyTheme(m.styles)
	}

	// Avoid spawning preview renders while scans are in flight (glamour is not concurrent-safe).
	if m.status == "loading" {
		return nil
	}
	return m.syncSelectionPreview()
}

func (m *Model) handleThemeDetected(msg theme.DetectedMsg) (tea.Model, tea.Cmd) {
	return m, m.applyTheme(msg.Dark)
}
