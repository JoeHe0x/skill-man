package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func TestListHeightForMessageItems(t *testing.T) {
	t.Parallel()

	items := []list.Item{
		listItem{kind: itemKindMessage, title: "A", desc: "path-a"},
		listItem{kind: itemKindMessage, title: "B", desc: "path-b"},
	}
	if got := listHeightForItems(items); got != 1 {
		t.Fatalf("message items want delegate height 1, got %d", got)
	}
}

func TestSetAgentListItemsUsesCompactDelegateHeight(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir(), t.TempDir())
	items := []list.Item{
		listItem{kind: itemKindMessage, title: "✓ Cursor", desc: ".cursor"},
	}
	m.setAgentListItems(items)

	if m.agentListDelegate.Height() != 1 {
		t.Fatalf("agent list delegate height = %d, want 1 for bind rows", m.agentListDelegate.Height())
	}
}
