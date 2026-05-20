package installui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Enter         key.Binding
	Cancel        key.Binding
	Home          key.Binding
	Up            key.Binding
	Down          key.Binding
	PgUp          key.Binding
	PgDown        key.Binding
	Toggle        key.Binding
	InstallSearch key.Binding
	Find          key.Binding
}

var keys = keyMap{
	Enter:         key.NewBinding(key.WithKeys("enter")),
	Cancel:        key.NewBinding(key.WithKeys("esc")),
	Home:          key.NewBinding(key.WithKeys("esc")),
	Up:            key.NewBinding(key.WithKeys("up", "ctrl+k")),
	Down:          key.NewBinding(key.WithKeys("down", "ctrl+j")),
	PgUp:          key.NewBinding(key.WithKeys("pgup")),
	PgDown:        key.NewBinding(key.WithKeys("pgdown")),
	Toggle:        key.NewBinding(key.WithKeys(" ")),
	InstallSearch: key.NewBinding(key.WithKeys("/")),
	Find:          key.NewBinding(key.WithKeys("ctrl+f")),
}
