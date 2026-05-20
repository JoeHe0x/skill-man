package feature

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

// Feature is a self-contained UI component that owns its own state and
// can consume messages before the base Model.
type Feature interface {
	// Name returns a human-readable label for debugging.
	Name() string

	// Active reports whether the feature should receive messages.
	Active() bool

	// Update handles a message. Returns (commands, consumed).
	// If consumed is true, the message won't be passed to other features or the base model.
	Update(msg tea.Msg) (tea.Cmd, bool)

	// View returns the rendered content for this feature.
	View(width, height int) string

	// Init returns the initial command when the feature activates.
	Init() tea.Cmd
}

// Deps provides shared dependencies for features.
type Deps struct {
	CWD  string
	Home string
}

// Result is returned by a feature when it completes.
type Result struct {
	Tab  panel.Tab
	Name string // affected entity name for reselection
}
