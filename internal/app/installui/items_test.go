package installui

import (
	"slices"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestNewDirChoicesRespectsAgentFilter(t *testing.T) {
	choices := newDirChoices([]string{"cursor"})
	var cursorDir *dirChoice
	for i := range choices {
		if slices.ContainsFunc(choices[i].agents, func(a agent.Agent) bool { return a.ID == "cursor" }) {
			cursorDir = &choices[i]
			break
		}
	}
	if cursorDir == nil || !cursorDir.desired {
		t.Fatal("expected cursor skill dir selected by default")
	}
}

func TestResultsToItems_showsInstallCount(t *testing.T) {
	items := resultsToItems([]domaininstall.Candidate{{
		Name: "react-hooks", Source: "vercel-labs/skills@react-hooks", Installs: "12.5K installs",
	}})
	li := items[0].(Row)
	if li.Meta != "12.5K installs" {
		t.Fatalf("expected install meta, got %q", li.Meta)
	}
}
