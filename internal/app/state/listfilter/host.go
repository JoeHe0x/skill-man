package listfilter

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Host exposes inline main-list filter key handling needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	MainFilterState() list.FilterState
	MainUpdate(tea.KeyMsg) tea.Cmd
	SetFooterContext(string)
	ListFilterStatusLine() string
	SyncSelectionPreview() tea.Cmd
}
