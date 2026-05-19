package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

var defaultScanAgents = agent.DefaultAgents()

// Reproduces: user binds only codex + cursor; windsurf must not get a config entry.
func TestBindCodexAndCursorDoesNotBindWindsurf(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorProjectPath := filepath.Join(root, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorProjectPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeMCPJSON(t, cursorProjectPath, "filesystem", "npx", []string{"-y", "@modelcontextprotocol/server-filesystem", root})

	servers, err := servicemcp.Scan(t.Context(), root, home, defaultScanAgents)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(servers) == 0 {
		t.Fatal("expected scanned MCP server")
	}
	srv := servers[0]
	t.Logf("scanned server agents=%v bindings=%d", srv.GetAgents(), srv.BindingCount())

	choices := newMCPBindChoices(srv, root, home)
	for i := range choices {
		choices[i].desired = false
	}
	for i, c := range choices {
		switch {
		case c.agent.ID == "codex" && c.scope == extension.ScopeGlobal:
			choices[i].desired = true
		case c.agent.ID == "cursor" && c.scope == extension.ScopeProject:
			choices[i].desired = true
		}
		t.Logf("choice %s %s initial=%v desired=%v", c.agent.ID, c.scope, c.initial, choices[i].desired)
	}

	mgr := servicemcp.NewManager()
	if err := applyMCPBindChoices(mgr, srv, choices, root, home); err != nil {
		t.Fatalf("apply: %v", err)
	}

	windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	body, err := os.ReadFile(windsurfPath)
	if err == nil && strings.Contains(string(body), "filesystem") {
		t.Fatalf("windsurf config should not contain server after binding only codex+cursor:\n%s", body)
	}
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read windsurf config: %v", err)
	}

	servers2, err := servicemcp.Scan(t.Context(), root, home, defaultScanAgents)
	if err != nil {
		t.Fatalf("rescan: %v", err)
	}
	for _, s := range servers2 {
		for _, id := range s.GetAgents() {
			if id == "windsurf" {
				t.Fatalf("after rescan, server %q lists windsurf in agents %v bindings=%d",
					s.GetName(), s.GetAgents(), s.BindingCount())
			}
		}
	}
}

// When server only has cursor binding, windsurf rows must start unchecked.
func TestMCPBindChoicesWindsurfInitiallyUnbound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorProjectPath := filepath.Join(root, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorProjectPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeMCPJSON(t, cursorProjectPath, "filesystem", "npx", []string{"-y", "pkg"})

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: cursorProjectPath,
			Scope:      extension.ScopeProject,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "pkg"},
		Bindings: []mcpdomain.Binding{{
			ConfigPath: cursorProjectPath,
			ConfigKey:  "filesystem",
			Scope:      extension.ScopeProject,
			Agents:     []string{"cursor"},
			Command:    "npx",
			Args:       []string{"-y", "pkg"},
		}},
	}

	for _, c := range newMCPBindChoices(srv, root, home) {
		if c.agent.ID == "windsurf" && c.initial {
			t.Fatalf("windsurf %s should not be initially bound", c.scope)
		}
	}
}
