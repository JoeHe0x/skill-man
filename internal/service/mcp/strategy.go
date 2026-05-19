package mcp

import (
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

// ScanStrategy discovers MCP config files under agent-specific directories.
type ScanStrategy struct{}

func (s ScanStrategy) DefaultDir() string {
	return ".cursor"
}

func (s ScanStrategy) AgentDir(a agent.Agent) string {
	return agent.MCPEntityDir(a)
}

func (s ScanStrategy) SkipDir(dirName string) bool {
	return dirName == ".git" || dirName == "node_modules"
}

func (s ScanStrategy) TargetFiles() []string {
	return []string{"mcp.json", "mcp.json.disabled", "mcp_config.json", "config.toml"}
}

func (s ScanStrategy) ParseFile(filePath, projectRoot, home string, scope extension.Scope) (*mcpdomain.Server, error) {
	servers, err := ParseConfigAtPath(filePath, projectRoot, home, scope)
	if err != nil || len(servers) == 0 {
		return nil, err
	}
	return servers[0], nil
}

func (s ScanStrategy) Dedupe(servers []*mcpdomain.Server) []*mcpdomain.Server {
	// Backend: one record per config file (no merge-by-display-name).
	// The TUI groups rows visually in panel.mcpListItems; bind/toggle/remove use each Server as-is.
	return dedupeByConfigLocation(servers)
}

func mergeAgentIDs(a, b []string) []string {
	set := map[string]bool{}
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		set[id] = true
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}
