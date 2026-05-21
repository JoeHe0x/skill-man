package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

type paletteItem struct {
	title   string
	desc    string
	search  string
	enabled bool
	run     func(paletteActionHost) (tea.Model, tea.Cmd)
}

type commandPalette struct {
	input    textinput.Model
	all      []paletteItem
	filtered []int
	cursor   int
}

type paletteFeature struct {
	host paletteActionHost
	ui   *commandPalette
}

func newCommandPalette(width int) *commandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type to filter commands…"
	ti.CharLimit = 64
	ti.Prompt = "> "
	ti.Focus()
	w := paletteInputWidth(width)
	if w > 0 {
		ti.Width = w
	}
	return &commandPalette{input: ti}
}

func paletteInputWidth(outerWidth int) int {
	return min(56, max(24, outerWidth-8))
}

func (f *paletteFeature) Name() string { return "palette" }
func (f *paletteFeature) Active() bool { return f.ui != nil }
func (f *paletteFeature) Init() tea.Cmd {
	return nil
}
func (f *paletteFeature) View(width, height int) string { return "" }

func (f *paletteFeature) canOpen() bool {
	if f.host.PromptActive() || f.ui != nil {
		return false
	}
	switch f.host.State() {
	case stateHome, stateListing, stateSearching:
		return true
	default:
		return false
	}
}

func (f *paletteFeature) Open() (tea.Model, tea.Cmd) {
	if !f.canOpen() {
		return f.host.TeaModel(), nil
	}
	f.host.TransitionTo(stateCommandPalette)
	p := newCommandPalette(f.host.ContentWidth())
	p.refresh(f.host, "")
	f.ui = p
	return f.host.TeaModel(), textinput.Blink
}

func (f *paletteFeature) Close() {
	if f.ui == nil {
		return
	}
	f.ui = nil
	if f.host.State() == stateCommandPalette {
		f.host.TransitionTo(f.host.LastState())
	}
}

func (f *paletteFeature) resizeInput() {
	if f.ui != nil {
		f.ui.input.Width = paletteInputWidth(f.host.ContentWidth())
	}
}

func (p *commandPalette) refresh(h paletteActionHost, query string) {
	p.all = buildPaletteCatalog(h)
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		p.filtered = p.filtered[:0]
		for i, item := range p.all {
			if item.enabled {
				p.filtered = append(p.filtered, i)
			}
		}
	} else {
		var candidates []string
		var indexMap []int
		for i, item := range p.all {
			if !item.enabled {
				continue
			}
			candidates = append(candidates, item.search)
			indexMap = append(indexMap, i)
		}
		matches := fuzzy.Find(query, candidates)
		p.filtered = p.filtered[:0]
		for _, match := range matches {
			p.filtered = append(p.filtered, indexMap[match.Index])
		}
	}
	if p.cursor >= len(p.filtered) {
		p.cursor = max(0, len(p.filtered)-1)
	}
}

func (m *Model) runRegistryCommand(name string) (tea.Model, tea.Cmd) {
	switch name {
	case "help":
		return m.helpScreen.Open()
	case "list":
		m.transitionTo(stateListing)
		return m, m.syncSelectionPreview()
	case "find":
		return m.startListFilter()
	case "reload":
		return m, m.beginScanAllCmd()
	case "add":
		return m.showAddPrompt()
	case "remove":
		return m.handleRemoveSelected()
	case "update":
		return m.handleUpdate()
	case "init":
		if m.activeTab == panel.TabSkills && m.activePanel().Capabilities().Init {
			return m.showInitPrompt()
		}
		return m, m.flashFooter("Init is only available on the Skills tab")
	case "agent":
		return m.handleOpenAgentFilter()
	case "inspect":
		return m.handleInspectSelected()
	case "quit":
		return m, tea.Quit
	default:
		return m, m.flashFooter("Unknown command: " + name)
	}
}

func (f *paletteFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.ui == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.handleKeys(msg)
		return cmd, true
	}
	return nil, false
}

func (f *paletteFeature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := f.ui

	switch {
	case key.Matches(msg, keys.Quit):
		return f.host.TeaModel(), tea.Quit
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		f.Close()
		return f.host.TeaModel(), nil
	case key.Matches(msg, keys.Down):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor + 1) % len(p.filtered)
		}
		return f.host.TeaModel(), nil
	case key.Matches(msg, keys.Up):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor - 1 + len(p.filtered)) % len(p.filtered)
		}
		return f.host.TeaModel(), nil
	case key.Matches(msg, keys.Enter), msg.Type == tea.KeyEnter:
		if len(p.filtered) == 0 {
			return f.host.TeaModel(), nil
		}
		idx := p.filtered[p.cursor]
		item := p.all[idx]
		f.Close()
		return item.run(f.host)
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	p.refresh(f.host, p.input.Value())
	return f.host.TeaModel(), cmd
}

func (f *paletteFeature) renderOverlay(base string) string {
	if f.ui == nil {
		return base
	}
	p := f.ui
	boxW := min(62, max(36, f.host.ContentWidth()-4))
	innerW := paletteInputWidth(boxW)

	styles := f.host.Styles()
	title := styles.PanelTitle.Render("Command Palette")
	input := p.input.View()
	help := styles.Hint.Render("↑↓ select · Enter run · Esc close")

	var rows []string
	maxRows := min(8, max(3, f.host.Height()/4))
	start := 0
	if p.cursor >= maxRows {
		start = p.cursor - maxRows + 1
	}
	end := min(len(p.filtered), start+maxRows)
	for i := start; i < end; i++ {
		idx := p.filtered[i]
		item := p.all[idx]
		line := item.title
		if item.desc != "" {
			line += " — " + item.desc
		}
		line = truncate(line, innerW)
		if i == p.cursor {
			rows = append(rows, styles.ItemSelected.Render("› "+line))
		} else {
			rows = append(rows, styles.ItemDesc.Render("  "+line))
		}
	}
	if len(p.filtered) == 0 {
		rows = append(rows, styles.EmptyPreview.Render("  No matching commands"))
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		input,
		strings.Join(rows, "\n"),
		help,
	)
	box := styles.Modal.Width(boxW).Render(body)
	return lipgloss.Place(f.host.Width()-2, f.host.Height()-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) openCommandPalette() (tea.Model, tea.Cmd) {
	return m.cmdPalette.Open()
}
