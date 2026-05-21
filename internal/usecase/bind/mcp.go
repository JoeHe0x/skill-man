package bind

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// MCPConfigKey returns the MCP config key for a member group.
func MCPConfigKey(members []*mcpdomain.Server) string {
	if len(members) == 0 {
		return ""
	}
	if k := members[0].ConfigKey; k != "" {
		return k
	}
	return members[0].GetName()
}

// MCPBindTemplate picks a member to copy command/args/url from when writing new bindings.
func MCPBindTemplate(members []*mcpdomain.Server) *mcpdomain.Server {
	for _, srv := range members {
		if srv.Command != "" || srv.URL != "" {
			cp := *srv
			cp.ConfigKey = MCPConfigKey(members)
			return &cp
		}
	}
	if len(members) == 0 {
		return nil
	}
	cp := *members[0]
	cp.ConfigKey = MCPConfigKey(members)
	return &cp
}

// NewMCPChoices builds bind rows for every MCP bind target.
func (b Binder) NewMCPChoices(members []*mcpdomain.Server) []Choice {
	targets := servicemcp.ListBindTargets(b.CWD, b.Home)
	choices := make([]Choice, 0, len(targets))
	for _, t := range targets {
		bound := mcpTargetBoundFromMembers(members, t)
		choices = append(choices, Choice{
			Agent:      t.Agent,
			Scope:      t.Scope,
			ConfigPath: t.ConfigPath,
			Initial:    bound,
			Desired:    bound,
		})
	}
	return choices
}

// ApplyMCP applies pending MCP bind/unbind changes.
func (b Binder) ApplyMCP(srv *mcpdomain.Server, choices []Choice) error {
	var errs []error
	for _, c := range choices {
		if c.Scope == "" || c.ConfigPath == "" {
			continue
		}
		label := mcpRowTitle(c.Agent.Name, c.Scope) + " → " + servicemcp.ShortPath(b.Home, c.ConfigPath)
		target := servicemcp.BindTarget{Agent: c.Agent, Scope: c.Scope, ConfigPath: c.ConfigPath}
		var err error
		switch {
		case c.Desired && !c.Initial:
			err = b.MCP.BindAtTarget(srv, target, b.CWD, b.Home)
		case !c.Desired && c.Initial:
			err = b.MCP.UnbindAtTarget(srv, target, b.CWD, b.Home)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", label, err))
		}
	}
	return errors.Join(errs...)
}

func mcpTargetBoundFromMembers(members []*mcpdomain.Server, t servicemcp.BindTarget) bool {
	key := MCPConfigKey(members)
	for _, srv := range members {
		if !servicemcp.ConfigPathsEqual(srv.ConfigPath, t.ConfigPath) {
			continue
		}
		srvKey := srv.ConfigKey
		if srvKey == "" {
			srvKey = srv.GetName()
		}
		if key != "" && srvKey != key {
			continue
		}
		return true
	}
	return false
}

// MCPTargetBound reports whether srv is bound at target (used by tests and choice building).
func MCPTargetBound(srv *mcpdomain.Server, t servicemcp.BindTarget) bool {
	key := srv.ConfigKey
	if key == "" {
		key = srv.GetName()
	}

	if srv.Bindings != nil && len(srv.Bindings) == 0 {
		return false
	}

	bindings := srv.AllBindings()
	if len(bindings) == 0 {
		return servicemcp.ConfigPathsEqual(srv.ConfigPath, t.ConfigPath) &&
			slices.Contains(srv.Agents, t.Agent.ID)
	}

	targetPath := filepath.Clean(t.ConfigPath)
	for _, binding := range bindings {
		if !servicemcp.ConfigPathsEqual(binding.ConfigPath, targetPath) || binding.ConfigKey != key {
			continue
		}
		if len(binding.Agents) == 0 {
			return true
		}
		if slices.Contains(binding.Agents, t.Agent.ID) {
			return true
		}
	}
	return false
}

func mcpRowTitle(name string, scope extension.Scope) string {
	return fmt.Sprintf("%s (%s)", name, scope)
}
