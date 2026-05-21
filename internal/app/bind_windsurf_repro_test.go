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
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
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
	srv := findScannedServer(t, servers, cursorProjectPath, "filesystem")
	t.Logf("scanned server agents=%v bindings=%d path=%s", srv.GetAgents(), srv.BindingCount(), srv.ConfigPath)

	mgr := servicemcp.NewManager()
	b := usecasebind.NewBinder(nil, mgr, root, home)
	choices := b.NewMCPChoices([]*mcpdomain.Server{srv})
	for i := range choices {
		choices[i].Desired = false
	}
	for i, c := range choices {
		switch {
		case c.Agent.ID == "codex" && c.Scope == extension.ScopeGlobal:
			choices[i].Desired = true
		case c.Agent.ID == "cursor" && c.Scope == extension.ScopeProject:
			choices[i].Desired = true
		}
		t.Logf("choice %s %s initial=%v desired=%v", c.Agent.ID, c.Scope, c.Initial, choices[i].Desired)
	}

	if err := b.ApplyMCP(srv, choices); err != nil {
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

	b := usecasebind.NewBinder(nil, servicemcp.NewManager(), root, home)
	for _, c := range b.NewMCPChoices([]*mcpdomain.Server{srv}) {
		if c.Agent.ID == "windsurf" && c.Initial {
			t.Fatalf("windsurf %s should not be initially bound", c.Scope)
		}
	}
}

func findScannedServer(t *testing.T, servers []*mcpdomain.Server, configPath, key string) *mcpdomain.Server {
	t.Helper()
	for _, s := range servers {
		if s.ConfigPath == configPath && s.ConfigKey == key {
			return s
		}
	}
	t.Fatalf("no server at %s key %q among %d scanned", configPath, key, len(servers))
	return nil
}
