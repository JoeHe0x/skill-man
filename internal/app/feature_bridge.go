package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func (m *Model) startInstallFlow() (tea.Model, tea.Cmd) {
	return m.install.StartFlow()
}

func (m *Model) backgroundInstallActive() bool {
	return m.install.BackgroundActive()
}

func (m *Model) renderInstallDialogArea() string {
	return m.install.RenderDialogArea()
}

func (m *Model) renderBackgroundInstallOverlay(main string, mainHeight int) string {
	return m.install.RenderBackgroundOverlay(main, mainHeight)
}

func (m *Model) syncInstallHint() {
	m.install.SyncHint()
}

func (m *Model) openCommandPalette() (tea.Model, tea.Cmd) {
	return m.cmdPalette.Open()
}

func (m *Model) openHelpOverlay() (tea.Model, tea.Cmd) {
	return m.helpScreen.Open()
}

func (m *Model) runRegistryCommand(name string) (tea.Model, tea.Cmd) {
	switch name {
	case "help":
		return m.helpScreen.Open()
	case "list":
		m.transitionTo(stateListing)
		return m, m.SyncSelectionPreview()
	case "find":
		return m.startListFilter()
	case "reload":
		return m, m.beginScanAllCmd()
	case "add":
		return m.showAddPrompt()
	case "remove":
		return m.HandleRemoveSelected()
	case "update":
		return m.HandleUpdate()
	case "init":
		if m.activeTab == panel.TabSkills && m.activePanel().Capabilities().Init {
			return m.showInitPrompt()
		}
		return m, m.flashFooter("Init is only available on the Skills tab")
	case "agent":
		return m.handleOpenAgentFilter()
	case "inspect":
		return m.HandleInspectSelected()
	case "quit":
		return m, tea.Quit
	default:
		return m, m.flashFooter("Unknown command: " + name)
	}
}
