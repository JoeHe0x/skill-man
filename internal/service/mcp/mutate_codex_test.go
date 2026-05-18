package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestBindCodexStdioOmitsURL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:   "server-filesystem",
			Scope:  extension.ScopeGlobal,
			Agents: []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
	}

	mgr := NewManager()
	codex, _ := agent.AgentByID("codex")
	if err := mgr.Bind(srv, codex, root, home); err != nil {
		t.Fatalf("Bind codex: %v", err)
	}

	codexPath := filepath.Join(home, ".codex", "config.toml")
	body, err := os.ReadFile(codexPath)
	if err != nil {
		t.Fatalf("read codex config: %v", err)
	}
	text := string(body)
	if strings.Contains(text, "url") {
		t.Fatalf("stdio bind must not write url to codex config.toml:\n%s", text)
	}
	if !strings.Contains(text, "command") || !strings.Contains(text, "filesystem") {
		t.Fatalf("expected stdio filesystem entry:\n%s", text)
	}
}

func TestSanitizeCodexServerRemovesURLForStdio(t *testing.T) {
	t.Parallel()

	sc := sanitizeCodexServer(codexServerConfig{
		Command: "npx",
		Args:    []string{"-y", "pkg"},
		URL:     "",
	})
	if sc.URL != "" {
		t.Fatalf("expected empty url, got %q", sc.URL)
	}

	sc = sanitizeCodexServer(codexServerConfig{
		Command: "npx",
		Args:    []string{"-y", "pkg"},
		URL:     "https://example.com/mcp",
	})
	if sc.URL != "" {
		t.Fatalf("stdio must drop url, got %q", sc.URL)
	}
	if sc.Command == "" {
		t.Fatal("expected command preserved")
	}
}

func TestRepairCodexConfigFileStripsStdioURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	const broken = `
[mcp_servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
url = ""
`
	if err := os.WriteFile(configPath, []byte(broken), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	repaired, err := RepairCodexConfigFile(configPath)
	if err != nil {
		t.Fatalf("RepairCodexConfigFile: %v", err)
	}
	if !repaired {
		t.Fatal("expected repair to rewrite config")
	}
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(body), "url") {
		t.Fatalf("repaired config must not contain url for stdio:\n%s", body)
	}
}

func TestToggleCodexRepairsStaleURLOnStdio(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	const broken = `
[mcp_servers.filesystem]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
url = ""
`
	if err := os.WriteFile(configPath, []byte(broken), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			ConfigPath: configPath,
			Scope:      extension.ScopeProject,
			Disabled:   true,
		},
		ConfigKey: "filesystem",
		Command:   "npx",
	}

	if err := toggleCodexServer(srv); err != nil {
		t.Fatalf("toggleCodexServer: %v", err)
	}
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(body), "url") {
		t.Fatalf("toggle should strip stale url from stdio server:\n%s", body)
	}
}
