package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) configureMainList() {
	m.list.SetFilteringEnabled(true)
	m.list.SetShowStatusBar(true)
	m.list.SetShowPagination(true)
	m.list.SetStatusBarItemName("item", "items")
	m.list.KeyMap.Filter = keys.Filter
	m.list.KeyMap.ClearFilter = keys.Home
}

func (m *Model) listFilterActive() bool {
	if m.state == stateInstalling || m.state == stateBindingAgent ||
		m.state == stateFilteringAgent || m.state == stateConfirming ||
		m.state == stateInspecting || m.state == stateCommandPalette || m.prompt != nil {
		return false
	}
	return m.list.FilterState() == list.Filtering
}

func (m *Model) startListFilter() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Find {
		return m, m.flashFooter("Find is not available for this tab")
	}
	m.state = stateListing
	m.lastState = stateListing
	m.focusedPane = focusPaneList
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	return m, cmd
}

func (m *Model) handleListFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		return m, tea.Quit
	}
	prev := m.list.FilterState()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if prev == list.Filtering && m.list.FilterState() != list.Filtering {
		m.setFooterContext(m.listFilterStatusLine())
	}
	return m, tea.Batch(cmd, m.syncSelectionPreview())
}

func (m *Model) listFilterStatusLine() string {
	n := len(m.list.VisibleItems())
	if m.list.FilterValue() != "" {
		return fmt.Sprintf("filter %q → %d item(s)", m.list.FilterValue(), n)
	}
	return fmt.Sprintf("%d item(s)", n)
}
