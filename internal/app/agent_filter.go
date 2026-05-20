package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func currentAgentFilterID(agentIDs []string) string {
	if len(agentIDs) == 0 {
		return "all"
	}
	return agentIDs[0]
}

func newAgentFilterListItems(agents []agent.Agent, currentID string) []list.Item {
	items := make([]list.Item, 0, len(agents)+1)
	items = append(items, panel.Item{
		Kind:  panel.ItemMessage,
		Title: filterAgentTitle("All agents", currentID == "all"),
		Desc:  "Show skills and MCP for every agent",
		Meta:  "all",
	})
	for _, a := range agents {
		active := strings.EqualFold(a.ID, currentID)
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: filterAgentTitle(a.Name, active),
			Desc:  bindAgentDesc(a),
			Meta:  a.ID,
		})
	}
	return items
}

func filterAgentTitle(name string, active bool) string {
	if active {
		return "● " + name
	}
	return "  " + name
}

func (m *Model) agentsForFilterDialog() []agent.Agent {
	if m.activeTab == panel.TabSkills {
		return agent.AgentsWithLocalSkillDir(m.allAgents, m.cwd, m.home)
	}
	return m.allAgents
}

func (m *Model) handleOpenAgentFilter() (tea.Model, tea.Cmd) {
	m.transitionTo(stateFilteringAgent)

	current := currentAgentFilterID(m.agentIDs)
	visible := m.agentsForFilterDialog()
	items := newAgentFilterListItems(visible, current)
	m.setAgentListItems(items)

	selIdx := 0
	for i, item := range items {
		li, ok := item.(panel.Item)
		if ok && strings.EqualFold(li.Meta, current) {
			selIdx = i
			break
		}
	}
	m.agentList.Select(selIdx)

	hint := "↑↓: select agent | Enter: apply filter | Esc: cancel"
	if m.activeTab == panel.TabSkills {
		hint = "↑↓: select agent (local skills dir only) | Enter: apply | Esc: cancel"
	}
	m.setFooterContext(hint)
	return m, nil
}

func (m *Model) handleAgentFilterUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleAgentFilterKeys(msg)
	}
	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) handleAgentFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		selected, ok := m.agentList.SelectedItem().(panel.Item)
		if !ok || selected.Meta == "" {
			return m, nil
		}
		m.setAgentFilter(selected.Meta)
		m.transitionTo(m.lastState)
		m.refreshActiveList()
		return m, tea.Batch(m.flashFooter(fmt.Sprintf("Agent filter: %s", m.agentDisplay())), m.syncSelectionPreview())

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		m.transitionTo(m.lastState)
		return m, m.flashFooter("Agent filter cancelled")
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) renderAgentFilterDialog() string {
	leftWidth, _, _, _ := m.paneSizes()
	dialogWidth := min(max(44, leftWidth-4), 56)
	if dialogWidth > leftWidth-2 {
		dialogWidth = max(20, leftWidth-2)
	}
	dialogHeight := min(max(16, m.height-8), 28)
	innerWidth := dialogWidth - 4
	listHeight := dialogHeight - 8
	if listHeight < 4 {
		listHeight = 4
	}

	m.agentList.SetSize(innerWidth, listHeight)
	subtitle := "Filter skills and MCP by agent"
	if m.activeTab == panel.TabSkills {
		subtitle = "Agents with a local skills directory"
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.panelTitle.Render("Agent filter"),
		m.styles.hint.Render(subtitle),
		m.agentList.View(),
	)
	return m.styles.modal.Width(dialogWidth).Render(body)
}

func (m *Model) renderAgentFilterDialogArea() string {
	leftWidth, mainHeight, _, _ := m.paneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, m.renderAgentFilterDialog())
}
