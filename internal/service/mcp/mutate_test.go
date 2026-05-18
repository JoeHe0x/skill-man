package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestToggleAndRemoveJSONServer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	writeJSON(t, path, map[string]any{
		"mcpServers": map[string]any{
			"demo": map[string]any{"command": "echo", "args": []string{"hi"}},
		},
	})

	mgr := NewManager()
	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "demo",
			ConfigPath: path,
			Path:       dir,
			Scope:      extension.ScopeProject,
		},
		ConfigKey: "demo",
		Command:   "echo",
		Args:      []string{"hi"},
	}

	if err := mgr.ToggleDisable(srv); err != nil {
		t.Fatalf("ToggleDisable: %v", err)
	}
	srv.Disabled = true
	if err := mgr.ToggleDisable(srv); err != nil {
		t.Fatalf("ToggleDisable enable: %v", err)
	}

	if err := mgr.Remove(srv); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	body, _ := os.ReadFile(path)
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		t.Fatalf("parse: %v", err)
	}
	servers := root["mcpServers"].(map[string]any)
	if len(servers) != 0 {
		t.Fatalf("expected empty mcpServers, got %v", servers)
	}
}

func TestBindJSONServer(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeJSON(t, source, map[string]any{
		"mcpServers": map[string]any{
			"shared": map[string]any{"command": "npx", "args": []string{"-y", "pkg"}},
		},
	})

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "shared",
			ConfigPath: source,
			Path:       filepath.Dir(source),
			Scope:      extension.ScopeProject,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "shared",
		Command:   "npx",
		Args:      []string{"-y", "pkg"},
	}

	target, _ := agent.AgentByID("claude-code")
	if err := NewManager().Bind(srv, target, root, ""); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	dest := filepath.Join(root, ".mcp.json")
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if !json.Valid(body) {
		t.Fatal("invalid json written")
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
