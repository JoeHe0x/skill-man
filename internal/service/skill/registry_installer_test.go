package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestInstallRegistrySkill_multipleAgentDirs(t *testing.T) {
	workspace := t.TempDir()
	files := []RegistrySnapshotFile{{
		Path:     "SKILL.md",
		Contents: "---\nname: shared-skill\ndescription: x\n---\n",
	}}

	cursor, _ := agent.AgentByID("cursor")
	claude, _ := agent.AgentByID("claude-code")

	_, err := InstallRegistrySkill(workspace, "", extension.ScopeProject, "acme/repo@shared-skill", files, []agent.Agent{cursor, claude})
	if err != nil {
		t.Fatalf("InstallRegistrySkill: %v", err)
	}

	sharedPath := filepath.Join(workspace, ".agents/skills", "shared-skill", "SKILL.md")
	claudePath := filepath.Join(workspace, ".claude/skills", "shared-skill", "SKILL.md")
	if _, err := os.Stat(sharedPath); err != nil {
		t.Fatalf("missing shared dir install: %v", err)
	}
	if _, err := os.Stat(claudePath); err != nil {
		t.Fatalf("missing claude dir install: %v", err)
	}
}
