package palette

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/strutil"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

type commandPalette struct {
	input    textinput.Model
	all      []item
	filtered []int
	cursor   int
}

// Feature owns the Ctrl+P command palette.
type Feature struct {
	host ActionHost
	ui   *commandPalette
}

// New returns a palette feature wired to host.
func New(host ActionHost) *Feature {
	return &Feature{host: host}
}

func newCommandPalette(width int) *commandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type to filter commands…"
	ti.CharLimit = 64
	ti.Prompt = "> "
	ti.Focus()
	w := inputWidth(width)
	if w > 0 {
		ti.Width = w
	}
	return &commandPalette{input: ti}
}

func inputWidth(outerWidth int) int {
	return min(56, max(24, outerWidth-8))
}

func (f *Feature) Name() string { return "palette" }
func (f *Feature) Active() bool { return f.ui != nil }
func (f *Feature) Init() tea.Cmd {
	return nil
}
func (f *Feature) View(width, height int) string { return "" }

func (f *Feature) canOpen() bool {
	if f.host.PromptActive() || f.ui != nil {
		return false
	}
	switch f.host.State() {
	case session.Home, session.Listing, session.Searching:
		return true
	default:
		return false
	}
}

func (f *Feature) Open() (tea.Model, tea.Cmd) {
	if !f.canOpen() {
		return f.host.TeaModel(), nil
	}
	f.host.TransitionTo(session.CommandPalette)
	p := newCommandPalette(f.host.ContentWidth())
	p.refresh(f.host, "")
	f.ui = p
	return f.host.TeaModel(), textinput.Blink
}

func (f *Feature) Close() {
	if f.ui == nil {
		return
	}
	f.ui = nil
	if f.host.State() == session.CommandPalette {
		f.host.TransitionTo(f.host.LastState())
	}
}

func (f *Feature) ResizeInput() {
	if f.ui != nil {
		f.ui.input.Width = inputWidth(f.host.ContentWidth())
	}
}

func (p *commandPalette) refresh(h ActionHost, query string) {
	p.all = buildCatalog(h)
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

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
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

func (f *Feature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := f.ui
	keys := uikeys.Default

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

// MatchCount returns the number of filtered palette entries (for tests).
func (f *Feature) MatchCount() int {
	if f.ui == nil {
		return 0
	}
	return len(f.ui.filtered)
}

// TopMatchTitle returns the title of the top filtered entry (for tests).
func (f *Feature) TopMatchTitle() string {
	if f.ui == nil || len(f.ui.filtered) == 0 {
		return ""
	}
	return f.ui.all[f.ui.filtered[0]].title
}

func (f *Feature) RenderOverlay(base string) string {
	if f.ui == nil {
		return base
	}
	p := f.ui
	boxW := min(62, max(36, f.host.ContentWidth()-4))
	innerW := inputWidth(boxW)

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
		line = strutil.Truncate(line, innerW)
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
