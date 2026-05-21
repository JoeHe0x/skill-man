package listing

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

// Host exposes listing/home key handling needs from the app Model.
type Host interface {
	TeaModel() tea.Model

	MainFilterState() list.FilterState
	MainUpdate(tea.KeyMsg) tea.Cmd
	ClearError()
	TransitionTo(session.State) bool
	ActivePanel() panel.Panel
	StaticPreview() string
	SetPreviewContent(string)
	SyncSelectionPreview() tea.Cmd
	ToggleHelpAll()
	OpenHelpOverlay() (tea.Model, tea.Cmd)
	OpenCommandPalette() (tea.Model, tea.Cmd)
	SetFocusedList()
	SetFocusedPreview()
	PreviewUpdate(tea.KeyMsg) tea.Cmd
	SwitchExtensionTab(reverse bool) tea.Cmd
	StartListFilter() (tea.Model, tea.Cmd)
	OpenAgentFilter() (tea.Model, tea.Cmd)
	BeginScanAllCmd() tea.Cmd
	HandleUpdate() (tea.Model, tea.Cmd)
	HandleInspectSelected() (tea.Model, tea.Cmd)
	HandleBindSelected() (tea.Model, tea.Cmd)
	HandleDisableSelected() (tea.Model, tea.Cmd)
	HandleRemoveSelected() (tea.Model, tea.Cmd)
	StartInstallFlow() (tea.Model, tea.Cmd)
	ShowInitPrompt() (tea.Model, tea.Cmd)
	SetFooterContext(string)
	ActiveTab() panel.Tab
}
