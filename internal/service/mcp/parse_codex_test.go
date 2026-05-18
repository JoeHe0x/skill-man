package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestParseCodexConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	const body = `
[mcp_servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]

[mcp_servers.disabled_srv]
command = "echo"
enabled = false

[mcp_servers.remote]
url = "https://example.com/mcp"
`
	if err := os.WriteFile(configPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := ParseCodexConfigFile(configPath, dir, "", extension.ScopeProject, nil)
	if err != nil {
		t.Fatalf("ParseCodexConfigFile: %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("expected 3 servers (including disabled), got %d", len(servers))
	}
	for _, srv := range servers {
		if srv.GetName() == "disabled_srv" && !srv.IsDisabled() {
			t.Fatal("expected disabled_srv to be marked disabled")
		}
		if len(srv.GetAgents()) != 1 || srv.GetAgents()[0] != "codex" {
			t.Fatalf("expected codex agent, got %v", srv.GetAgents())
		}
	}
}
