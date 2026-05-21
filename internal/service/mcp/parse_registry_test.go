package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestParseConfigAtPath_usesRegistryForCodexToml(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[mcp_servers]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseConfigAtPath(path, dir, "", extension.ScopeProject)
	if err != nil {
		t.Fatalf("ParseConfigAtPath config.toml: %v", err)
	}
}

func TestParseConfigAtPath_usesRegistryForClaudeJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseConfigAtPath(path, dir, "", extension.ScopeProject)
	if err != nil {
		t.Fatalf("ParseConfigAtPath .claude.json: %v", err)
	}
}
