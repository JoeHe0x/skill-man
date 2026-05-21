package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/feature/palette"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
)

func (m *Model) StartListFilter() (tea.Model, tea.Cmd) { return m.startListFilter() }
func (m *Model) OpenAgentFilter() (tea.Model, tea.Cmd) { return m.handleOpenAgentFilter() }
func (m *Model) OpenHelpScreen() (tea.Model, tea.Cmd)  { return m.helpScreen.Open() }
func (m *Model) GoToListingWithPreview() (tea.Model, tea.Cmd) {
	m.transitionTo(stateListing)
	return m, m.SyncSelectionPreview()
}
func (m *Model) SetActiveTab(tab panel.Tab) tea.Cmd {
	return m.setActiveTab(tab)
}
func (m *Model) StartInstallFlow() (tea.Model, tea.Cmd) { return m.startInstallFlow() }
func (m *Model) ShowInitPrompt() (tea.Model, tea.Cmd)   { return m.showInitPrompt() }
func (m *Model) ShowAddPrompt() (tea.Model, tea.Cmd)    { return m.showAddPrompt() }
func (m *Model) HandleUpdate() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m.updateItem(panel.Item{})
	}
	return m.updateItem(item)
}
func (m *Model) HandleInspectSelected() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m, nil
	}
	return m.inspectItem(item)
}
func (m *Model) HandleBindSelected() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m, nil
	}
	return m.bind.StartFromItem(item)
}
func (m *Model) HandleDisableSelected() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m, nil
	}
	return m.disableItem(item)
}
func (m *Model) HandleRemoveSelected() (tea.Model, tea.Cmd) {
	item, ok := m.selectedPanelItem()
	if !ok {
		return m, nil
	}
	return m.removeItem(item)
}
func (m *Model) RunRegistryCommand(name string) (tea.Model, tea.Cmd) {
	return m.runRegistryCommand(name)
}
func (m *Model) ActivePanel() panel.Panel { return m.activePanel() }
func (m *Model) SelectedListItem() (panel.Item, bool) {
	item, ok := m.Main.SelectedItem().(panel.Item)
	return item, ok
}
func (m *Model) CommandSpecs() []commands.Spec { return m.registry.Specs() }

var _ palette.ActionHost = (*Model)(nil)
