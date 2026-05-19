package app

import "github.com/charmbracelet/bubbles/key"

// keyMap defines the keybindings for the application
type keyMap struct {
	Quit          key.Binding
	Home          key.Binding
	HelpToggle    key.Binding
	HelpScreen    key.Binding
	InstallSearch key.Binding
	Up            key.Binding
	Down          key.Binding
	PgUp          key.Binding
	PgDown        key.Binding
	List          key.Binding
	Find          key.Binding
	Filter        key.Binding
	Agent         key.Binding
	Reload        key.Binding
	Update        key.Binding
	Enter         key.Binding
	Delete        key.Binding
	Confirm       key.Binding
	Cancel        key.Binding
	Disable       key.Binding
	Bind          key.Binding
	Toggle        key.Binding
	Tab           key.Binding
	ShiftTab      key.Binding
	Add           key.Binding
	Init          key.Binding
	Palette       key.Binding
}

// defaultKeyMap returns the default keybindings
var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Home: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "home"),
	),
	HelpToggle: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more keys"),
	),
	HelpScreen: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("f1", "commands"),
	),
	InstallSearch: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "new search"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "ctrl+k"),
		key.WithHelp("↑/ctrl+k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "ctrl+j"),
		key.WithHelp("↓/ctrl+j", "down"),
	),
	PgUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PgDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
	List: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "list"),
	),
	Find: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "filter"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Agent: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "agent"),
	),
	Reload: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "reload"),
	),
	Update: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "update"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "inspect"),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete", "del"),
		key.WithHelp("del", "remove"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n/esc", "cancel"),
	),
	Disable: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "toggle disable"),
	),
	Bind: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "bind agents"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch skills/mcp"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "switch skills/mcp"),
	),
	Add: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "search/install"),
	),
	Init: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new skill"),
	),
	Palette: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "palette"),
	),
}
