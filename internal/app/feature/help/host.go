package help

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

// Host exposes help overlay needs from the app Model.
type Host interface {
	IsHelpOverlay() bool
	TransitionTo(session.State) bool
	LastState() session.State
	ContentWidth() int
	ChromeHeights() (int, int)
	Width() int
	Height() int
	SetFooterContext(string)
	Styles() theme.Styles
	TeaModel() tea.Model
}
