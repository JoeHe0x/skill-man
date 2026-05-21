package inspect

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// HandleKeys routes keys while browsing the skill file tree.
func HandleKeys(h Host, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := h.TeaModel()
	keys := uikeys.Default
	switch {
	case key.Matches(msg, keys.Home):
		h.TransitionTo(session.Listing)
		return m, tea.Batch(h.FlashFooter("Returned to skill list"), h.SyncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		return m, h.PreviewUpdate(msg)
	}

	oldSelected := h.TreeSelected()
	_, cmd := h.TreeUpdate(msg)
	newSelected := h.TreeSelected()

	if newSelected.Path != "" && newSelected.Path != oldSelected.Path && !newSelected.IsDir {
		return m, tea.Batch(cmd, h.PreviewFileCmd(newSelected.Path))
	}

	return m, cmd
}
