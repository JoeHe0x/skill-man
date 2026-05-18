package mcp

import (
	"path/filepath"
	"testing"
)

func TestListBindTargetsIncludesMultipleScopes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	targets := ListBindTargets(root, home)

	if len(targets) < 6 {
		t.Fatalf("expected at least 6 bind targets (4 agents × project/global), got %d", len(targets))
	}

	var cursorGlobal, cursorProject bool
	for _, tgt := range targets {
		if tgt.Agent.ID == "cursor" && tgt.Scope == "global" {
			cursorGlobal = true
			want := filepath.Join(home, ".cursor", "mcp.json")
			if tgt.ConfigPath != want {
				t.Fatalf("cursor global path = %q, want %q", tgt.ConfigPath, want)
			}
		}
		if tgt.Agent.ID == "cursor" && tgt.Scope == "project" {
			cursorProject = true
		}
	}
	if !cursorGlobal || !cursorProject {
		t.Fatalf("cursor missing scopes: global=%v project=%v", cursorGlobal, cursorProject)
	}
}
