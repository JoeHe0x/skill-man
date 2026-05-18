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
