package mcp

import (
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestListBindTargetsDedupesSharedConfigPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")

	var windsurfCount int
	for _, tgt := range ListBindTargets(root, home) {
		if filepath.Clean(tgt.ConfigPath) == filepath.Clean(windsurfPath) {
			windsurfCount++
		}
	}
	if windsurfCount != 1 {
		t.Fatalf("expected 1 bind target for windsurf config, got %d", windsurfCount)
	}
}

func TestListBindTargetsIncludesCursorGlobalPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorGlobal := filepath.Join(home, ".cursor", "mcp.json")

	var found bool
	for _, tgt := range ListBindTargets(root, home) {
		if tgt.Agent.ID == "cursor" && filepath.Clean(tgt.ConfigPath) == filepath.Clean(cursorGlobal) {
			found = true
			if tgt.Scope != extension.ScopeGlobal {
				t.Fatalf("cursor global scope = %s", tgt.Scope)
			}
		}
	}
	if !found {
		t.Fatal("expected cursor global ~/.cursor/mcp.json bind target")
	}
}

func TestListBindTargetsMatchesDiscoverConfigLocations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	if len(ListBindTargets(root, home)) != len(discoverConfigLocations(root, home)) {
		t.Fatal("ListBindTargets and discoverConfigLocations should return the same paths")
	}
}

func TestListBindTargetsKeepsDistinctCodexPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	codexProject := filepath.Join(root, ".codex", "config.toml")
	codexGlobal := filepath.Join(home, ".codex", "config.toml")

	var projectOK, globalOK bool
	for _, tgt := range ListBindTargets(root, home) {
		if tgt.Agent.ID != "codex" {
			continue
		}
		switch filepath.Clean(tgt.ConfigPath) {
		case filepath.Clean(codexProject):
			projectOK = true
		case filepath.Clean(codexGlobal):
			globalOK = true
		}
	}
	if !projectOK || !globalOK {
		t.Fatalf("expected separate codex project/global targets, project=%v global=%v", projectOK, globalOK)
	}
}
