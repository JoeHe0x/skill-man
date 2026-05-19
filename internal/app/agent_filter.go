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
	items = append(items, listItem{
		kind:  itemKindMessage,
		title: filterAgentTitle("All agents", currentID == "all"),
		desc:  "Show skills and MCP for every agent",
		meta:  "all",
	})
	for _, a := range agents {
		active := strings.EqualFold(a.ID, currentID)
		items = append(items, listItem{
			kind:  itemKindMessage,
			title: filterAgentTitle(a.Name, active),
			desc:  bindAgentDesc(a),
			meta:  a.ID,
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
	m.lastState = m.state
	m.state = stateFilteringAgent

	current := currentAgentFilterID(m.agentIDs)
	visible := m.agentsForFilterDialog()
	items := newAgentFilterListItems(visible, current)
	m.listDelegate.SetHeight(listHeightForItems(items))
	m.agentList.SetItems(items)

	selIdx := 0
	for i, item := range items {
		li, ok := item.(listItem)
		if ok && strings.EqualFold(li.meta, current) {
			selIdx = i
			break
		}
	}
	m.agentList.Select(selIdx)

	hint := "↑↓: select agent | Enter: apply filter | Esc: cancel"
	if m.activeTab == panel.TabSkills {
		hint = "↑↓: select agent (local skills dir only) | Enter: apply | Esc: cancel"
	}
	m.hint = hint
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
		selected, ok := m.agentList.SelectedItem().(listItem)
		if !ok || selected.meta == "" {
			return m, nil
		}
		m.setAgentFilter(selected.meta)
		m.state = m.lastState
		m.hint = fmt.Sprintf("Agent filter: %s", m.agentDisplay())
		m.refreshActiveList()
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		m.state = m.lastState
		m.hint = "Agent filter cancelled"
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) renderAgentFilterDialog() string {
	leftWidth, _, _, _ := m.paneSizes()

	// Determine dynamic width based on item contents
	maxWidth := len("Agent filter")
	subtitle := "Filter skills and MCP by agent"
	if m.activeTab == panel.TabSkills {
		subtitle = "Agents with a local skills directory"
	}
	if len(subtitle) > maxWidth {
		maxWidth = len(subtitle)
	}

	for _, item := range m.agentList.Items() {
		if li, ok := item.(listItem); ok {
			// Title and desc are rendered on one line with "  " between them
			itemLen := len("  ") + len(li.title) + len("  ") + len(li.desc)
			if itemLen > maxWidth {
				maxWidth = itemLen
			}
		}
	}

	dialogWidth := maxWidth + 8 // 4 for inner padding, 4 for list padding
	if dialogWidth > leftWidth-2 {
		dialogWidth = leftWidth - 2
	}
	if dialogWidth < 20 {
		dialogWidth = 20
	}

	// Dynamic height based on items
	numItems := len(m.agentList.Items())
	listHeight := numItems + 2 // padding

	dialogHeight := listHeight + 8 // dialog frame padding
	dialogHeight = min(max(10, dialogHeight), min(m.height-8, 28))
	listHeight = dialogHeight - 8
	if listHeight < 2 {
		listHeight = 2
	}

	innerWidth := dialogWidth - 4
	m.agentList.SetSize(innerWidth, listHeight)

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
