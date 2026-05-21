package listfilter

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// HandleKeys routes keys while the main list inline filter is active.
func HandleKeys(h Host, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := h.TeaModel()
	keys := uikeys.Default

	if key.Matches(msg, keys.Quit) {
		return m, tea.Quit
	}

	prev := h.MainFilterState()
	cmd := h.MainUpdate(msg)
	if prev == list.Filtering && h.MainFilterState() != list.Filtering {
		h.SetFooterContext(h.ListFilterStatusLine())
	}
	return m, tea.Batch(cmd, h.SyncSelectionPreview())
}
