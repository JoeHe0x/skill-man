package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
)

// paletteActionHost exposes palette catalog actions without *Model in callbacks.
type paletteActionHost interface {
	paletteHost

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

func (m *Model) StartListFilter() (tea.Model, tea.Cmd) { return m.startListFilter() }
func (m *Model) OpenAgentFilter() (tea.Model, tea.Cmd) { return m.handleOpenAgentFilter() }
func (m *Model) OpenHelpScreen() (tea.Model, tea.Cmd)  { return m.helpScreen.Open() }
func (m *Model) GoToListingWithPreview() (tea.Model, tea.Cmd) {
	m.transitionTo(stateListing)
	return m, m.syncSelectionPreview()
}
func (m *Model) SetActiveTab(tab panel.Tab) tea.Cmd {
	return m.setActiveTab(tab)
}
func (m *Model) StartInstallFlow() (tea.Model, tea.Cmd) { return m.startInstallFlow() }
func (m *Model) ShowInitPrompt() (tea.Model, tea.Cmd)   { return m.showInitPrompt() }
func (m *Model) ShowAddPrompt() (tea.Model, tea.Cmd)    { return m.showAddPrompt() }
func (m *Model) HandleUpdate() (tea.Model, tea.Cmd)     { return m.handleUpdate() }
func (m *Model) HandleInspectSelected() (tea.Model, tea.Cmd) {
	return m.handleInspectSelected()
}
func (m *Model) HandleBindSelected() (tea.Model, tea.Cmd) { return m.handleBindSelected() }
func (m *Model) HandleDisableSelected() (tea.Model, tea.Cmd) {
	return m.handleDisableSelected()
}
func (m *Model) HandleRemoveSelected() (tea.Model, tea.Cmd) {
	return m.handleRemoveSelected()
}
func (m *Model) RunRegistryCommand(name string) (tea.Model, tea.Cmd) {
	return m.runRegistryCommand(name)
}
func (m *Model) ActivePanel() panel.Panel { return m.activePanel() }
func (m *Model) SelectedListItem() (panel.Item, bool) {
	item, ok := m.list.SelectedItem().(panel.Item)
	return item, ok
}
func (m *Model) CommandSpecs() []commands.Spec { return m.registry.Specs() }

var _ paletteActionHost = (*Model)(nil)
