package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// agentBindChoice tracks desired vs initial bind state for one agent in the bind UI.
type agentBindChoice struct {
	agent   agent.Agent
	initial bool
	desired bool
}

func newMCPBindChoices(srv *mcpdomain.Server) []agentBindChoice {
	choices := make([]agentBindChoice, 0, len(servicemcp.AgentsWithMCPDir()))
	for _, a := range servicemcp.AgentsWithMCPDir() {
		bound := slices.Contains(srv.GetAgents(), a.ID)
		choices = append(choices, agentBindChoice{
			agent:   a,
			initial: bound,
			desired: bound,
		})
	}
	return choices
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

func bindChoicesToListItems(choices []agentBindChoice) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		items = append(items, listItem{
			kind:  itemKindMessage,
			title: bindAgentTitle(c.agent.Name, c.desired),
			desc:  bindAgentDesc(c.agent),
			meta:  c.agent.ID,
		})
	}
	return items
}

func applyMCPBindChoices(mgr *servicemcp.Manager, srv *mcpdomain.Server, choices []agentBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		var err error
		switch {
		case c.desired && !c.initial:
			err = mgr.Bind(srv, c.agent, projectRoot, home)
		case !c.desired && c.initial:
			err = mgr.Unbind(srv, c.agent, projectRoot, home)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.agent.Name, err))
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
