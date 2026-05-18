package mcp

import (
	"os"
	"path/filepath"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type extraConfigProbe struct {
	path   func(projectRoot, home string) string
	scope  extension.Scope
	agents []string
}

// Probes for MCP configs that do not live under a single agent directory walk root.
var extraConfigProbes = []extraConfigProbe{
	{
		path:   func(projectRoot, _ string) string { return filepath.Join(projectRoot, ".mcp.json") },
		scope:  extension.ScopeProject,
		agents: []string{"claude-code"},
	},
	{
		path:   func(_, home string) string { return filepath.Join(home, ".claude.json") },
		scope:  extension.ScopeGlobal,
		agents: []string{"claude-code"},
	},
}

func scanExtraConfigPaths(projectRoot, home string) ([]*mcpdomain.Server, error) {
	var servers []*mcpdomain.Server
	seen := map[string]bool{}

	for _, probe := range extraConfigProbes {
		configPath := probe.path(projectRoot, home)
		if configPath == "" || seen[configPath] {
			continue
		}
		if _, err := os.Stat(configPath); err != nil {
			continue
		}
		seen[configPath] = true

		var parsed []*mcpdomain.Server
		var err error
		if filepath.Base(configPath) == ".claude.json" {
			parsed, err = ParseClaudeJSON(configPath, projectRoot, home)
		} else {
			parsed, err = ParseConfigAtPathForAgents(configPath, projectRoot, home, probe.scope, probe.agents)
		}
		if err != nil {
			return nil, err
		}
		servers = append(servers, parsed...)
	}

	return servers, nil
}
