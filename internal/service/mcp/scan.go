package mcp

import (
	"context"
	"os"

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

	// Parse any bind/scan candidate path not found by directory walk (keeps list and bind aligned).
	for _, loc := range ListBindTargets(projectRoot, home) {
		if seen[loc.ConfigPath] {
			continue
		}
		if _, err := os.Stat(loc.ConfigPath); err != nil {
			continue
		}
		seen[loc.ConfigPath] = true
		agentIDs := []string{loc.Agent.ID}
		parsed, err := ParseConfigAtPathForAgents(loc.ConfigPath, projectRoot, home, loc.Scope, agentIDs)
		if err != nil {
			return nil, err
		}
		servers = append(servers, parsed...)
	}

	return strategy.Dedupe(servers), nil
}
