package list

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

// HeightForItems picks a delegate row height for the given list items.
func HeightForItems(items []list.Item) int {
	h := 3

	allMessages := len(items) > 0
	for _, it := range items {
		if li, ok := it.(panel.Item); !ok || li.Kind != panel.ItemMessage {
			allMessages = false
			break
		}
	}

	if allMessages {
		return 1
	}

	for _, it := range items {
		li, ok := it.(panel.Item)
		if !ok || len(li.DetailLines) == 0 {
			continue
		}
		need := 2 + len(li.DetailLines)
		if need > h {
			h = need
		}
	}
	return h
}
