package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestParseConfigFileExpandsServers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "mcp.json")
	const body = `{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    },
    "remote": {
      "url": "https://example.com/mcp"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := ParseConfigFile(configPath, dir, "", extension.ScopeProject)
	if err != nil {
		t.Fatalf("ParseConfigFile: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	names := map[string]bool{}
	for _, srv := range servers {
		names[srv.GetName()] = true
		if srv.ConfigPath != configPath {
			t.Fatalf("unexpected config path: %s", srv.ConfigPath)
		}
	}
	if !names["server-filesystem"] || !names["example"] {
		t.Fatalf("unexpected server names: %v", names)
	}
	for _, srv := range servers {
		if srv.ConfigKey == "filesystem" && srv.GetName() != "server-filesystem" {
			t.Fatalf("filesystem entry name: %s", srv.GetName())
		}
	}
}
