package app

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
)

// actionable describes operations the app layer can perform on a selected panel.Item.
// panel.Item now implements these methods directly; this interface documents the contract.
type actionable interface {
	list.Item

	CanInspect() bool
	CanDisable() bool
	CanRemove() bool
	CanBind() bool
	CanUpdate() bool

	InspectTarget() panel.InspectTarget
	DisableTarget() panel.DisableTarget
	RemoveTarget() panel.RemoveTarget
	BindTarget() panel.BindTarget
	UpdateTarget() panel.UpdateTarget
}

func commandListItems(specs []commands.Spec) []list.Item {
	items := panel.CommandItems(specs)
	out := make([]list.Item, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}
