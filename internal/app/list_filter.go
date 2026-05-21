package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) configureMainList() {
	m.Main.SetFilteringEnabled(true)
	m.Main.SetShowPagination(true)
	m.Main.SetStatusBarItemName("item", "items")
	m.Main.KeyMap.Filter = keys.Filter
	m.Main.KeyMap.ClearFilter = keys.Home
}

func (m *Model) startListFilter() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Find {
		return m, m.flashFooter("Find is not available for this tab")
	}
	m.transitionTo(stateListing)
	m.focusedPane = focusPaneList
	var cmd tea.Cmd
	m.Main, cmd = m.Main.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	return m, cmd
}
