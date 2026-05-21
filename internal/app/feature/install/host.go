package install

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

// Host exposes install-flow needs from the app Model.
type Host interface {
	ActiveTab() panel.Tab
	ActivePanelSearchInstall() bool
	CWD() string
	Home() string
	Width() int
	Height() int
	AgentIDs() []string
	ErrMsg() string
	TransitionTo(session.State) bool
	State() session.State
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
