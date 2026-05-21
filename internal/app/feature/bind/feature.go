package bind

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

// Feature owns the agent bind overlay.
type Feature struct {
	host Host
	snap bindSession
}

// New returns a bind feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string  { return "bind" }
func (f *Feature) Active() bool  { return f.host.IsBinding() }
func (f *Feature) Init() tea.Cmd { return nil }
func (f *Feature) View(width, height int) string {
	return ""
}

func (f *Feature) Clear() {
	f.snap.clear()
}

func (f *Feature) StartFromItem(item panel.Item) (tea.Model, tea.Cmd) {
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
	if !f.host.TransitionTo(session.BindingAgent) {
		return f.host.TeaModel(), nil
	}

	b := f.host.Binder()
	cwd, home := f.host.CWD(), f.host.Home()

	if eff.Skill != nil {
		f.snap.mcp = nil
		f.snap.skill = eff.Skill
		f.snap.agents = b.NewSkillChoices(eff.Skill)
		f.host.SetAgentListItems(ChoicesToListItems(f.snap.agents, cwd, home))
		f.host.AgentListSelect(0)
		return f.host.TeaModel(), nil
	}

	f.snap.skill = nil
	key := eff.MCPKey
	members := f.host.MCPMembersForConfigKey(key)
	if len(members) == 0 {
		members = append([]*mcpdomain.Server(nil), eff.MCPMembers...)
	}
	f.snap.mcpMembers = members
	f.snap.mcp = usecasebind.MCPBindTemplate(f.snap.mcpMembers)
	f.snap.agents = b.NewMCPChoices(f.snap.mcpMembers)
	f.host.SetAgentListItems(ChoicesToListItems(f.snap.agents, cwd, home))
	f.host.AgentListSelect(0)
	f.host.SetFooterContext(fmt.Sprintf("Bind key `%s` · space: toggle · enter: apply", key))
	return f.host.TeaModel(), nil
}

func (f *Feature) SyncHint() {
	selected, total := 0, len(f.snap.agents)
	for _, c := range f.snap.agents {
		if c.Desired {
			selected++
		}
	}
	changes := 0
	for _, c := range f.snap.agents {
		if c.Desired != c.Initial {
			changes++
		}
	}
	f.host.SetFooterContext(fmt.Sprintf("Bind · %d/%d selected · %d pending change(s)", selected, total, changes))
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
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

func (f *Feature) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := f.host.TeaModel()
	keys := uikeys.Default
	switch {
	case key.Matches(msg, keys.Enter):
		if f.snap.mcp != nil {
			srv := f.snap.mcp
			if err := f.host.Binder().ApplyMCP(srv, f.snap.agents); err != nil {
				f.host.ReportError(err)
			}
			f.Clear()
			f.host.TransitionTo(session.Listing)
			var cmds []tea.Cmd
			if f.host.ErrMsg() == "" {
				key := usecasebind.MCPConfigKey(f.snap.mcpMembers)
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
					return uimsg.ReselectMCP{Name: key}
				},
			))
			return m, tea.Batch(cmds...)
		}
		if f.snap.skill != nil {
			skill := f.snap.skill
			if err := f.host.Binder().ApplySkill(context.Background(), skill, f.snap.agents); err != nil {
				f.host.ReportError(err)
			}
			f.Clear()
			f.host.TransitionTo(session.Listing)
			var cmds []tea.Cmd
			if f.host.ErrMsg() == "" {
				cmds = append(cmds, f.host.FlashFooter(fmt.Sprintf("Updated agent bindings for %s", skill.GetName())))
			}
			cmds = append(cmds, tea.Sequence(
				f.host.BeginScanAllCmd(),
				func() tea.Msg { return uimsg.ReselectSkill{Name: skill.GetName()} },
			))
			return m, tea.Batch(cmds...)
		}
		f.host.TransitionTo(session.Listing)
		return m, nil

	case key.Matches(msg, keys.Cancel):
		f.Clear()
		f.host.TransitionTo(session.Listing)
		return m, f.host.FlashFooter("Agent binding cancelled")

	case key.Matches(msg, keys.Toggle):
		idx := f.host.AgentListIndex()
		if idx < 0 || idx >= len(f.snap.agents) {
			return m, nil
		}
		f.snap.agents[idx].Desired = !f.snap.agents[idx].Desired
		f.host.SetAgentListItems(ChoicesToListItems(f.snap.agents, f.host.CWD(), f.host.Home()))
		f.host.AgentListSelect(idx)
		f.SyncHint()
		return m, nil
	}

	_, cmd := f.host.AgentListUpdate(msg)
	return m, cmd
}

// SessionAgents exposes agents for tests in the app package.
func (f *Feature) SessionAgents() []usecasebind.Choice {
	return f.snap.agents
}
