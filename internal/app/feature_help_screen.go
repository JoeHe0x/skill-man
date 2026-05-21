package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/render"
)

type helpScreenOverlay struct {
	pager viewport.Model
}

func newHelpScreenOverlay() helpScreenOverlay {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	return helpScreenOverlay{pager: vp}
}

type helpScreenFeature struct {
	m       *Model
	overlay helpScreenOverlay
}

func (f *helpScreenFeature) Name() string { return "helpScreen" }
func (f *helpScreenFeature) Active() bool {
	return f.m.state == stateHelpOverlay
}
func (f *helpScreenFeature) Init() tea.Cmd                 { return nil }
func (f *helpScreenFeature) View(width, height int) string { return "" }

func (f *helpScreenFeature) Clear() {
	f.overlay = newHelpScreenOverlay()
}

func (f *helpScreenFeature) Open() (tea.Model, tea.Cmd) {
	if f.m.state == stateHelpOverlay {
		return f.m, nil
	}
	f.m.transitionTo(stateHelpOverlay)

	w := max(40, f.m.contentWidth()-6)
	headerH, footerH := f.m.chromeHeights()
	h := max(12, f.m.height-headerH-footerH-6)
	content, err := render.Markdown(helpPreview, w)
	if err != nil {
		content = helpPreview
	}
	f.overlay.pager.Width = w
	f.overlay.pager.Height = h
	f.overlay.pager.SetContent(content)
	f.overlay.pager.GotoTop()
	f.m.setFooterContext("Command reference · Esc close · PgUp/PgDn scroll")
	return f.m, nil
}

func (f *helpScreenFeature) close() {
	if f.m.state != stateHelpOverlay {
		return
	}
	f.m.transitionTo(f.m.lastState)
}

func (f *helpScreenFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.handleKeys(msg)
		return cmd, true
	case tea.MouseMsg:
		_, cmd := f.handleMouse(msg)
		return cmd, true
	}
	return nil, false
}

func (f *helpScreenFeature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home), key.Matches(msg, keys.HelpScreen):
		f.close()
		return f.m, nil
	}
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return f.m, cmd
}

func (f *helpScreenFeature) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return f.m, cmd
}

func (f *helpScreenFeature) renderOverlay(base string) string {
	boxW := min(f.m.contentWidth(), f.m.width-2)
	innerW := max(32, boxW-4)
	headerH, footerH := f.m.chromeHeights()
	innerH := max(10, f.m.height-headerH-footerH-8)

	f.overlay.pager.Width = innerW
	f.overlay.pager.Height = innerH

	title := f.m.styles.PanelTitleFocus.Render("Command Reference")
	hint := f.m.styles.Hint.Render("Esc close · PgUp/PgDn scroll · click links in supporting terminals")
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		hint,
		f.overlay.pager.View(),
	)
	box := f.m.styles.Modal.Width(boxW).Render(body)
	return lipgloss.Place(f.m.width-2, f.m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) openHelpOverlay() (tea.Model, tea.Cmd) {
	return m.helpScreen.Open()
}
