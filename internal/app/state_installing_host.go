package app

import (
	tea "github.com/charmbracelet/bubbletea"

	stateinstalling "github.com/JoeHe0x/skill-man/internal/app/state/installing"
)

func (m *Model) handleInstallingUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return stateinstalling.HandleKeys(m, msg)
}

func (m *Model) InstallWizardOpen() bool {
	return m.install.WizardOpen()
}

func (m *Model) InstallWizardHandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.install.HandleUIMsg(msg)
}

var _ stateinstalling.Host = (*Model)(nil)
