package app

import "github.com/JoeHe0x/skill-man/internal/app/session"

// SessionState is the Bubble Tea session mode (re-exported from session package).
type SessionState = session.State

const (
	stateHome           = session.Home
	stateListing        = session.Listing
	stateSearching      = session.Searching
	stateInstalling     = session.Installing
	stateConfirming     = session.Confirming
	stateHelpOverlay    = session.HelpOverlay
	stateBindingAgent   = session.BindingAgent
	stateFilteringAgent = session.FilteringAgent
	stateInspecting     = session.Inspecting
	stateCommandPalette = session.CommandPalette
)
