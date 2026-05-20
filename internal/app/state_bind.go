package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func (m *Model) handleBindSelected() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().Bind {
		m.setFooterContext("Bind is not available for this tab")
		return m, nil
	}
	selected, ok := m.list.SelectedItem().(actionable)
	if !ok || !selected.CanBind() {
		m.setFooterContext("Select an item first to manage agent bindings")
		return m, nil
	}

	m.transitionTo(stateBindingAgent)

	target := selected.BindTarget()
	switch target.Kind {
	case "mcp":
		m.binds.skill = nil
		key := target.MCPKey
		members := m.mcpMembersForConfigKey(key)
		if len(members) == 0 {
			members = append([]*mcpdomain.Server(nil), target.MCPMembers...)
		}
		m.binds.mcpMembers = members
		m.binds.mcp = mcpBindTemplate(m.binds.mcpMembers)
		m.binds.agents = newMCPBindChoices(m.binds.mcpMembers, m.cwd, m.home)
		m.setAgentListItems(bindChoicesToListItems(m.binds.agents, m.cwd, m.home))
		m.agentList.Select(0)
		m.setFooterContext(fmt.Sprintf("Bind key `%s` · space: toggle · enter: apply", key))
		return m, nil

	case "skill":
		m.binds.mcp = nil
		m.binds.skill = target.Skill
		m.binds.agents = newSkillBindChoices(target.Skill, m.cwd, m.home)
		m.setAgentListItems(bindChoicesToListItems(m.binds.agents, m.cwd, m.home))
		m.agentList.Select(0)
		return m, nil
	}

	m.transitionTo(stateListing)
	return m, m.flashFooter("Select a skill or MCP server first")
}

func (m *Model) handleBindingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.binds.mcp != nil {
			srv := m.binds.mcp
			if err := applyMCPBindChoices(m.mcpManager, srv, m.binds.agents, m.cwd, m.home); err != nil {
				m.reportError(err)
			}
			m.clearBindingSession()
			m.transitionTo(stateListing)
			var cmds []tea.Cmd
			if m.errMsg == "" {
				key := mcpConfigKeyFromMembers(m.binds.mcpMembers)
				if key == "" {
					key = srv.GetName()
				}
				cmds = append(cmds, m.flashFooter(fmt.Sprintf("Updated MCP bindings for %s", key)))
			}
			cmds = append(cmds, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg {
					key := srv.ConfigKey
					if key == "" {
						key = srv.GetName()
					}
					return reselectMCPMsg{name: key}
				},
			))
			return m, tea.Batch(cmds...)
		}
		if m.binds.skill != nil {
			skill := m.binds.skill
			if err := applySkillBindChoices(context.Background(), m.skillManager, skill, m.binds.agents, m.cwd, m.home); err != nil {
				m.reportError(err)
			}
			m.clearBindingSession()
			m.transitionTo(stateListing)
			var cmds []tea.Cmd
			if m.errMsg == "" {
				cmds = append(cmds, m.flashFooter(fmt.Sprintf("Updated agent bindings for %s", skill.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				m.scanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: skill.GetName()} },
			))
			return m, tea.Batch(cmds...)
		}
		m.transitionTo(stateListing)
		return m, nil

	case key.Matches(msg, keys.Cancel):
		m.clearBindingSession()
		m.transitionTo(stateListing)
		return m, m.flashFooter("Agent binding cancelled")

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		if idx < 0 || idx >= len(m.binds.agents) {
			return m, nil
		}
		m.binds.agents[idx].desired = !m.binds.agents[idx].desired
		m.setAgentListItems(bindChoicesToListItems(m.binds.agents, m.cwd, m.home))
		m.agentList.Select(idx)
		m.syncBindHint()
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}
