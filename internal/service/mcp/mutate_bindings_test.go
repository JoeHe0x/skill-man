package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestToggleDisableAllBindings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.json")
	pathB := filepath.Join(dir, "b.json")
	writeJSON(t, pathA, mcpConfig("fs", "echo", []string{"a"}))
	writeJSON(t, pathB, mcpConfig("fs", "echo", []string{"b"}))

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{Name: "server-filesystem"},
		Bindings: []mcpdomain.Binding{
			{ConfigPath: pathA, ConfigKey: "fs", Command: "echo", Args: []string{"a"}},
			{ConfigPath: pathB, ConfigKey: "fs", Command: "echo", Args: []string{"b"}},
		},
	}

	mgr := NewManager()
	if err := mgr.ToggleDisable(srv); err != nil {
		t.Fatalf("ToggleDisable: %v", err)
	}
	assertServerDisabled(t, pathA, "fs", true)
	assertServerDisabled(t, pathB, "fs", true)

	if err := mgr.ToggleDisable(srv); err != nil {
		t.Fatalf("ToggleDisable enable: %v", err)
	}
	assertServerDisabled(t, pathA, "fs", false)
	assertServerDisabled(t, pathB, "fs", false)
}

func TestRemoveAllBindings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.json")
	pathB := filepath.Join(dir, "b.json")
	writeJSON(t, pathA, mcpConfig("fs", "echo", []string{"a"}))
	writeJSON(t, pathB, mcpConfig("fs", "echo", []string{"b"}))

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{Name: "server-filesystem"},
		Bindings: []mcpdomain.Binding{
			{ConfigPath: pathA, ConfigKey: "fs"},
			{ConfigPath: pathB, ConfigKey: "fs"},
		},
	}

	if err := NewManager().Remove(srv); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	assertEmptyMCPServers(t, pathA)
	assertEmptyMCPServers(t, pathB)
}

func mcpConfig(key, command string, args []string) map[string]any {
	return map[string]any{
		"mcpServers": map[string]any{
			key: map[string]any{"command": command, "args": args},
		},
	}
}

func assertServerDisabled(t *testing.T, path, key string, want bool) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		t.Fatalf("parse: %v", err)
	}
	servers := root["mcpServers"].(map[string]any)
	entry := servers[key].(map[string]any)
	disabled, _ := entry["disabled"].(bool)
	if disabled != want {
		t.Fatalf("%s disabled=%v, want %v", path, disabled, want)
	}
}

func assertEmptyMCPServers(t *testing.T, path string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		t.Fatalf("parse: %v", err)
	}
	servers := root["mcpServers"].(map[string]any)
	if len(servers) != 0 {
		t.Fatalf("expected empty mcpServers in %s, got %v", path, servers)
	}
}
