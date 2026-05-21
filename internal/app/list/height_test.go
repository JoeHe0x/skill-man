package list

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestHeightForMessageItems(t *testing.T) {
	t.Parallel()

	items := []list.Item{
		panel.Item{Kind: panel.ItemMessage, Title: "A", Desc: "path-a"},
		panel.Item{Kind: panel.ItemMessage, Title: "B", Desc: "path-b"},
	}
	if got := HeightForItems(items); got != 1 {
		t.Fatalf("message items want delegate height 1, got %d", got)
	}
}
