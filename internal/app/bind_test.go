package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

func TestMcpTargetBound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	codexPath := filepath.Join(home, ".codex", "config.toml")

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:   "filesystem",
			Scope:  extension.ScopeGlobal,
			Agents: []string{"cursor"},
		},
		ConfigKey: "filesystem",
	}

	target := servicemcp.BindTarget{
		Agent:      agent.Agent{ID: "cursor"},
		Scope:      extension.ScopeGlobal,
		ConfigPath: cursorPath,
	}

	t.Run("empty bindings falls back to server agents", func(t *testing.T) {
		if !mcpTargetBound(srv, target) {
			t.Fatal("expected bound via server-level agents")
		}
	})

	t.Run("agent id in binding", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				Scope:     extension.ScopeGlobal,
				ConfigKey: "filesystem",
				Agents:    []string{"codex"},
			}},
		}
		codexTarget := servicemcp.BindTarget{
			Agent: agent.Agent{ID: "codex"},
			Scope: extension.ScopeGlobal,
		}
		if !mcpTargetBound(s, codexTarget) {
			t.Fatal("expected bound via binding agents list")
		}
	})

	t.Run("path clean normalizes slashes", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: cursorPath + "/",
				Scope:      extension.ScopeGlobal,
				ConfigKey:  "filesystem",
			}},
		}
		if !mcpTargetBound(s, target) {
			t.Fatal("expected bound via cleaned config path")
		}
	})

	t.Run("empty bindings slice must not use stale aggregated agents", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			BaseExtension: extension.BaseExtension{
				ConfigPath: cursorPath,
				Scope:      extension.ScopeProject,
				Agents:     []string{"cursor", "codex", "windsurf"},
			},
			Bindings: []mcpdomain.Binding{},
		}
		windsurfTarget := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "windsurf"},
			Scope:      extension.ScopeProject,
			ConfigPath: filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"),
		}
		if mcpTargetBound(s, windsurfTarget) {
			t.Fatal("empty Bindings with stale srv.Agents must not mark windsurf bound")
		}
	})

	t.Run("path match requires agent when binding lists agents", func(t *testing.T) {
		windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: windsurfPath,
				ConfigKey:  "filesystem",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"codex"},
			}},
		}
		target := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "windsurf"},
			Scope:      extension.ScopeGlobal,
			ConfigPath: windsurfPath,
		}
		if mcpTargetBound(s, target) {
			t.Fatal("path match alone must not bind windsurf when binding agents omit windsurf")
		}
	})

	t.Run("different path not bound", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: codexPath,
				Scope:      extension.ScopeGlobal,
				ConfigKey:  "filesystem",
			}},
		}
		if mcpTargetBound(s, target) {
			t.Fatal("expected not bound for unrelated config path")
		}
	})
}

func TestMCPBindChoicesOneRowPerTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:   "server-filesystem",
			Agents: []string{"cursor"},
		},
		ConfigKey: "filesystem",
	}
	choices := newMCPBindChoices(srv, root, home)
	targets := servicemcp.ListBindTargets(root, home)
	if len(choices) != len(targets) {
		t.Fatalf("expected %d MCP bind targets, got %d choices", len(targets), len(choices))
	}
	if len(choices) < 4 {
		t.Fatalf("expected at least 4 MCP bind targets (multi-scope), got %d", len(choices))
	}
	seen := map[string]bool{}
	for _, c := range choices {
		key := c.agent.ID + "|" + string(c.scope)
		if seen[key] {
			t.Fatalf("duplicate target: %s", key)
		}
		seen[key] = true
	}
}

func TestApplyMCPBindChoicesMultipleAgents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeMCPJSON(t, cursorPath, "filesystem", "npx", []string{"-y", "pkg"})

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: cursorPath,
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "pkg"},
	}

	choices := newMCPBindChoices(srv, root, home)
	for i := range choices {
		switch {
		case choices[i].agent.ID == "codex" && choices[i].scope == extension.ScopeGlobal:
			choices[i].desired = true
		case choices[i].agent.ID == "windsurf" && choices[i].scope == extension.ScopeGlobal:
			choices[i].desired = true
		}
	}

	mgr := servicemcp.NewManager()
	if err := applyMCPBindChoices(mgr, srv, choices, root, home); err != nil {
		t.Fatalf("apply: %v", err)
	}

	codexPath := filepath.Join(home, ".codex", "config.toml")
	if _, err := os.Stat(codexPath); err != nil {
		t.Fatalf("codex config: %v", err)
	}
	windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	if _, err := os.Stat(windsurfPath); err != nil {
		t.Fatalf("windsurf config: %v", err)
	}
}

func TestBindChoicesTogglePreservesOthers(t *testing.T) {
	t.Parallel()

	choices := []agentBindChoice{
		{agent: agent.Agent{Name: "Cursor", ID: "cursor"}, scope: extension.ScopeGlobal, initial: true, desired: true},
		{agent: agent.Agent{Name: "Codex", ID: "codex"}, scope: extension.ScopeGlobal, initial: false, desired: false},
	}
	choices[1].desired = true

	if !choices[0].desired || !choices[1].desired {
		t.Fatal("expected both desired after toggling codex")
	}
}

func writeMCPJSON(t *testing.T, path, key, command string, args []string) {
	t.Helper()
	content := `{"mcpServers":{"` + key + `":{"command":"` + command + `","args":[`
	for i, a := range args {
		if i > 0 {
			content += ","
		}
		content += `"` + a + `"`
	}
	content += `]}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
