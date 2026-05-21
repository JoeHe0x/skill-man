package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type bindFeature struct {
	m       *Model
	session bindSession
}

func (f *bindFeature) Name() string  { return "bind" }
func (f *bindFeature) Active() bool  { return f.m.state == stateBindingAgent }
func (f *bindFeature) Init() tea.Cmd { return nil }
func (f *bindFeature) View(width, height int) string {
	return ""
}

func (f *bindFeature) Clear() {
	f.session.clear()
}

func (f *bindFeature) startFromItem(item panel.Item) (tea.Model, tea.Cmd) {
	if !f.m.activePanel().Capabilities().Bind || !item.CanBind() {
		if !f.m.activePanel().Capabilities().Bind {
			f.m.setFooterContext("Bind is not available for this tab")
		} else {
			f.m.setFooterContext("Select an item first to manage agent bindings")
		}
		return f.m, nil
	}
	eff, ok := item.BindEffect()
	if !ok {
		f.m.setFooterContext("Select a skill or MCP server first")
		return f.m, nil
	}
	if !f.m.transitionTo(stateBindingAgent) {
		return f.m, nil
	}

	if eff.Skill != nil {
		f.session.mcp = nil
		f.session.skill = eff.Skill
		f.session.agents = newSkillBindChoices(eff.Skill, f.m.cwd, f.m.home)
		f.m.setAgentListItems(bindChoicesToListItems(f.session.agents, f.m.cwd, f.m.home))
		f.m.agentList.Select(0)
		return f.m, nil
	}

	f.session.skill = nil
	key := eff.MCPKey
	members := f.m.mcpMembersForConfigKey(key)
	if len(members) == 0 {
		members = append([]*mcpdomain.Server(nil), eff.MCPMembers...)
	}
	f.session.mcpMembers = members
	f.session.mcp = mcpBindTemplate(f.session.mcpMembers)
	f.session.agents = newMCPBindChoices(f.session.mcpMembers, f.m.cwd, f.m.home)
	f.m.setAgentListItems(bindChoicesToListItems(f.session.agents, f.m.cwd, f.m.home))
	f.m.agentList.Select(0)
	f.m.setFooterContext(fmt.Sprintf("Bind key `%s` · space: toggle · enter: apply", key))
	return f.m, nil
}

func (f *bindFeature) syncHint() {
	selected, total := 0, len(f.session.agents)
	for _, c := range f.session.agents {
		if c.desired {
			selected++
		}
	}
	changes := 0
	for _, c := range f.session.agents {
		if c.desired != c.initial {
			changes++
		}
	}
	f.m.setFooterContext(fmt.Sprintf("Bind · %d/%d selected · %d pending change(s)", selected, total, changes))
}

func (f *bindFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.handleKeys(msg)
		return cmd, true
	}
	return nil, false
}

func (f *bindFeature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if f.session.mcp != nil {
			srv := f.session.mcp
			if err := applyMCPBindChoices(f.m.mcpManager, srv, f.session.agents, f.m.cwd, f.m.home); err != nil {
				f.m.reportError(err)
			}
			f.Clear()
			f.m.transitionTo(stateListing)
			var cmds []tea.Cmd
			if f.m.errMsg == "" {
				key := mcpConfigKeyFromMembers(f.session.mcpMembers)
				if key == "" {
					key = srv.GetName()
				}
				cmds = append(cmds, f.m.flashFooter(fmt.Sprintf("Updated MCP bindings for %s", key)))
			}
			cmds = append(cmds, tea.Sequence(
				f.m.beginScanAllCmd(),
				func() tea.Msg {
					key := srv.ConfigKey
					if key == "" {
						key = srv.GetName()
					}
					return reselectMCPMsg{name: key}
				},
			))
			return f.m, tea.Batch(cmds...)
		}
		if f.session.skill != nil {
			skill := f.session.skill
			if err := applySkillBindChoices(context.Background(), f.m.skillManager, skill, f.session.agents, f.m.cwd, f.m.home); err != nil {
				f.m.reportError(err)
			}
			f.Clear()
			f.m.transitionTo(stateListing)
			var cmds []tea.Cmd
			if f.m.errMsg == "" {
				cmds = append(cmds, f.m.flashFooter(fmt.Sprintf("Updated agent bindings for %s", skill.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				f.m.beginScanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: skill.GetName()} },
			))
			return f.m, tea.Batch(cmds...)
		}
		f.m.transitionTo(stateListing)
		return f.m, nil

	case key.Matches(msg, keys.Cancel):
		f.Clear()
		f.m.transitionTo(stateListing)
		return f.m, f.m.flashFooter("Agent binding cancelled")

	case key.Matches(msg, keys.Toggle):
		idx := f.m.agentList.Index()
		if idx < 0 || idx >= len(f.session.agents) {
			return f.m, nil
		}
		f.session.agents[idx].desired = !f.session.agents[idx].desired
		f.m.setAgentListItems(bindChoicesToListItems(f.session.agents, f.m.cwd, f.m.home))
		f.m.agentList.Select(idx)
		f.syncHint()
		return f.m, nil
	}

	var cmd tea.Cmd
	f.m.agentList, cmd = f.m.agentList.Update(msg)
	return f.m, cmd
}
