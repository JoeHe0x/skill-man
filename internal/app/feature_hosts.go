package app

import (
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
	tea "github.com/charmbracelet/bubbletea"
)

// confirmHost exposes confirm-dialog needs from Model.
type confirmHost interface {
	IsConfirming() bool
	TransitionTo(SessionState) bool
	SetFooterContext(string)
	SetStatus(string)
	PaneSizes() (int, int, int, int)
	Styles() theme.Styles
	Mutator() usecase.Mutator
	TeaModel() tea.Model
}

func (m *Model) IsConfirming() bool { return m.state == stateConfirming }

func (m *Model) SetStatus(s string) { m.status = s }

func (m *Model) PaneSizes() (int, int, int, int) { return m.paneSizes() }

func (m *Model) Styles() theme.Styles { return m.styles }

func (m *Model) Mutator() usecase.Mutator { return m.mutator }

func (m *Model) ClearError() { m.clearError() }

func (m *Model) CancelInstallFlow(hint string) { m.cancelInstallFlow(hint) }

func (m *Model) HandleInspectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleInspectingKeys(msg)
}

func (m *Model) HandleAgentFilterUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.handleAgentFilterUpdate(msg)
}

func (m *Model) TeaModel() tea.Model { return m }

// installHost exposes install-flow needs from Model.
type installHost interface {
	ActiveTab() panel.Tab
	ActivePanelSearchInstall() bool
	CWD() string
	Home() string
	Width() int
	Height() int
	AgentIDs() []string
	ErrMsg() string
	TransitionTo(SessionState) bool
	State() SessionState
	SetFooterContext(string)
	SetStatus(string)
	ClearError()
	ReportError(error)
	PaneSizes() (int, int, int, int)
	PaneSizesFor(int) (int, int, int, int)
	Styles() theme.Styles
	BeginScanAllCmd() tea.Cmd
	FlashFooter(string) tea.Cmd
	TeaModel() tea.Model
}

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

// promptHost exposes prompt overlay needs from Model.
type promptHost interface {
	PromptActive() bool
	State() SessionState
	CancelInstallFlow(string)
	SetFooterContext(string)
	Styles() theme.Styles
	TeaModel() tea.Model
}

func (m *Model) PromptActive() bool { return m.prompt != nil && m.prompt.Active() }

func (m *Model) State() SessionState { return m.state }

// helpHost exposes help overlay needs from Model.
type helpHost interface {
	IsHelpOverlay() bool
	TransitionTo(SessionState) bool
	LastState() SessionState
	ContentWidth() int
	ChromeHeights() (int, int)
	Width() int
	Height() int
	SetFooterContext(string)
	Styles() theme.Styles
	TeaModel() tea.Model
}

func (m *Model) IsHelpOverlay() bool { return m.state == stateHelpOverlay }

func (m *Model) LastState() SessionState { return m.lastState }

func (m *Model) ContentWidth() int { return m.contentWidth() }

func (m *Model) ChromeHeights() (int, int) { return m.chromeHeights() }

// inspectHost exposes skill inspect flow needs from Model.
type inspectHost interface {
	IsInspecting() bool
	HandleInspectingKeys(tea.KeyMsg) (tea.Model, tea.Cmd)
	TeaModel() tea.Model
}

func (m *Model) IsInspecting() bool { return m.state == stateInspecting }

// agentFilterHost exposes agent filter overlay needs from Model.
type agentFilterHost interface {
	IsFilteringAgent() bool
	HandleAgentFilterUpdate(tea.Msg) (tea.Model, tea.Cmd)
	TeaModel() tea.Model
}

func (m *Model) IsFilteringAgent() bool { return m.state == stateFilteringAgent }

// paletteHost exposes command palette chrome needs.
type paletteHost interface {
	State() SessionState
	LastState() SessionState
	PromptActive() bool
	TransitionTo(SessionState) bool
	ContentWidth() int
	Width() int
	Height() int
	Styles() theme.Styles
	TeaModel() tea.Model
}

var (
	_ confirmHost     = (*Model)(nil)
	_ installHost     = (*Model)(nil)
	_ promptHost      = (*Model)(nil)
	_ helpHost        = (*Model)(nil)
	_ inspectHost     = (*Model)(nil)
	_ agentFilterHost = (*Model)(nil)
	_ paletteHost     = (*Model)(nil)
)
