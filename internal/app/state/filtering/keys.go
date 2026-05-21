package filtering

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// HandleUpdate routes messages while the agent filter overlay is active.
func HandleUpdate(h Host, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return HandleKeys(h, msg)
	}
	return h.TeaModel(), h.AgentFilterListUpdate(msg)
}

// HandleKeys routes keys in agent-filter mode.
func HandleKeys(h Host, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := h.TeaModel()
	keys := uikeys.Default

	switch {
	case key.Matches(msg, keys.Enter):
		selected, ok := h.AgentSelectedItem()
		if !ok || selected.Meta == "" {
			return m, nil
		}
		h.ApplyAgentFilter(selected.Meta)
		h.TransitionTo(h.LastState())
		h.RefreshActiveList()
		return m, tea.Batch(
			h.FlashFooter(fmt.Sprintf("Agent filter: %s", h.AgentDisplay())),
			h.SyncSelectionPreview(),
		)

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		h.TransitionTo(h.LastState())
		return m, h.FlashFooter("Agent filter cancelled")
	}

	return m, h.AgentFilterListUpdate(msg)
}
