package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/render"
)

type helpOverlay struct {
	pager viewport.Model
}

func newHelpOverlay() helpOverlay {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	return helpOverlay{pager: vp}
}

func (m *Model) openHelpOverlay() (tea.Model, tea.Cmd) {
	if m.state == stateHelpOverlay {
		return m, nil
	}
	m.transitionTo(stateHelpOverlay)

	w := max(40, m.contentWidth()-6)
	headerH, footerH := m.chromeHeights()
	h := max(12, m.height-headerH-footerH-6)
	content, err := render.Markdown(helpPreview, w)
	if err != nil {
		content = helpPreview
	}
	m.helpOverlay.pager.Width = w
	m.helpOverlay.pager.Height = h
	m.helpOverlay.pager.SetContent(content)
	m.helpOverlay.pager.GotoTop()
	m.setFooterContext("Command reference · Esc close · PgUp/PgDn scroll")
	return m, nil
}

func (m *Model) closeHelpOverlay() {
	if m.state != stateHelpOverlay {
		return
	}
	m.transitionTo(m.lastState)
}

func (m *Model) handleHelpOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home), key.Matches(msg, keys.HelpScreen):
		m.closeHelpOverlay()
		return m, nil
	}
	var cmd tea.Cmd
	m.helpOverlay.pager, cmd = m.helpOverlay.pager.Update(msg)
	return m, cmd
}

func (m *Model) handleHelpOverlayMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.helpOverlay.pager, cmd = m.helpOverlay.pager.Update(msg)
	return m, cmd
}

func (m *Model) renderHelpOverlay(base string) string {
	boxW := min(m.contentWidth(), m.width-2)
	innerW := max(32, boxW-4)
	headerH, footerH := m.chromeHeights()
	innerH := max(10, m.height-headerH-footerH-8)

	m.helpOverlay.pager.Width = innerW
	m.helpOverlay.pager.Height = innerH

	title := m.styles.PanelTitleFocus.Render("Command Reference")
	hint := m.styles.Hint.Render("Esc close · PgUp/PgDn scroll · click links in supporting terminals")
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		hint,
		m.helpOverlay.pager.View(),
	)
	box := m.styles.Modal.Width(boxW).Render(body)
	return lipgloss.Place(m.width-2, m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}
