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
	run     func(*Model) (tea.Model, tea.Cmd)
}

type commandPalette struct {
	input    textinput.Model
	all      []paletteItem
	filtered []int
	cursor   int
}

type paletteFeature struct {
	m  *Model
	ui *commandPalette
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
	if f.m.prompt.Active() || f.ui != nil {
		return false
	}
	switch f.m.state {
	case stateHome, stateListing, stateSearching:
		return true
	default:
		return false
	}
}

func (f *paletteFeature) Open() (tea.Model, tea.Cmd) {
	if !f.canOpen() {
		return f.m, nil
	}
	f.m.transitionTo(stateCommandPalette)
	p := newCommandPalette(f.m.contentWidth())
	p.refresh(f.m, "")
	f.ui = p
	return f.m, textinput.Blink
}

func (f *paletteFeature) Close() {
	if f.ui == nil {
		return
	}
	f.ui = nil
	if f.m.state == stateCommandPalette {
		f.m.transitionTo(f.m.lastState)
	}
}

func (f *paletteFeature) resizeInput() {
	if f.ui != nil {
		f.ui.input.Width = paletteInputWidth(f.m.contentWidth())
	}
}

func (p *commandPalette) refresh(m *Model, query string) {
	p.all = m.paletteCatalog()
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

func (m *Model) paletteCatalog() []paletteItem {
	caps := m.activePanel().Capabilities()
	var items []paletteItem

	add := func(title, desc, search string, enabled bool, run func(*Model) (tea.Model, tea.Cmd)) {
		items = append(items, paletteItem{
			title:   title,
			desc:    desc,
			search:  strings.ToLower(title + " " + desc + " " + search),
			enabled: enabled,
			run:     run,
		})
	}

	add("Reload", "Rescan skills and MCP on disk", "reload rescan refresh ctrl+r", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m, m.beginScanAllCmd()
	})
	add("Filter list", "Fuzzy filter in the skills/MCP list", "find filter search ctrl+f /", caps.Find, func(m *Model) (tea.Model, tea.Cmd) {
		return m.startListFilter()
	})
	add("Agent filter", "Choose which agent context to show", "agent filter ctrl+a", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m.handleOpenAgentFilter()
	})
	add("Command reference", "Show help and slash commands", "help commands f1 ?", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m.helpScreen.Open()
	})
	add("Focus list", "Return to the extension list", "list home ctrl+l", true, func(m *Model) (tea.Model, tea.Cmd) {
		m.transitionTo(stateListing)
		return m, m.syncSelectionPreview()
	})
	add("Tab: Skills", "Switch to the Skills panel", "tab skills", m.activeTab != panel.TabSkills, func(m *Model) (tea.Model, tea.Cmd) {
		return m, m.setActiveTab(panel.TabSkills)
	})
	add("Tab: MCP", "Switch to the MCP panel", "tab mcp", m.activeTab != panel.TabMCP, func(m *Model) (tea.Model, tea.Cmd) {
		return m, m.setActiveTab(panel.TabMCP)
	})

	if caps.SearchInstall && m.activeTab == panel.TabSkills {
		add("Search & install", "Search skills.sh and install", "install add registry ctrl+d", true, func(m *Model) (tea.Model, tea.Cmd) {
			return m.startInstallFlow()
		})
	}
	if caps.Init && m.activeTab == panel.TabSkills {
		add("New skill template", "Create SKILL.md scaffold", "init new ctrl+n", true, func(m *Model) (tea.Model, tea.Cmd) {
			return m.showInitPrompt()
		})
	}
	if caps.Update && m.activeTab == panel.TabSkills {
		add("Update skills", "Update selected or all local skills", "update ctrl+u", true, func(m *Model) (tea.Model, tea.Cmd) {
			return m.handleUpdate()
		})
	}

	if sel, ok := m.list.SelectedItem().(panel.Item); ok {
		if caps.Inspect && sel.CanInspect() {
			add("Inspect", "Browse or preview "+sel.Title, "inspect enter tree", true, func(m *Model) (tea.Model, tea.Cmd) {
				return m.handleInspectSelected()
			})
		}
		if caps.Bind && sel.CanBind() {
			add("Bind agents", "Manage agent bindings for "+sel.Title, "bind b", true, func(m *Model) (tea.Model, tea.Cmd) {
				return m.handleBindSelected()
			})
		}
		if caps.Disable && sel.CanDisable() {
			add("Toggle disable", "Enable or disable "+sel.Title, "toggle disable x", true, func(m *Model) (tea.Model, tea.Cmd) {
				return m.handleDisableSelected()
			})
		}
		if caps.Remove && sel.CanRemove() {
			add("Remove", "Delete "+sel.Title+" (confirmed)", "remove delete del", true, func(m *Model) (tea.Model, tea.Cmd) {
				return m.handleRemoveSelected()
			})
		}
	}

	for _, spec := range m.registry.Specs() {
		if !spec.Implemented {
			continue
		}
		spec := spec
		enabled := true
		switch spec.Name {
		case "find":
			enabled = caps.Find
		case "update":
			enabled = caps.Update && m.activeTab == panel.TabSkills
		case "init":
			enabled = caps.Init && m.activeTab == panel.TabSkills
		case "inspect":
			enabled = caps.Inspect && m.activeTab == panel.TabSkills
		case "remove":
			enabled = caps.Remove
		case "add":
			enabled = caps.SearchInstall && m.activeTab == panel.TabSkills
		}
		add(spec.Name, spec.Summary, spec.Usage+" "+strings.Join(spec.Aliases, " "), enabled, func(m *Model) (tea.Model, tea.Cmd) {
			return m.runRegistryCommand(spec.Name)
		})
	}

	add("Quit", "Exit skill-man", "quit exit ctrl+c q", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m, tea.Quit
	})

	return items
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
		return f.m, tea.Quit
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		f.Close()
		return f.m, nil
	case key.Matches(msg, keys.Down):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor + 1) % len(p.filtered)
		}
		return f.m, nil
	case key.Matches(msg, keys.Up):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor - 1 + len(p.filtered)) % len(p.filtered)
		}
		return f.m, nil
	case key.Matches(msg, keys.Enter), msg.Type == tea.KeyEnter:
		if len(p.filtered) == 0 {
			return f.m, nil
		}
		idx := p.filtered[p.cursor]
		item := p.all[idx]
		f.Close()
		return item.run(f.m)
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	p.refresh(f.m, p.input.Value())
	return f.m, cmd
}

func (f *paletteFeature) renderOverlay(base string) string {
	if f.ui == nil {
		return base
	}
	p := f.ui
	boxW := min(62, max(36, f.m.contentWidth()-4))
	innerW := paletteInputWidth(boxW)

	title := f.m.styles.PanelTitle.Render("Command Palette")
	input := p.input.View()
	help := f.m.styles.Hint.Render("↑↓ select · Enter run · Esc close")

	var rows []string
	maxRows := min(8, max(3, f.m.height/4))
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
			rows = append(rows, f.m.styles.ItemSelected.Render("› "+line))
		} else {
			rows = append(rows, f.m.styles.ItemDesc.Render("  "+line))
		}
	}
	if len(p.filtered) == 0 {
		rows = append(rows, f.m.styles.EmptyPreview.Render("  No matching commands"))
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		input,
		strings.Join(rows, "\n"),
		help,
	)
	box := f.m.styles.Modal.Width(boxW).Render(body)
	return lipgloss.Place(f.m.width-2, f.m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) openCommandPalette() (tea.Model, tea.Cmd) {
	return m.cmdPalette.Open()
}
