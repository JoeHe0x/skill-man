package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	statelistfilter "github.com/JoeHe0x/skill-man/internal/app/state/listfilter"
)

func (m *Model) handleListFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statelistfilter.HandleKeys(m, msg)
}

func (m *Model) ListFilterStatusLine() string {
	n := visiblePanelListCount(m.Main.VisibleItems())
	if m.Main.FilterValue() != "" {
		return fmt.Sprintf("filter %q → %d item(s)", m.Main.FilterValue(), n)
	}
	return fmt.Sprintf("%d item(s)", n)
}

var _ statelistfilter.Host = (*Model)(nil)

// listFilterActive reports whether inline main-list filtering should consume keys.
func (m *Model) listFilterActive() bool {
	if m.state == stateInstalling || m.state == stateBindingAgent ||
		m.state == stateFilteringAgent || m.state == stateConfirming ||
		m.state == stateInspecting || m.state == stateCommandPalette || m.prompt.Active() {
		return false
	}
	return m.Main.FilterState() == list.Filtering
}
