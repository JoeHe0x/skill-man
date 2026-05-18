package mcp

import (
	"path/filepath"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// configFormat identifies how an MCP config file is encoded.
type configFormat int

const (
	formatJSON configFormat = iota
	formatTOML
)

func configFormatForPath(path string) configFormat {
	switch filepath.Base(path) {
	case "config.toml":
		return formatTOML
	default:
		return formatJSON
	}
}

// targetConfigPath returns the MCP config file path for an agent at the given scope.
func targetConfigPath(a agent.Agent, scope extension.Scope, projectRoot, home string) string {
	mcpDir := agent.MCPEntityDir(a)
	if mcpDir == "" {
		return ""
	}

	base := projectRoot
	if scope == extension.ScopeGlobal {
		base = home
	}
	if base == "" {
		return ""
	}

	switch a.ID {
	case "codex":
		return filepath.Join(base, mcpDir, "config.toml")
	case "windsurf":
		if home == "" {
			return ""
		}
		return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	default:
		if a.ID == "claude-code" && scope == extension.ScopeProject {
			return filepath.Join(projectRoot, ".mcp.json")
		}
		return filepath.Join(base, mcpDir, "mcp.json")
	}
}

// AgentsWithMCPDir returns agents that declare an MCP config directory.
func AgentsWithMCPDir() []agent.Agent {
	var out []agent.Agent
	for _, a := range agent.DefaultAgents() {
		if agent.MCPEntityDir(a) != "" {
			out = append(out, a)
		}
	}
	return out
}
