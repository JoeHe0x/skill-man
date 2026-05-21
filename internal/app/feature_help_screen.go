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
	host    helpHost
	overlay helpScreenOverlay
}

func (f *helpScreenFeature) Name() string { return "helpScreen" }
func (f *helpScreenFeature) Active() bool {
	return f.host.IsHelpOverlay()
}
func (f *helpScreenFeature) Init() tea.Cmd                 { return nil }
func (f *helpScreenFeature) View(width, height int) string { return "" }

func (f *helpScreenFeature) Clear() {
	f.overlay = newHelpScreenOverlay()
}

func (f *helpScreenFeature) Open() (tea.Model, tea.Cmd) {
	if f.host.IsHelpOverlay() {
		return f.host.TeaModel(), nil
	}
	f.host.TransitionTo(stateHelpOverlay)

	w := max(40, f.host.ContentWidth()-6)
	headerH, footerH := f.host.ChromeHeights()
	h := max(12, f.host.Height()-headerH-footerH-6)
	content, err := render.Markdown(helpPreview, w)
	if err != nil {
		content = helpPreview
	}
	f.overlay.pager.Width = w
	f.overlay.pager.Height = h
	f.overlay.pager.SetContent(content)
	f.overlay.pager.GotoTop()
	f.host.SetFooterContext("Command reference · Esc close · PgUp/PgDn scroll")
	return f.host.TeaModel(), nil
}

func (f *helpScreenFeature) close() {
	if !f.host.IsHelpOverlay() {
		return
	}
	f.host.TransitionTo(f.host.LastState())
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
	m := f.host.TeaModel()
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home), key.Matches(msg, keys.HelpScreen):
		f.close()
		return m, nil
	}
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return m, cmd
}

func (f *helpScreenFeature) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return f.host.TeaModel(), cmd
}

func (f *helpScreenFeature) renderOverlay(base string) string {
	boxW := min(f.host.ContentWidth(), f.host.Width()-2)
	innerW := max(32, boxW-4)
	headerH, footerH := f.host.ChromeHeights()
	innerH := max(10, f.host.Height()-headerH-footerH-8)

	f.overlay.pager.Width = innerW
	f.overlay.pager.Height = innerH

	styles := f.host.Styles()
	title := styles.PanelTitleFocus.Render("Command Reference")
	hint := styles.Hint.Render("Esc close · PgUp/PgDn scroll · click links in supporting terminals")
	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		hint,
		f.overlay.pager.View(),
	)
	box := styles.Modal.Width(boxW).Render(body)
	return lipgloss.Place(f.host.Width()-2, f.host.Height()-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) openHelpOverlay() (tea.Model, tea.Cmd) {
	return m.helpScreen.Open()
}
