package command

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

// Run executes cmd asynchronously and delivers MutationCompleted.
func Run(cmd Cmd) tea.Cmd {
	return func() tea.Msg {
		result := cmd.Execute(context.Background())
		return uimsg.MutationCompleted{
			Err:        result.Err,
			Message:    result.Message,
			SelectName: result.AffectedName,
			Kind:       result.Kind,
		}
	}
}
