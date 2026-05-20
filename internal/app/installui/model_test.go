package installui

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

func TestSyncSearchSuggestions_includesResults(t *testing.T) {
	m := New(Config{
		Styles:   theme.NewStyles(true),
		Provider: serviceinstall.NewSkillsCLIProvider(),
	})
	m.results = []domaininstall.Candidate{
		{Name: "react-hooks", Source: "vercel-labs/skills@react-hooks"},
	}
	m.syncSearchSuggestions()
	sugs := m.searchInput.AvailableSuggestions()
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
	m := &Model{}
	m.rememberSearchQuery("go")
	m.rememberSearchQuery("GO")
	if len(m.recentQueries) != 1 {
		t.Fatalf("expected 1 recent query, got %d", len(m.recentQueries))
	}
}
