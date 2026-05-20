package app

import (
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/JoeHe0x/skill-man/internal/render"
)

type themeDetectedMsg struct {
	dark bool
}

func detectTerminalThemeCmd() tea.Cmd {
	return func() tea.Msg {
		configureTerminalColorProfile()
		return themeDetectedMsg{dark: lipgloss.HasDarkBackground()}
	}
}

func configureTerminalColorProfile() {
	p := colorprofile.Detect(os.Stdout, os.Environ())
	lipgloss.SetColorProfile(colorProfileToTermenv(p))
}

func colorProfileToTermenv(p colorprofile.Profile) termenv.Profile {
	switch p {
	case colorprofile.TrueColor:
		return termenv.TrueColor
	case colorprofile.ANSI256:
		return termenv.ANSI256
	case colorprofile.ANSI:
		return termenv.ANSI
	default:
		return termenv.Ascii
	}
}

func (m *Model) applyTheme(dark bool) tea.Cmd {
	if m.darkTheme == dark && m.themeReady {
		return nil
	}
	m.darkTheme = dark
	m.themeReady = true
	m.styles = newStyles(dark)

	if m.listDelegate != nil {
		m.listDelegate.styles = m.styles
	}
	if m.agentListDelegate != nil {
		m.agentListDelegate.styles = m.styles
	}
	m.tree.setStyles(m.styles)
	m.list.SetDelegate(m.listDelegate)
	m.agentList.SetDelegate(m.agentListDelegate)

	initHelpStyles(&m.help, m.styles)
	render.SetDarkTheme(dark)

	if m.install.flow != nil {
		m.install.flow.progress = progress.New(progress.WithDefaultGradient(), progress.WithWidth(m.install.flow.progress.Width))
	}

	return m.syncSelectionPreview()
}

func (m *Model) handleThemeDetected(msg themeDetectedMsg) (tea.Model, tea.Cmd) {
	return m, m.applyTheme(msg.dark)
}
