package filtering

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

// Host exposes agent-filter overlay key handling needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	TransitionTo(session.State) bool
	LastState() session.State
	AgentSelectedItem() (panel.Item, bool)
	AgentFilterListUpdate(tea.Msg) tea.Cmd
	ApplyAgentFilter(id string)
	RefreshActiveList()
	AgentDisplay() string
	FlashFooter(string) tea.Cmd
	SyncSelectionPreview() tea.Cmd
}
