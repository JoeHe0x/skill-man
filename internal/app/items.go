package app

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
)

func commandListItems(specs []commands.Spec) []list.Item {
	items := panel.CommandItems(specs)
	out := make([]list.Item, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
