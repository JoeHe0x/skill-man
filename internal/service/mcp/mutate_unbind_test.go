package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestUnbindAbsentAgentIsNoOp(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeJSON(t, cursorPath, mcpConfig("filesystem", "npx", []string{"-y", "pkg"}))

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

	claude, _ := agent.AgentByID("claude-code")
	if err := NewManager().Unbind(srv, claude, root, home); err != nil {
		t.Fatalf("Unbind absent agent: %v", err)
	}
}

func TestBindMultipleAgentsFromProjectSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	projectCursor := filepath.Join(root, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(projectCursor), 0o755); err != nil {
		t.Fatalf("mkdir project cursor: %v", err)
	}
	writeJSON(t, projectCursor, mcpConfig("filesystem", "npx", []string{"-y", "pkg", root}))

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: projectCursor,
			Scope:      extension.ScopeProject,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "pkg", root},
	}

	mgr := NewManager()
	codex, _ := agent.AgentByID("codex")
	windsurf, _ := agent.AgentByID("windsurf")
	if err := mgr.Bind(srv, codex, root, home); err != nil {
		t.Fatalf("Bind codex: %v", err)
	}
	if err := mgr.Bind(srv, windsurf, root, home); err != nil {
		t.Fatalf("Bind windsurf: %v", err)
	}

	codexPath := filepath.Join(home, ".codex", "config.toml")
	if _, err := os.Stat(codexPath); err != nil {
		t.Fatalf("expected global codex config at %s: %v", codexPath, err)
	}
}
