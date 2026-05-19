package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

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

type skillBindChoice struct {
	dir     string
	agents  []agent.Agent
	initial bool
	desired bool
}

func newSkillBindChoices(skill *skilldomain.Skill) []skillBindChoice {
	// Group agents by their skill directory
	groups := make(map[string][]agent.Agent)
	for _, a := range agent.DefaultAgents() {
		dir := a.EntityDirs[agent.EntitySkill]
		if dir == "" {
			continue
		}
		groups[dir] = append(groups[dir], a)
	}

	// Sort directories for deterministic order
	var dirs []string
	for dir := range groups {
		dirs = append(dirs, dir)
	}
	slices.Sort(dirs)

	choices := make([]skillBindChoice, 0, len(dirs))
	for _, dir := range dirs {
		groupAgents := groups[dir]

		// Consider it initially bound if ANY agent in the group is bound
		bound := false
		for _, a := range groupAgents {
			if slices.Contains(skill.GetAgents(), a.ID) {
				bound = true
				break
			}
		}

		choices = append(choices, skillBindChoice{
			dir:     dir,
			agents:  groupAgents,
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

func skillBindChoicesToListItems(choices []skillBindChoice) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {

		// If multiple agents share this path, list them in parenthesis
		var agentNames []string
		for _, a := range c.agents {
			agentNames = append(agentNames, a.Name)
		}
		names := strings.Join(agentNames, ", ")

		title := bindAgentTitle(c.dir, c.desired)
		desc := names

		items = append(items, listItem{
			kind:  itemKindMessage,
			title: title,
			desc:  desc,
			meta:  c.dir, // use dir as meta for filtering/identification if needed
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

func applySkillBindChoices(ctx context.Context, mgr manager.ExtensionManager[*skilldomain.Skill], skill *skilldomain.Skill, choices []skillBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		var err error
		switch {
		case c.desired && !c.initial:
			// Bind to all agents in this group
			for _, a := range c.agents {
				if e := mgr.Bind(ctx, skill, a, projectRoot, home); e != nil {
					err = errors.Join(err, fmt.Errorf("%s: %w", a.Name, e))
				}
			}
		case !c.desired && c.initial:
			// Unbind from all agents in this group
			for _, a := range c.agents {
				if e := mgr.Unbind(ctx, skill, a, projectRoot, home); e != nil {
					err = errors.Join(err, fmt.Errorf("%s: %w", a.Name, e))
				}
			}
		}
		if err != nil {
			errs = append(errs, err)
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
