package filter

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	featbind "github.com/JoeHe0x/skill-man/internal/app/feature/bind"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func currentAgentFilterID(agentIDs []string) string {
	if len(agentIDs) == 0 {
		return "all"
	}
	return agentIDs[0]
}

func agentsForFilterDialog(h Host) []agent.Agent {
	if h.ActiveTab() == panel.TabSkills {
		return agent.AgentsWithLocalSkillDir(h.AllAgents(), h.CWD(), h.Home())
	}
	return h.AllAgents()
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

// Open enters agent-filter mode and populates the overlay list.
func Open(h Host) (tea.Model, tea.Cmd) {
	h.TransitionTo(session.FilteringAgent)

	current := currentAgentFilterID(h.AgentIDs())
	visible := agentsForFilterDialog(h)
	items := newAgentFilterListItems(visible, current)
	h.SetAgentListItems(items)

	selIdx := 0
	for i, item := range items {
		li, ok := item.(panel.Item)
		if ok && strings.EqualFold(li.Meta, current) {
			selIdx = i
			break
		}
	}
	h.AgentListSelect(selIdx)

	hint := "↑↓: select agent | Enter: apply filter | Esc: cancel"
	if h.ActiveTab() == panel.TabSkills {
		hint = "↑↓: select agent (local skills dir only) | Enter: apply | Esc: cancel"
	}
	h.SetFooterContext(hint)
	return h.TeaModel(), nil
}

// RenderDialog builds the modal body (tests and overlay).
func RenderDialog(h Host) string {
	leftWidth, _, _, _ := h.PaneSizes()
	dialogWidth := min(max(44, leftWidth-4), 56)
	if dialogWidth > leftWidth-2 {
		dialogWidth = max(20, leftWidth-2)
	}
	dialogHeight := min(max(16, h.Height()-8), 28)
	innerWidth := dialogWidth - 4
	listHeight := dialogHeight - 8
	if listHeight < 4 {
		listHeight = 4
	}

	h.AgentListSetSize(innerWidth, listHeight)
	subtitle := "Filter skills and MCP by agent"
	if h.ActiveTab() == panel.TabSkills {
		subtitle = "Agents with a local skills directory"
	}
	styles := h.Styles()
	body := lipgloss.JoinVertical(lipgloss.Left,
		styles.PanelTitle.Render("Agent filter"),
		styles.Hint.Render(subtitle),
		h.AgentListView(),
	)
	return styles.Modal.Width(dialogWidth).Render(body)
}

// RenderMainOverlay places the dialog in the main pane.
func RenderMainOverlay(h Host) string {
	leftWidth, mainHeight, _, _ := h.PaneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, RenderDialog(h))
}
