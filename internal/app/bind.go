package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// agentBindChoice tracks desired vs initial bind state for one agent in the bind UI.
type agentBindChoice struct {
	agent   agent.Agent
	scope   extension.Scope // MCP only; empty for skills
	initial bool
	desired bool
}

func newMCPBindChoices(srv *mcpdomain.Server, projectRoot, home string) []agentBindChoice {
	targets := servicemcp.ListBindTargets(projectRoot, home)
	choices := make([]agentBindChoice, 0, len(targets))
	for _, t := range targets {
		bound := mcpTargetBound(srv, t)
		choices = append(choices, agentBindChoice{
			agent:   t.Agent,
			scope:   t.Scope,
			initial: bound,
			desired: bound,
		})
	}
	return choices
}

func mcpTargetBound(srv *mcpdomain.Server, t servicemcp.BindTarget) bool {
	key := srv.ConfigKey
	if key == "" {
		key = srv.GetName()
	}
	for _, b := range srv.AllBindings() {
		if b.ConfigPath != t.ConfigPath || b.Scope != t.Scope || b.ConfigKey != key {
			continue
		}
		return true
	}
	return false
}

func newSkillBindChoices(skill *skilldomain.Skill) []agentBindChoice {
	var agents []agent.Agent
	for _, a := range agent.DefaultAgents() {
		if a.EntityDirs[agent.EntitySkill] == "" {
			continue
		}
		agents = append(agents, a)
	}
	choices := make([]agentBindChoice, 0, len(agents))
	for _, a := range agents {
		bound := slices.Contains(skill.GetAgents(), a.ID)
		choices = append(choices, agentBindChoice{
			agent:   a,
			initial: bound,
			desired: bound,
		})
	}
	return choices
}

func bindChoicesToListItems(choices []agentBindChoice, projectRoot, home string) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		title := bindAgentTitle(c.agent.Name, c.desired)
		desc := bindAgentDesc(c.agent)
		if c.scope != "" {
			title = bindAgentTitle(mcpBindRowTitle(c.agent.Name, c.scope), c.desired)
			desc = mcpBindRowDesc(c.agent, c.scope, projectRoot, home)
		}
		items = append(items, listItem{
			kind:  itemKindMessage,
			title: title,
			desc:  desc,
			meta:  c.agent.ID,
		})
	}
	return items
}

func mcpBindRowTitle(name string, scope extension.Scope) string {
	return fmt.Sprintf("%s (%s)", name, scope)
}

func mcpBindRowDesc(a agent.Agent, scope extension.Scope, projectRoot, home string) string {
	path := servicemcp.TargetConfigPath(a, scope, projectRoot, home)
	return servicemcp.ShortPath(home, path)
}

func applyMCPBindChoices(mgr *servicemcp.Manager, srv *mcpdomain.Server, choices []agentBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		if c.scope == "" {
			continue
		}
		label := mcpBindRowTitle(c.agent.Name, c.scope)
		var err error
		switch {
		case c.desired && !c.initial:
			err = mgr.BindAt(srv, c.agent, c.scope, projectRoot, home)
		case !c.desired && c.initial:
			err = mgr.UnbindAt(srv, c.agent, c.scope, projectRoot, home)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", label, err))
		}
	}
	return errors.Join(errs...)
}

func applySkillBindChoices(ctx context.Context, mgr manager.ExtensionManager[*skilldomain.Skill], skill *skilldomain.Skill, choices []agentBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		var err error
		switch {
		case c.desired && !c.initial:
			err = mgr.Bind(ctx, skill, c.agent, projectRoot, home)
		case !c.desired && c.initial:
			err = mgr.Unbind(ctx, skill, c.agent, projectRoot, home)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.agent.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Model) clearBindingSession() {
	m.bindingSkill = nil
	m.bindingMCP = nil
	m.bindingAgents = nil
}

func bindAgentTitle(name string, checked bool) string {
	if checked {
		return "✓ " + name
	}
	return "  " + name
}

func bindAgentDesc(a agent.Agent) string {
	if dir := agent.MCPEntityDir(a); dir != "" {
		return dir
	}
	return a.EntityDirs[agent.EntitySkill]
}
