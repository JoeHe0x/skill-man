package palette

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/commands"
)

// ActionHost exposes palette catalog actions without *Model in callbacks.
type ActionHost interface {
	State() session.State
	LastState() session.State
	PromptActive() bool
	TransitionTo(session.State) bool
	ContentWidth() int
	Width() int
	Height() int
	Styles() theme.Styles
	TeaModel() tea.Model

	BeginScanAllCmd() tea.Cmd
	ActiveTab() panel.Tab
	StartListFilter() (tea.Model, tea.Cmd)
	OpenAgentFilter() (tea.Model, tea.Cmd)
	OpenHelpScreen() (tea.Model, tea.Cmd)
	GoToListingWithPreview() (tea.Model, tea.Cmd)
	SetActiveTab(panel.Tab) tea.Cmd
	StartInstallFlow() (tea.Model, tea.Cmd)
	ShowInitPrompt() (tea.Model, tea.Cmd)
	ShowAddPrompt() (tea.Model, tea.Cmd)
	HandleUpdate() (tea.Model, tea.Cmd)
	HandleInspectSelected() (tea.Model, tea.Cmd)
	HandleBindSelected() (tea.Model, tea.Cmd)
	HandleDisableSelected() (tea.Model, tea.Cmd)
	HandleRemoveSelected() (tea.Model, tea.Cmd)
	RunRegistryCommand(string) (tea.Model, tea.Cmd)
	ActivePanel() panel.Panel
	SelectedListItem() (panel.Item, bool)
	CommandSpecs() []commands.Spec
}
