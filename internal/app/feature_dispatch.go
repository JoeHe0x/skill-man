package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/feature"
)

// dispatchToFeatures routes the message to the first active feature.
// Returns (cmd, true) if a feature consumed the message; (nil, false) otherwise.
func (m *Model) dispatchToFeatures(msg tea.Msg) (tea.Cmd, bool) {
	for _, f := range m.features {
		if !f.Active() {
			continue
		}
		cmd, consumed := f.Update(msg)
		if consumed {
			return cmd, true
		}
	}
	return nil, false
}

// initFeatures returns the initial batch command for all registered features.
func (m *Model) initFeatures() tea.Cmd {
	var cmds []tea.Cmd
	for _, f := range m.features {
		if f.Active() {
			cmds = append(cmds, f.Init())
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// featureDeps builds the shared dependency struct for features.
func (m *Model) featureDeps() feature.Deps {
	return feature.Deps{
		CWD:  m.cwd,
		Home: m.home,
	}
}
