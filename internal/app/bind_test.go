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

func TestMCPBindChoicesOnePerAgent(t *testing.T) {
	t.Parallel()

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:   "server-filesystem",
			Agents: []string{"cursor"},
		},
	}
	choices := newMCPBindChoices(srv)
	if len(choices) < 2 {
		t.Fatalf("expected at least 2 MCP-capable agents, got %d", len(choices))
	}
	seen := map[string]bool{}
	for _, c := range choices {
		if seen[c.agent.ID] {
			t.Fatalf("duplicate agent: %s", c.agent.ID)
		}
		seen[c.agent.ID] = true
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

	choices := newMCPBindChoices(srv)
	for i := range choices {
		if choices[i].agent.ID == "codex" || choices[i].agent.ID == "windsurf" {
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
		{agent: agent.Agent{Name: "Cursor", ID: "cursor"}, initial: true, desired: true},
		{agent: agent.Agent{Name: "Codex", ID: "codex"}, initial: false, desired: false},
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
