package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

func TestSyncSearchSuggestions_includesResults(t *testing.T) {
	flow := newInstallFlow(serviceinstall.NewSkillsCLIProvider(), newItemDelegate(theme.NewStyles(true)))
	flow.results = []domaininstall.Candidate{
		{Name: "react-hooks", Source: "vercel-labs/skills@react-hooks"},
	}
	flow.syncSearchSuggestions()

	sugs := flow.searchInput.AvailableSuggestions()
	if len(sugs) < 2 {
		t.Fatalf("expected suggestions, got %v", sugs)
	}
	found := false
	for _, s := range sugs {
		if s == "react-hooks" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected react-hooks in %v", sugs)
	}
}

func TestRememberSearchQuery_dedupes(t *testing.T) {
	flow := &installFlow{}
	flow.rememberSearchQuery("go")
	flow.rememberSearchQuery("GO")
	if len(flow.recentQueries) != 1 {
		t.Fatalf("expected 1 recent query, got %d", len(flow.recentQueries))
	}
}
