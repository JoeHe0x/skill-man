package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

func TestApplyMCPBindChoicesAllScopes(t *testing.T) {
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

	mgr := servicemcp.NewManager()
	b := usecasebind.NewBinder(nil, mgr, root, home)
	choices := b.NewMCPChoices([]*mcpdomain.Server{srv})
	for i := range choices {
		choices[i].Desired = true
	}

	if err := b.ApplyMCP(srv, choices); err != nil {
		t.Fatalf("apply all targets: %v", err)
	}

	for _, tgt := range servicemcp.ListBindTargets(root, home) {
		if _, err := os.Stat(tgt.ConfigPath); err != nil {
			t.Fatalf("missing config for %s %s at %s: %v", tgt.Agent.ID, tgt.Scope, tgt.ConfigPath, err)
		}
	}
}
