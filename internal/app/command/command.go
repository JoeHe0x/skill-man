package command

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

// Cmd is an executable mutation operation on extensions.
type Cmd interface {
	// Label returns a human-readable label for the operation (e.g. "removed", "disabled").
	Label() string
	// Execute performs the operation and returns its result.
	Execute(ctx context.Context) Result
}

// Result is the outcome of executing a command.
type Result struct {
	AffectedName string
	Message      string
	Err          error
	TargetTab    panel.Tab
}
