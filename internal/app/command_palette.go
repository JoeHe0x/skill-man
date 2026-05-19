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
	w := min(56, max(24, outerWidth-8))
	return w
}

func (m *Model) canOpenPalette() bool {
	if m.prompt != nil || m.palette != nil {
		return false
	}
	switch m.state {
	case stateHome, stateListing, stateSearching:
		return true
	default:
		return false
	}
}

func (m *Model) openCommandPalette() (tea.Model, tea.Cmd) {
	if !m.canOpenPalette() {
		return m, nil
	}
	m.lastState = m.state
	m.state = stateCommandPalette
	p := newCommandPalette(m.contentWidth())
	p.refresh(m, "")
	m.palette = p
	return m, textinput.Blink
}

func (m *Model) closeCommandPalette() {
	if m.palette == nil {
		return
	}
	m.palette = nil
	if m.state == stateCommandPalette {
		m.state = m.lastState
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
		return m.handleReload()
	})
	add("Filter list", "Fuzzy filter in the skills/MCP list", "find filter search ctrl+f /", caps.Find, func(m *Model) (tea.Model, tea.Cmd) {
		return m.startListFilter()
	})
	add("Agent filter", "Choose which agent context to show", "agent filter ctrl+a", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m.handleOpenAgentFilter()
	})
	add("Command reference", "Show help and slash commands", "help commands f1 ?", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m.handleHelp()
	})
	add("Focus list", "Return to the extension list", "list home ctrl+l", true, func(m *Model) (tea.Model, tea.Cmd) {
		return m.handleList()
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

	if sel, ok := m.list.SelectedItem().(listItem); ok {
		switch sel.kind {
		case itemKindSkill:
			if caps.Inspect {
				add("Inspect skill", "Browse files for "+sel.title, "inspect enter tree", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleInspectSelected()
				})
			}
			if caps.Bind {
				add("Bind agents", "Manage agent bindings for "+sel.title, "bind b", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleBindSelected()
				})
			}
			if caps.Disable {
				add("Toggle disable", "Enable or disable "+sel.title, "toggle disable x", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleDisableSelected()
				})
			}
			if caps.Remove {
				add("Remove skill", "Delete "+sel.title+" (confirmed)", "remove delete del", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleRemoveSelected()
				})
			}
		case itemKindMCP:
			if caps.Bind {
				add("Bind MCP", "Manage MCP bindings for "+sel.title, "bind b", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleBindSelected()
				})
			}
			if caps.Disable {
				add("Toggle MCP", "Enable or disable "+sel.title, "toggle disable x", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleDisableSelected()
				})
			}
			if caps.Remove {
				add("Remove MCP", "Delete "+sel.title+" (confirmed)", "remove delete del", true, func(m *Model) (tea.Model, tea.Cmd) {
					return m.handleRemoveSelected()
				})
			}
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
		return m.handleHelp()
	case "list":
		return m.handleList()
	case "find":
		return m.startListFilter()
	case "reload":
		return m.handleReload()
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

func (m *Model) handlePaletteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.palette == nil {
		return m, nil
	}
	p := m.palette

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		m.closeCommandPalette()
		return m, nil
	case key.Matches(msg, keys.Down):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor + 1) % len(p.filtered)
		}
		return m, nil
	case key.Matches(msg, keys.Up):
		if len(p.filtered) > 0 {
			p.cursor = (p.cursor - 1 + len(p.filtered)) % len(p.filtered)
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		if len(p.filtered) == 0 {
			return m, nil
		}
		idx := p.filtered[p.cursor]
		item := p.all[idx]
		m.closeCommandPalette()
		return item.run(m)
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	p.refresh(m, p.input.Value())
	return m, cmd
}

func (m *Model) renderPaletteOverlay(base string) string {
	if m.palette == nil {
		return base
	}
	p := m.palette
	boxW := min(62, max(36, m.contentWidth()-4))
	innerW := paletteInputWidth(boxW)

	title := m.styles.panelTitle.Render("Command Palette")
	input := p.input.View()
	help := m.styles.hint.Render("↑↓ select · Enter run · Esc close")

	var rows []string
	maxRows := min(8, max(3, m.height/4))
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
			rows = append(rows, m.styles.itemSelected.Render("› "+line))
		} else {
			rows = append(rows, m.styles.itemDesc.Render("  "+line))
		}
	}
	if len(p.filtered) == 0 {
		rows = append(rows, m.styles.emptyPreview.Render("  No matching commands"))
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		input,
		strings.Join(rows, "\n"),
		help,
	)
	box := m.styles.modal.Width(boxW).Render(body)
	return lipgloss.Place(m.width-2, m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}
