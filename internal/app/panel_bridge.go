package app

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func panelToListItems(items []panel.Item) []list.Item {
	out := make([]list.Item, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}

func appViewState(state SessionState) panel.ViewState {
	switch state {
	case stateSearching:
		return panel.ViewSearching
	case stateInstalling:
		return panel.ViewInstalling
	case stateHelpOverlay:
		return panel.ViewListing
	case stateBindingAgent:
		return panel.ViewBinding
	case stateInspecting:
		return panel.ViewInspecting
	default:
		return panel.ViewListing
	}
}
