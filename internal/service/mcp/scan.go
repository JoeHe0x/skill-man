package mcp

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
)

// Scan discovers MCP servers from mcp.json files under known agent config directories.
func Scan(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]*mcpdomain.Server, error) {
	strategy := ScanStrategy{}
	placeholders, err := manager.ScanExtensions(ctx, projectRoot, home, agents, strategy)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var servers []*mcpdomain.Server
	for _, ph := range placeholders {
		if seen[ph.ConfigPath] {
			continue
		}
		seen[ph.ConfigPath] = true
		expanded, err := ParseConfigAtPath(ph.ConfigPath, projectRoot, home, ph.Scope)
		if err != nil {
			return nil, err
		}
		servers = append(servers, expanded...)
	}

	extra, err := scanExtraConfigPaths(projectRoot, home)
	if err != nil {
		return nil, err
	}
	servers = append(servers, extra...)

	return strategy.Dedupe(servers), nil
}
