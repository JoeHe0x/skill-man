package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanExtraClaudeJSON(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	project := filepath.Join(home, "workspace")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	config := `{
  "mcpServers":{"global":{"command":"node","args":["server.js"]}},
  "projects":{"` + filepath.ToSlash(project) + `":{"mcpServers":{"local":{"command":"node","args":["local.js"]}}}}
}`
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := scanExtraConfigPaths(project, home)
	if err != nil {
		t.Fatalf("scanExtraConfigPaths: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].GetAgents()[0] != "claude-code" {
		t.Fatalf("expected claude-code agent, got %v", servers[0].GetAgents())
	}
}
