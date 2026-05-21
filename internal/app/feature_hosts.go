package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	featbind "github.com/JoeHe0x/skill-man/internal/app/feature/bind"
	featconfirm "github.com/JoeHe0x/skill-man/internal/app/feature/confirm"
	featfilter "github.com/JoeHe0x/skill-man/internal/app/feature/filter"
	feathelp "github.com/JoeHe0x/skill-man/internal/app/feature/help"
	featinspect "github.com/JoeHe0x/skill-man/internal/app/feature/inspect"
	featinstall "github.com/JoeHe0x/skill-man/internal/app/feature/install"
	featpalette "github.com/JoeHe0x/skill-man/internal/app/feature/palette"
	featprompt "github.com/JoeHe0x/skill-man/internal/app/feature/prompt"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	stateinspect "github.com/JoeHe0x/skill-man/internal/app/state/inspect"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

func (m *Model) IsConfirming() bool { return m.state == stateConfirming }

func (m *Model) SetStatus(s string) { m.status = s }

func (m *Model) PaneSizes() (int, int, int, int) { return m.paneSizes() }

func (m *Model) Styles() theme.Styles { return m.styles }

func (m *Model) Mutator() usecase.Mutator { return m.mutator }

func (m *Model) ClearError() { m.clearError() }

func (m *Model) CancelInstallFlow(hint string) { m.install.CancelFlow(hint) }

func (m *Model) TeaModel() tea.Model { return m }

func (m *Model) ActiveTab() panel.Tab { return m.activeTab }

func (m *Model) ActivePanelSearchInstall() bool {
	return m.activePanel().Capabilities().SearchInstall
}

func (m *Model) Width() int  { return m.width }
func (m *Model) Height() int { return m.height }

func (m *Model) AgentIDs() []string { return m.agentIDs }

func (m *Model) PaneSizesFor(mainHeight int) (int, int, int, int) {
	return m.paneSizesFor(mainHeight)
}

func (m *Model) PromptActive() bool { return m.prompt.Active() }

func (m *Model) State() session.State { return m.state }

func (m *Model) IsHelpOverlay() bool { return m.state == stateHelpOverlay }

func (m *Model) ContentWidth() int { return m.contentWidth() }

func (m *Model) ChromeHeights() (int, int) { return m.chromeHeights() }

func (m *Model) IsInspecting() bool { return m.state == stateInspecting }

func (m *Model) PreviewWidth() int {
	w := m.Preview.Width
	if w == 0 {
		return max(40, m.width/2)
	}
	return w
}

func (m *Model) PreviewGenPtr() *int { return &m.PreviewGen }

func (m *Model) SetTreeRoot(path string) { m.Tree.SetRoot(path) }

func (m *Model) HandleInspectSelected() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m, nil
	}
	return m.inspect.EnterFromItem(item)
}

func (m *Model) IsFilteringAgent() bool { return m.state == stateFilteringAgent }

func (m *Model) AllAgents() []agent.Agent { return m.allAgents }

func (m *Model) AgentListSelect(i int) { m.Agent.Select(i) }

func (m *Model) AgentListSetSize(w, h int) { m.Agent.SetSize(w, h) }

func (m *Model) AgentListView() string { return m.Agent.View() }

func (m *Model) OpenAgentFilter() (tea.Model, tea.Cmd) { return m.agentFilter.Open() }

func (m *Model) TransitionTo(s session.State) bool { return m.transitionTo(s) }

func (m *Model) SetFooterContext(s string) { m.setFooterContext(s) }

func (m *Model) BeginScanAllCmd() tea.Cmd { return m.beginScanAllCmd() }

func (m *Model) FlashFooter(s string) tea.Cmd { return m.flashFooter(s) }

func (m *Model) ReportError(err error) { m.reportError(err) }

func (m *Model) ErrMsg() string { return m.errMsg }

func (m *Model) CWD() string  { return m.cwd }
func (m *Model) Home() string { return m.home }

func (m *Model) Binder() usecasebind.Binder { return m.binder }

func (m *Model) MCPMembersForConfigKey(key string) []*mcpdomain.Server {
	return m.mcpMembersForConfigKey(key)
}

func (m *Model) IsBinding() bool { return m.state == stateBindingAgent }

func (m *Model) ActivePanelCanBind() bool {
	return m.activePanel().Capabilities().Bind
}

func (m *Model) SetAgentListItems(items []list.Item) { m.setAgentListItems(items) }

func (m *Model) AgentListIndex() int { return m.Agent.Index() }

func (m *Model) AgentListUpdate(msg tea.Msg) (list.Model, tea.Cmd) {
	return m.Agent.Update(msg)
}

var (
	_ featinstall.Host       = (*Model)(nil)
	_ featbind.Host          = (*Model)(nil)
	_ featconfirm.Host       = (*Model)(nil)
	_ featprompt.Host        = (*Model)(nil)
	_ feathelp.Host          = (*Model)(nil)
	_ featpalette.ActionHost = (*Model)(nil)
	_ featinspect.Host       = (*Model)(nil)
	_ featfilter.Host        = (*Model)(nil)
	_ stateinspect.Host      = (*Model)(nil)
)
