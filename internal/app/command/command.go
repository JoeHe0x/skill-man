package command

import (
	"context"

	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

// Cmd is an executable mutation operation on extensions (Bubble Tea adapter).
type Cmd interface {
	// Label returns a human-readable label for the operation (e.g. "removed", "disabled").
	Label() string
	// Execute performs the operation and returns its result.
	Execute(ctx context.Context) Result
}

// Result is the UI-layer outcome of executing a command.
type Result struct {
	AffectedName string
	Message      string
	Err          error
	Kind         usecase.Kind
}

func resultFrom(out usecase.Outcome) Result {
	return Result{
		AffectedName: out.AffectedName,
		Message:      out.Message,
		Err:          out.Err,
		Kind:         out.Kind,
	}
}
