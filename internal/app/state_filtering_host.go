package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	statefiltering "github.com/JoeHe0x/skill-man/internal/app/state/filtering"
	statelisting "github.com/JoeHe0x/skill-man/internal/app/state/listing"
)

func (m *Model) handleAgentFilterUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return statefiltering.HandleUpdate(m, msg)
}

func (m *Model) handleAgentFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statefiltering.HandleKeys(m, msg)
}

func (m *Model) LastState() session.State { return m.lastState }

func (m *Model) AgentSelectedItem() (panel.Item, bool) {
	item, ok := m.Agent.SelectedItem().(panel.Item)
	return item, ok
}

func (m *Model) ApplyAgentFilter(id string) { statelisting.SetAgentFilter(m, id) }

func (m *Model) AgentDisplay() string { return m.agentDisplay() }

func (m *Model) AgentFilterListUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.Agent, cmd = m.Agent.Update(msg)
	return cmd
}

var _ statefiltering.Host = (*Model)(nil)
