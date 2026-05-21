package prompt

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

// Host exposes prompt overlay needs from the app Model.
type Host interface {
	State() session.State
	CancelInstallFlow(string)
	SetFooterContext(string)
	Styles() theme.Styles
	TeaModel() tea.Model
}

// Action runs after the user confirms prompt input.
type Action func(text string) tea.Cmd
