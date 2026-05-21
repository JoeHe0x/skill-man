package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

// runCommand executes a command and delivers mutationCompletedMsg when done.
func runCommand(cmd command.Cmd) tea.Cmd {
	return func() tea.Msg {
		result := cmd.Execute(context.Background())
		return mutationCompletedMsg{
			err:        result.Err,
			message:    result.Message,
			selectName: result.AffectedName,
			targetTab:  result.TargetTab,
		}
	}
}

// applyMutationResult runs the standard post-mutation lifecycle: error reporting,
// status/footer update, optional flash, rescan, and reselection.
func (m *Model) applyMutationResult(msg mutationCompletedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.reportError(msg.err)
		m.updateFooterForState(m.state)
		return m, m.beginScanAllCmd()
	}
	m.clearError()
	m.status = "ready"
	m.updateFooterForState(m.state)

	var flashCmd tea.Cmd
	if msg.message != "" {
		flashCmd = m.flashFooter(msg.message)
	}
	if msg.selectName == "" {
		return m, tea.Batch(flashCmd, m.beginScanAllCmd())
	}
	if msg.targetTab == panel.TabMCP {
		return m, tea.Batch(flashCmd, tea.Sequence(
			m.beginScanAllCmd(),
			func() tea.Msg { return reselectMCPMsg{name: msg.selectName} },
		))
	}
	return m, tea.Batch(flashCmd, tea.Sequence(
		m.beginScanAllCmd(),
		func() tea.Msg { return reselectSkillMsg{name: msg.selectName} },
	))
}
