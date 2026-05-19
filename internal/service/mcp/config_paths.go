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

// ConfigPathsEqual reports whether two paths refer to the same file.
func ConfigPathsEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if a == b {
		return true
	}
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return false
	}
	return filepath.Clean(ra) == filepath.Clean(rb)
}

// targetConfigPath returns the primary MCP config file path for an agent at the given scope.
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

// ListBindTargets returns every MCP config file path that scan and bind both use.
// Multiple rows can share an agent when that agent has more than one config file.
func ListBindTargets(projectRoot, home string) []BindTarget {
	var out []BindTarget
	seen := map[string]bool{}
	for _, loc := range discoverConfigLocations(projectRoot, home) {
		clean := filepath.Clean(loc.ConfigPath)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, loc)
	}
	return out
}

// discoverConfigLocations lists all MCP config file paths the scanner may read.
// Keep in sync with ScanStrategy.TargetFiles and extraConfigProbes.
func discoverConfigLocations(projectRoot, home string) []BindTarget {
	var out []BindTarget

	for _, a := range AgentsWithMCPDir() {
		switch a.ID {
		case "codex":
			addLocation(&out, a, extension.ScopeProject, filepath.Join(projectRoot, ".codex", "config.toml"))
			addLocation(&out, a, extension.ScopeGlobal, filepath.Join(home, ".codex", "config.toml"))
		case "cursor":
			addLocation(&out, a, extension.ScopeProject, filepath.Join(projectRoot, ".cursor", "mcp.json"))
			addLocation(&out, a, extension.ScopeGlobal, filepath.Join(home, ".cursor", "mcp.json"))
		case "windsurf":
			if home != "" {
				addLocation(&out, a, extension.ScopeGlobal, filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"))
			}
		case "claude-code":
			addLocation(&out, a, extension.ScopeProject, filepath.Join(projectRoot, ".mcp.json"))
			addLocation(&out, a, extension.ScopeProject, filepath.Join(projectRoot, ".claude", "mcp.json"))
			addLocation(&out, a, extension.ScopeGlobal, filepath.Join(home, ".claude.json"))
			if home != "" {
				addLocation(&out, a, extension.ScopeGlobal, filepath.Join(home, ".claude", "mcp.json"))
			}
		default:
			for _, scope := range []extension.Scope{extension.ScopeProject, extension.ScopeGlobal} {
				addLocation(&out, a, scope, targetConfigPath(a, scope, projectRoot, home))
			}
		}
	}
	return out
}

func addLocation(out *[]BindTarget, a agent.Agent, scope extension.Scope, path string) {
	if path == "" {
		return
	}
	*out = append(*out, BindTarget{Agent: a, Scope: scope, ConfigPath: path})
}
