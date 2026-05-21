package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	featbind "github.com/JoeHe0x/skill-man/internal/app/feature/bind"
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
			Desc:  featbind.AgentDesc(a),
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
	m.Agent.Select(selIdx)

	hint := "↑↓: select agent | Enter: apply filter | Esc: cancel"
	if m.activeTab == panel.TabSkills {
		hint = "↑↓: select agent (local skills dir only) | Enter: apply | Esc: cancel"
	}
	m.setFooterContext(hint)
	return m, nil
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

	m.Agent.SetSize(innerWidth, listHeight)
	subtitle := "Filter skills and MCP by agent"
	if m.activeTab == panel.TabSkills {
		subtitle = "Agents with a local skills directory"
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.PanelTitle.Render("Agent filter"),
		m.styles.Hint.Render(subtitle),
		m.Agent.View(),
	)
	return m.styles.Modal.Width(dialogWidth).Render(body)
}

func (m *Model) renderAgentFilterDialogArea() string {
	leftWidth, mainHeight, _, _ := m.paneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, m.renderAgentFilterDialog())
}
