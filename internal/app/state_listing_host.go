package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	featprompt "github.com/JoeHe0x/skill-man/internal/app/feature/prompt"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	statelisting "github.com/JoeHe0x/skill-man/internal/app/state/listing"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statelisting.HandleKeys(m, msg)
}

func (m *Model) OpenHelpOverlay() (tea.Model, tea.Cmd) { return m.openHelpOverlay() }

func (m *Model) OpenCommandPalette() (tea.Model, tea.Cmd) { return m.openCommandPalette() }

func (m *Model) SwitchExtensionTab(reverse bool) tea.Cmd { return m.switchExtensionTab(reverse) }

func (m *Model) MainFilterState() list.FilterState { return m.Main.FilterState() }

func (m *Model) MainUpdate(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.Main, cmd = m.Main.Update(msg)
	return cmd
}

func (m *Model) StaticPreview() string { return m.activePanel().StaticPreview() }

func (m *Model) SetPreviewContent(s string) { m.Preview.SetContent(s) }

func (m *Model) ToggleHelpAll() { m.help.ShowAll = !m.help.ShowAll }

func (m *Model) SetFocusedList() { m.focusedPane = focusPaneList }

func (m *Model) SetFocusedPreview() { m.focusedPane = focusPanePreview }

func (m *Model) ShowPrompt(label, placeholder string, action func(text string) tea.Cmd) tea.Cmd {
	return m.showPrompt(label, placeholder, featprompt.Action(action))
}

func (m *Model) HidePrompt() { m.hidePrompt() }

func (m *Model) RefreshActiveList() { m.refreshActiveList() }

func (m *Model) SetMainListItems(items []panel.Item) {
	m.setMainListItems(panelToListItems(items))
}

func (m *Model) SetAgentIDs(ids []string) { m.agentIDs = ids }

func (m *Model) ActiveAgents() []agent.Agent { return m.activeAgents() }

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowFindPrompt(m) }

func (m *Model) showAddPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowAddPrompt(m) }

func (m *Model) showInitPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowInitPrompt(m) }

func (m *Model) setAgentFilter(id string) { statelisting.SetAgentFilter(m, id) }

var (
	_ statelisting.Host       = (*Model)(nil)
	_ statelisting.PromptHost = (*Model)(nil)
)
