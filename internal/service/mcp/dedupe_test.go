package mcp

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestScanStrategyDedupeKeepsSeparateConfigFiles(t *testing.T) {
	t.Parallel()

	servers := []*mcpdomain.Server{
		makeTestMCPServer("server-filesystem", "/home/joe/.cursor/mcp.json", "filesystem", extension.ScopeGlobal, []string{"cursor"}, "/home/joe"),
		makeTestMCPServer("server-filesystem", "/home/joe/.codex/config.toml", "filesystem", extension.ScopeGlobal, []string{"codex"}, "/mnt/c/Code/skill-man"),
		makeTestMCPServer("server-filesystem", "/mnt/c/Code/skill-man/.cursor/mcp.json", "filesystem", extension.ScopeProject, []string{"cursor"}, "/mnt/c/Code/skill-man"),
	}

	out := ScanStrategy{}.Dedupe(servers)
	if len(out) != 3 {
		t.Fatalf("expected 3 separate list entries (one per config file), got %d", len(out))
	}
}

func makeTestMCPServer(name, configPath, key string, scope extension.Scope, agents []string, root string) *mcpdomain.Server {
	args := []string{"-y", "@modelcontextprotocol/server-filesystem", root}
	return &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       name,
			ConfigPath: configPath,
			Scope:      scope,
			Agents:     agents,
		},
		ConfigKey: key,
		Command:   "npx",
		Args:      args,
		Bindings: []mcpdomain.Binding{{
			ConfigPath: configPath,
			ConfigKey:  key,
			Scope:      scope,
			Agents:     agents,
			Command:    "npx",
			Args:       args,
		}},
	}
}
