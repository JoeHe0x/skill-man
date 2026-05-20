package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestListHeightForMessageItems(t *testing.T) {
	t.Parallel()

	items := []list.Item{
		panel.Item{Kind: panel.ItemMessage, Title: "A", Desc: "path-a"},
		panel.Item{Kind: panel.ItemMessage, Title: "B", Desc: "path-b"},
	}
	if got := listHeightForItems(items); got != 1 {
		t.Fatalf("message items want delegate height 1, got %d", got)
	}
}

func TestSetAgentListItemsUsesCompactDelegateHeight(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir(), t.TempDir())
	items := []list.Item{
		panel.Item{Kind: panel.ItemMessage, Title: "✓ Cursor", Desc: ".cursor"},
	}
	m.setAgentListItems(items)

	if m.agentListDelegate.Height() != 1 {
		t.Fatalf("agent list delegate height = %d, want 1 for bind rows", m.agentListDelegate.Height())
	}
}
