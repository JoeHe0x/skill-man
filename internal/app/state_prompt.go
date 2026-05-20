package app

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.hidePrompt()
		if m.state == stateInstalling {
			m.cancelInstallFlow("Install cancelled")
			return m, nil
		}
		m.setFooterContext("Cancelled")
		return m, nil
	case key.Matches(msg, keys.Enter):
		text := m.prompt.input.Value()
		cmd := m.prompt.action(m, text)
		return m, cmd
	}
	var cmd tea.Cmd
	m.prompt.input, cmd = m.prompt.input.Update(msg)
	return m, cmd
}
