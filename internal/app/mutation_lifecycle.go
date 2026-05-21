package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

func runCommand(cmd command.Cmd) tea.Cmd {
	return command.Run(cmd)
}

func (m *Model) applyMutationResult(msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.reportError(msg.Err)
		m.updateFooterForState(m.state)
		return m, m.beginScanAllCmd()
	}
	m.clearError()
	m.status = "ready"
	m.updateFooterForState(m.state)

	var flashCmd tea.Cmd
	if msg.Message != "" {
		flashCmd = m.flashFooter(msg.Message)
	}
	if msg.SelectName == "" {
		return m, tea.Batch(flashCmd, m.beginScanAllCmd())
	}
	if msg.Kind == usecase.KindMCP {
		return m, tea.Batch(flashCmd, tea.Sequence(
			m.beginScanAllCmd(),
			func() tea.Msg { return uimsg.ReselectMCP{Name: msg.SelectName} },
		))
	}
	return m, tea.Batch(flashCmd, tea.Sequence(
		m.beginScanAllCmd(),
		func() tea.Msg { return uimsg.ReselectSkill{Name: msg.SelectName} },
	))
}
