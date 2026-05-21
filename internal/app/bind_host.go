package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

// bindHost exposes only what the bind feature needs from Model.
type bindHost interface {
	IsBinding() bool
	ActivePanelCanBind() bool
	CWD() string
	Home() string
	Binder() usecasebind.Binder
	MCPMembersForConfigKey(key string) []*mcpdomain.Server
	TransitionTo(SessionState) bool
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

func (m *Model) IsBinding() bool { return m.state == stateBindingAgent }

func (m *Model) ActivePanelCanBind() bool {
	return m.activePanel().Capabilities().Bind
}

func (m *Model) CWD() string  { return m.cwd }
func (m *Model) Home() string { return m.home }

func (m *Model) Binder() usecasebind.Binder { return m.binder }

func (m *Model) MCPMembersForConfigKey(key string) []*mcpdomain.Server {
	return m.mcpMembersForConfigKey(key)
}

func (m *Model) TransitionTo(s SessionState) bool { return m.transitionTo(s) }

func (m *Model) SetFooterContext(s string) { m.setFooterContext(s) }

func (m *Model) SetAgentListItems(items []list.Item) { m.setAgentListItems(items) }

func (m *Model) AgentListIndex() int { return m.agentList.Index() }

func (m *Model) AgentListSelect(i int) { m.agentList.Select(i) }

func (m *Model) AgentListUpdate(msg tea.Msg) (list.Model, tea.Cmd) {
	return m.agentList.Update(msg)
}

func (m *Model) BeginScanAllCmd() tea.Cmd { return m.beginScanAllCmd() }

func (m *Model) FlashFooter(s string) tea.Cmd { return m.flashFooter(s) }

func (m *Model) ReportError(err error) { m.reportError(err) }

func (m *Model) ErrMsg() string { return m.errMsg }

// Ensure Model implements bindHost at compile time.
var _ bindHost = (*Model)(nil)
