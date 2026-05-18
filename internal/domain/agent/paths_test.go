package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasLocalSkillDir(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	claudeDir := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	claude, ok := AgentByID("claude-code")
	if !ok {
		t.Fatal("claude-code agent missing")
	}
	if !HasLocalSkillDir(claude, root, home) {
		t.Fatal("expected .claude/skills under project root")
	}

	cursor, ok := AgentByID("cursor")
	if !ok {
		t.Fatal("cursor agent missing")
	}
	if HasLocalSkillDir(cursor, root, home) {
		t.Fatal("cursor uses .agents/skills which was not created")
	}
}

func TestAgentsWithLocalSkillDirDedupesByPresence(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	shared := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(shared, 0o755); err != nil {
		t.Fatal(err)
	}

	visible := AgentsWithLocalSkillDir(DefaultAgents(), root, home)
	if len(visible) == 0 {
		t.Fatal("expected at least one agent for .agents/skills")
	}
	for _, a := range visible {
		if !HasLocalSkillDir(a, root, home) {
			t.Fatalf("agent %s should have local skill dir", a.ID)
		}
	}
}
