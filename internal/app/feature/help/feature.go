package help

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
	"github.com/JoeHe0x/skill-man/internal/render"
)

type overlay struct {
	pager viewport.Model
}

func newOverlay() overlay {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = true
	return overlay{pager: vp}
}

// Feature owns the F1 command reference overlay.
type Feature struct {
	host    Host
	overlay overlay
}

// New returns a help feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host, overlay: newOverlay()}
}

func (f *Feature) Name() string { return "helpScreen" }
func (f *Feature) Active() bool {
	return f.host.IsHelpOverlay()
}
func (f *Feature) Init() tea.Cmd                 { return nil }
func (f *Feature) View(width, height int) string { return "" }

func (f *Feature) Clear() {
	f.overlay = newOverlay()
}

func (f *Feature) Open() (tea.Model, tea.Cmd) {
	if f.host.IsHelpOverlay() {
		return f.host.TeaModel(), nil
	}
	f.host.TransitionTo(session.HelpOverlay)

	w := max(40, f.host.ContentWidth()-6)
	headerH, footerH := f.host.ChromeHeights()
	h := max(12, f.host.Height()-headerH-footerH-6)
	content, err := render.Markdown(PreviewMarkdown, w)
	if err != nil {
		content = PreviewMarkdown
	}
	f.overlay.pager.Width = w
	f.overlay.pager.Height = h
	f.overlay.pager.SetContent(content)
	f.overlay.pager.GotoTop()
	f.host.SetFooterContext("Command reference · Esc close · PgUp/PgDn scroll")
	return f.host.TeaModel(), nil
}

func (f *Feature) close() {
	if !f.host.IsHelpOverlay() {
		return
	}
	f.host.TransitionTo(f.host.LastState())
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
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

func (f *Feature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := f.host.TeaModel()
	keys := uikeys.Default
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home), key.Matches(msg, keys.HelpScreen):
		f.close()
		return m, nil
	}
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return m, cmd
}

func (f *Feature) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	f.overlay.pager, cmd = f.overlay.pager.Update(msg)
	return f.host.TeaModel(), cmd
}

func (f *Feature) RenderOverlay(base string) string {
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
