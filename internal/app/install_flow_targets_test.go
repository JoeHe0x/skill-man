package app

import (
	"slices"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func TestNewInstallDirChoicesRespectsAgentFilter(t *testing.T) {
	choices := newInstallDirChoices([]string{"cursor"})
	if len(choices) == 0 {
		t.Fatal("expected install dir choices")
	}

	var cursorDir *installDirChoice
	for i := range choices {
		if slices.ContainsFunc(choices[i].agents, func(a agent.Agent) bool { return a.ID == "cursor" }) {
			cursorDir = &choices[i]
			break
		}
	}
	if cursorDir == nil {
		t.Fatal("expected a choice row containing cursor")
	}
	if !cursorDir.desired {
		t.Fatal("cursor's skill dir should be selected by default when filter is cursor")
	}

	for _, c := range choices {
		if c.skillDir == cursorDir.skillDir {
			continue
		}
		if c.desired {
			t.Fatalf("dir %q should not be selected when filter is cursor", c.skillDir)
		}
	}
}

func TestSelectedInstallAgentIDsFromDirs(t *testing.T) {
	targets := []installDirChoice{
		{
			skillDir: ".agents/skills",
			agents: []agent.Agent{
				{ID: "cursor", Name: "Cursor"},
				{ID: "amp", Name: "Amp"},
			},
			desired: true,
		},
		{
			skillDir: ".claude/skills",
			agents:   []agent.Agent{{ID: "claude-code", Name: "Claude Code"}},
			desired:  false,
		},
	}
	ids := selectedInstallAgentIDs(targets)
	if len(ids) != 2 || !slices.Contains(ids, "cursor") || !slices.Contains(ids, "amp") {
		t.Fatalf("unexpected agent ids: %v", ids)
	}
}
