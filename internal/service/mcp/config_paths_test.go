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
		}
		if tgt.Agent.ID == "cursor" && tgt.Scope == "project" {
			cursorProject = true
		}
	}
	if cursorGlobal {
		t.Fatalf("cursor should not have a global scope bind target")
	}
	if !cursorProject {
		t.Fatalf("cursor missing scopes: project=%v", cursorProject)
	}
}
