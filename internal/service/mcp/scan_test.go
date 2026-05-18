package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func TestScanFindsProjectMCPConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cursorDir := filepath.Join(root, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	config := `{"mcpServers":{"fs":{"command":"echo","args":["hello"]}}}`
	if err := os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := Scan(context.Background(), root, "", []agent.Agent{{
		ID:         "cursor",
		EntityDirs: map[agent.EntityType]string{agent.EntityMCP: ".cursor"},
	}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].GetName() != "fs" {
		t.Fatalf("unexpected server name: %s", servers[0].GetName())
	}
}

func TestScanClaudeCodeProjectMCP(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	config := `{"mcpServers":{"claude-tool":{"command":"echo","args":["hi"]}}}`
	if err := os.WriteFile(filepath.Join(root, ".mcp.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := Scan(context.Background(), root, "", nil)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].GetName() != "claude-tool" {
		t.Fatalf("unexpected server name: %s", servers[0].GetName())
	}
	if len(servers[0].GetAgents()) != 1 || servers[0].GetAgents()[0] != "claude-code" {
		t.Fatalf("expected claude-code agent, got %v", servers[0].GetAgents())
	}
}

func TestScanWindsurfMCPConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	windsurfDir := filepath.Join(root, ".codeium", "windsurf")
	if err := os.MkdirAll(windsurfDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	config := `{"mcpServers":{"github":{"serverUrl":"https://example.com/mcp"}}}`
	if err := os.WriteFile(filepath.Join(windsurfDir, "mcp_config.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := Scan(context.Background(), root, "", []agent.Agent{{
		ID:         "windsurf",
		EntityDirs: map[agent.EntityType]string{agent.EntityMCP: ".codeium/windsurf"},
	}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].URL != "https://example.com/mcp" {
		t.Fatalf("unexpected url: %s", servers[0].URL)
	}
}

func TestScanEmptyCodexConfigTomlDoesNotPanic(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("# no mcp_servers\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := Scan(context.Background(), root, "", []agent.Agent{{
		ID:         "codex",
		EntityDirs: map[agent.EntityType]string{agent.EntityMCP: ".codex"},
	}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("expected no servers, got %d", len(servers))
	}
}

func TestScanCodexConfigToml(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	config := `
[mcp_servers.fs]
command = "echo"
args = ["hello"]
`
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := Scan(context.Background(), root, "", []agent.Agent{{
		ID:         "codex",
		EntityDirs: map[agent.EntityType]string{agent.EntityMCP: ".codex"},
	}})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].GetName() != "fs" {
		t.Fatalf("unexpected server name: %s", servers[0].GetName())
	}
}
