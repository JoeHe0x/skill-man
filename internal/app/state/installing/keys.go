package installing

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HandleKeys routes keys while the install wizard overlay is open.
func HandleKeys(h Host, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !h.InstallWizardOpen() {
		return h.TeaModel(), nil
	}
	return h.InstallWizardHandleKey(msg)
}
