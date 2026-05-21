package filter

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

// Host exposes agent-filter overlay needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	IsFilteringAgent() bool
	TransitionTo(session.State) bool
	LastState() session.State
	AgentIDs() []string
	ActiveTab() panel.Tab
	AllAgents() []agent.Agent
	CWD() string
	Home() string
	SetAgentListItems(items []list.Item)
	AgentListSelect(i int)
	AgentListSetSize(w, h int)
	AgentListView() string
	SetFooterContext(string)
	PaneSizes() (leftW, mainH, rightW, rightH int)
	Width() int
	Height() int
	Styles() theme.Styles
	AgentSelectedItem() (panel.Item, bool)
	AgentFilterListUpdate(tea.Msg) tea.Cmd
	ApplyAgentFilter(id string)
	RefreshActiveList()
	AgentDisplay() string
	FlashFooter(string) tea.Cmd
	SyncSelectionPreview() tea.Cmd
}
