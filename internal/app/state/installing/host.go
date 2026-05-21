package installing

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Host exposes install-wizard key handling needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	InstallWizardOpen() bool
	InstallWizardHandleKey(tea.KeyMsg) (tea.Model, tea.Cmd)
}
