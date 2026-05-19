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

// TargetConfigPath returns the MCP config file path for an agent at the given scope.
func TargetConfigPath(a agent.Agent, scope extension.Scope, projectRoot, home string) string {
	return targetConfigPath(a, scope, projectRoot, home)
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
	case "claude-code":
		if scope == extension.ScopeProject {
			return filepath.Join(projectRoot, ".mcp.json")
		}
		return filepath.Join(home, ".claude.json")
	case "cursor":
		if scope == extension.ScopeGlobal {
			return "" // Cursor global MCP settings are managed in UI (SQLite/settings.json), not via ~/.cursor/mcp.json
		}
		return filepath.Join(base, mcpDir, "mcp.json")
	default:
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

// BindTarget is one bindable MCP config destination (agent + scope + resolved path).
type BindTarget struct {
	Agent      agent.Agent
	Scope      extension.Scope
	ConfigPath string
}

// ListBindTargets returns every agent/scope pair that has a writable MCP config path.
func ListBindTargets(projectRoot, home string) []BindTarget {
	var out []BindTarget
	for _, a := range AgentsWithMCPDir() {
		for _, scope := range []extension.Scope{extension.ScopeProject, extension.ScopeGlobal} {
			path := targetConfigPath(a, scope, projectRoot, home)
			if path == "" {
				continue
			}
			out = append(out, BindTarget{Agent: a, Scope: scope, ConfigPath: path})
		}
	}
	return out
}
