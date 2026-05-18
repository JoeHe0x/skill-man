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

func TestBindMultipleAgents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorPath), 0o755); err != nil {
		t.Fatalf("mkdir cursor: %v", err)
	}
	writeJSON(t, cursorPath, map[string]any{
		"mcpServers": map[string]any{
			"filesystem": map[string]any{
				"command": "npx",
				"args":    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
		},
	})

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: cursorPath,
			Path:       filepath.Dir(cursorPath),
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		Bindings: []mcpdomain.Binding{{
			ConfigPath: cursorPath,
			ConfigKey:  "filesystem",
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
			Command:    "npx",
			Args:       []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		}},
	}

	mgr := NewManager()
	codex, _ := agent.AgentByID("codex")
	claude, _ := agent.AgentByID("claude-code")
	if err := mgr.Bind(srv, codex, root, home); err != nil {
		t.Fatalf("Bind codex: %v", err)
	}
	if err := mgr.Bind(srv, claude, root, home); err != nil {
		t.Fatalf("Bind claude: %v", err)
	}

	codexPath := filepath.Join(home, ".codex", "config.toml")
	if _, err := os.Stat(codexPath); err != nil {
		t.Fatalf("codex config missing: %v", err)
	}
	claudePath := filepath.Join(root, ".mcp.json")
	body, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read claude config: %v", err)
	}
	var rootJSON map[string]any
	if err := json.Unmarshal(body, &rootJSON); err != nil {
		t.Fatalf("parse claude config: %v", err)
	}
	servers, _ := rootJSON["mcpServers"].(map[string]any)
	if servers["filesystem"] == nil {
		t.Fatalf("expected filesystem entry in %s", claudePath)
	}
}
