package bind

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

// Host exposes bind-dialog needs from the app Model.
type Host interface {
	IsBinding() bool
	ActivePanelCanBind() bool
	CWD() string
	Home() string
	Binder() usecasebind.Binder
	MCPMembersForConfigKey(key string) []*mcpdomain.Server
	TransitionTo(session.State) bool
	SetFooterContext(string)
	SetAgentListItems([]list.Item)
	AgentListIndex() int
	AgentListSelect(int)
	AgentListUpdate(tea.Msg) (list.Model, tea.Cmd)
	BeginScanAllCmd() tea.Cmd
	FlashFooter(string) tea.Cmd
	ReportError(error)
	ErrMsg() string
	TeaModel() tea.Model
}
