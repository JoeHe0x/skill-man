package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

type bindFeature struct {
	host    bindHost
	session bindSession
}

func (f *bindFeature) Name() string  { return "bind" }
func (f *bindFeature) Active() bool  { return f.host.IsBinding() }
func (f *bindFeature) Init() tea.Cmd { return nil }
func (f *bindFeature) View(width, height int) string {
	return ""
}

func (f *bindFeature) Clear() {
	f.session.clear()
}

func (f *bindFeature) startFromItem(item panel.Item) (tea.Model, tea.Cmd) {
	if !f.host.ActivePanelCanBind() || !item.CanBind() {
		if !f.host.ActivePanelCanBind() {
			f.host.SetFooterContext("Bind is not available for this tab")
		} else {
			f.host.SetFooterContext("Select an item first to manage agent bindings")
		}
		return f.host.TeaModel(), nil
	}
	eff, ok := item.BindEffect()
	if !ok {
		f.host.SetFooterContext("Select a skill or MCP server first")
		return f.host.TeaModel(), nil
	}
	if !f.host.TransitionTo(stateBindingAgent) {
		return f.host.TeaModel(), nil
	}

	b := f.host.Binder()
	cwd, home := f.host.CWD(), f.host.Home()

	if eff.Skill != nil {
		f.session.mcp = nil
		f.session.skill = eff.Skill
		f.session.agents = b.NewSkillChoices(eff.Skill)
		f.host.SetAgentListItems(bindChoicesToListItems(f.session.agents, cwd, home))
		f.host.AgentListSelect(0)
		return f.host.TeaModel(), nil
	}

	f.session.skill = nil
	key := eff.MCPKey
	members := f.host.MCPMembersForConfigKey(key)
	if len(members) == 0 {
		members = append([]*mcpdomain.Server(nil), eff.MCPMembers...)
	}
	f.session.mcpMembers = members
	f.session.mcp = usecasebind.MCPBindTemplate(f.session.mcpMembers)
	f.session.agents = b.NewMCPChoices(f.session.mcpMembers)
	f.host.SetAgentListItems(bindChoicesToListItems(f.session.agents, cwd, home))
	f.host.AgentListSelect(0)
	f.host.SetFooterContext(fmt.Sprintf("Bind key `%s` · space: toggle · enter: apply", key))
	return f.host.TeaModel(), nil
}

func (f *bindFeature) syncHint() {
	selected, total := 0, len(f.session.agents)
	for _, c := range f.session.agents {
		if c.Desired {
			selected++
		}
	}
	changes := 0
	for _, c := range f.session.agents {
		if c.Desired != c.Initial {
			changes++
		}
	}
	f.host.SetFooterContext(fmt.Sprintf("Bind · %d/%d selected · %d pending change(s)", selected, total, changes))
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
	m := f.host.TeaModel()
	switch {
	case key.Matches(msg, keys.Enter):
		if f.session.mcp != nil {
			srv := f.session.mcp
			if err := f.host.Binder().ApplyMCP(srv, f.session.agents); err != nil {
				f.host.ReportError(err)
			}
			f.Clear()
			f.host.TransitionTo(stateListing)
			var cmds []tea.Cmd
			if f.host.ErrMsg() == "" {
				key := usecasebind.MCPConfigKey(f.session.mcpMembers)
				if key == "" {
					key = srv.GetName()
				}
				cmds = append(cmds, f.host.FlashFooter(fmt.Sprintf("Updated MCP bindings for %s", key)))
			}
			cmds = append(cmds, tea.Sequence(
				f.host.BeginScanAllCmd(),
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
		if f.session.skill != nil {
			skill := f.session.skill
			if err := f.host.Binder().ApplySkill(context.Background(), skill, f.session.agents); err != nil {
				f.host.ReportError(err)
			}
			f.Clear()
			f.host.TransitionTo(stateListing)
			var cmds []tea.Cmd
			if f.host.ErrMsg() == "" {
				cmds = append(cmds, f.host.FlashFooter(fmt.Sprintf("Updated agent bindings for %s", skill.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				f.host.BeginScanAllCmd(),
				func() tea.Msg { return reselectSkillMsg{name: skill.GetName()} },
			))
			return m, tea.Batch(cmds...)
		}
		f.host.TransitionTo(stateListing)
		return m, nil

	case key.Matches(msg, keys.Cancel):
		f.Clear()
		f.host.TransitionTo(stateListing)
		return m, f.host.FlashFooter("Agent binding cancelled")

	case key.Matches(msg, keys.Toggle):
		idx := f.host.AgentListIndex()
		if idx < 0 || idx >= len(f.session.agents) {
			return m, nil
		}
		f.session.agents[idx].Desired = !f.session.agents[idx].Desired
		f.host.SetAgentListItems(bindChoicesToListItems(f.session.agents, f.host.CWD(), f.host.Home()))
		f.host.AgentListSelect(idx)
		f.syncHint()
		return m, nil
	}

	_, cmd := f.host.AgentListUpdate(msg)
	return m, cmd
}
