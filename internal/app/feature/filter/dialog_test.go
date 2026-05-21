package filter

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func TestFilterAgentTitle_activeMarker(t *testing.T) {
	if got := filterAgentTitle("Cursor", true); got != "● Cursor" {
		t.Fatalf("active title = %q", got)
	}
	if got := filterAgentTitle("Cursor", false); got != "  Cursor" {
		t.Fatalf("inactive title = %q", got)
	}
}

func TestNewAgentFilterListItems_includesAllRow(t *testing.T) {
	items := newAgentFilterListItems([]agent.Agent{{ID: "cursor", Name: "Cursor"}}, "all")
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2 (all + one agent)", len(items))
	}
}
