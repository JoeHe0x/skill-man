package confirm

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

// Host exposes confirm-dialog needs from the app Model.
type Host interface {
	IsConfirming() bool
	TransitionTo(session.State) bool
	SetFooterContext(string)
	SetStatus(string)
	PaneSizes() (int, int, int, int)
	Styles() theme.Styles
	Mutator() usecase.Mutator
	TeaModel() tea.Model
}
