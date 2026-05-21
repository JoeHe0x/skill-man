package app

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

func (m *Model) ShortHelp() []key.Binding {
	switch m.state {
	case stateConfirming:
		return []key.Binding{keys.Confirm, keys.Cancel}
	case stateInstalling:
		return m.installShortHelp()
	case stateFilteringAgent:
		return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.Cancel}
	case stateBindingAgent:
		return []key.Binding{keys.Up, keys.Down, keys.Toggle, keys.Enter, keys.Cancel}
	case stateInspecting:
		return []key.Binding{keys.Toggle, keys.Enter, keys.Home, keys.PgUp, keys.PgDown}
	case stateHelpOverlay:
		return []key.Binding{keys.Cancel, keys.PgUp, keys.PgDown}
	case stateCommandPalette:
		return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.Cancel}
	default:
		return m.browseShortHelp(m.state == stateSearching)
	}
}

func (m *Model) FullHelp() [][]key.Binding {
	switch m.state {
	case stateConfirming:
		return [][]key.Binding{{keys.Confirm, keys.Cancel}}
	case stateInstalling:
		return m.installFullHelp()
	case stateFilteringAgent:
		return [][]key.Binding{{keys.Up, keys.Down, keys.Enter, keys.Cancel}}
	case stateBindingAgent:
		return [][]key.Binding{{keys.Up, keys.Down, keys.Toggle, keys.Enter, keys.Cancel}}
	case stateInspecting:
		return [][]key.Binding{
			{keys.Toggle, keys.Enter, keys.Home},
			{keys.PgUp, keys.PgDown, keys.Quit},
		}
	case stateHelpOverlay:
		return [][]key.Binding{
			{keys.Cancel, keys.HelpScreen, keys.PgUp, keys.PgDown},
			{keys.Quit},
		}
	case stateCommandPalette:
		return [][]key.Binding{{keys.Up, keys.Down, keys.Enter, keys.Cancel}}
	default:
		return m.browseFullHelp()
	}
}

func (m *Model) browseShortHelp(searching bool) []key.Binding {
	caps := m.activePanel().Capabilities()
	// Keep the short footer compact; high-priority keys first so narrow terminals
	// still show del/remove (help truncates from the end).
	out := []key.Binding{keys.Palette, keys.HelpToggle, keys.Tab, keys.List}
	if caps.Remove {
		out = append(out, keys.Delete)
	}
	if caps.Find {
		out = append(out, keys.Find)
	}
	if caps.SearchInstall && m.activeTab == panel.TabSkills {
		out = append(out, keys.Add)
	}
	out = append(out, keys.Agent, keys.Reload)
	if caps.Update && m.activeTab == panel.TabSkills {
		out = append(out, keys.Update)
	}
	if searching {
		out = append(out, keys.Home)
	}
	if selected, ok := m.list.SelectedItem().(panel.Item); ok {
		if caps.Inspect && selected.CanInspect() {
			out = append(out, keys.Enter)
		}
		if caps.Disable && selected.CanDisable() {
			out = append(out, keys.Disable)
		}
		if caps.Bind && selected.CanBind() {
			out = append(out, keys.Bind)
		}
	}
	out = append(out, keys.Quit)
	return out
}

func (m *Model) browseFullHelp() [][]key.Binding {
	caps := m.activePanel().Capabilities()
	nav := []key.Binding{keys.Palette, keys.HelpToggle, keys.HelpScreen, keys.Tab, keys.ShiftTab, keys.List, keys.Home}
	ops := []key.Binding{keys.Agent, keys.Reload, keys.Quit}
	if caps.Find {
		ops = append([]key.Binding{keys.Find, keys.Filter}, ops...)
	}
	if caps.Remove {
		ops = append(ops, keys.Delete)
	}
	if caps.SearchInstall && m.activeTab == panel.TabSkills {
		ops = append([]key.Binding{keys.Add, keys.Init}, ops...)
	}
	if caps.Update && m.activeTab == panel.TabSkills {
		ops = append(ops, keys.Update)
	}
	item := []key.Binding{keys.Up, keys.Down, keys.PgUp, keys.PgDown}
	if caps.Inspect {
		item = append(item, keys.Enter)
	}
	if caps.Disable {
		item = append(item, keys.Disable)
	}
	if caps.Bind {
		item = append(item, keys.Bind)
	}
	return [][]key.Binding{nav, ops, item}
}

func (m *Model) installShortHelp() []key.Binding {
	if m.install.flow == nil {
		return []key.Binding{keys.Cancel, keys.Quit}
	}
	return m.install.flow.ShortHelp()
}

func (m *Model) installFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{keys.Up, keys.Down, keys.Enter, keys.InstallSearch, keys.Toggle, keys.Cancel},
		{keys.Quit},
	}
}

func (m *Model) renderHelpFooter() string {
	m.help.Width = max(20, m.width-6)
	return m.help.View(m)
}

func initHelpStyles(h *help.Model, s theme.Styles) {
	h.Styles.ShortKey = h.Styles.ShortKey.Foreground(s.HelpKey.GetForeground())
	h.Styles.ShortDesc = h.Styles.ShortDesc.Foreground(s.HelpDesc.GetForeground())
	h.Styles.FullKey = h.Styles.FullKey.Foreground(s.HelpKey.GetForeground())
	h.Styles.FullDesc = h.Styles.FullDesc.Foreground(s.HelpDesc.GetForeground())
}
