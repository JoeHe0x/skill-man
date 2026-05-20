package app

import (
	"fmt"
)

// allowedTransitions defines which state transitions are valid.
// Transitions not listed here are silently rejected.
var allowedTransitions = map[SessionState][]SessionState{
	stateHome:    {stateListing, stateSearching, stateInstalling, stateBindingAgent, stateInspecting, stateConfirming, stateHelpOverlay, stateCommandPalette, stateFilteringAgent},
	stateListing: {stateHome, stateSearching, stateInstalling, stateBindingAgent, stateInspecting, stateConfirming, stateHelpOverlay, stateCommandPalette, stateFilteringAgent},
	// Modal / overlay states can only return to listing or home.
	stateSearching:      {stateListing, stateHome},
	stateInstalling:     {stateListing, stateHome},
	stateConfirming:     {stateListing, stateHome},
	stateHelpOverlay:    {stateListing, stateHome},
	stateBindingAgent:   {stateListing, stateHome},
	stateFilteringAgent: {stateListing, stateHome},
	stateInspecting:     {stateListing, stateHome},
	stateCommandPalette: {stateListing, stateHome},
}

// transition attempts to move to the target state. Returns true if the
// transition is valid, false otherwise. Runs exit/enter hooks as needed.
func (m *Model) transitionTo(target SessionState) bool {
	if m.state == target {
		return true
	}
	if !canTransition(m.state, target) {
		return false
	}
	m.exitState(m.state)
	prev := m.state
	m.state = target
	m.enterState(target, prev)
	return true
}

func canTransition(from, to SessionState) bool {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// exitState runs cleanup hooks when leaving a state.
func (m *Model) exitState(state SessionState) {
	switch state {
	case stateInstalling:
		m.install.flow = nil
		// Background install keeps running; do not abort.
	case stateBindingAgent:
		m.clearBindingSession()
	case stateConfirming:
		m.pending = nil
	case stateFilteringAgent:
		// nothing to clean
	case stateInspecting:
		// nothing to clean
	case stateCommandPalette:
		m.palette = nil
	case stateHelpOverlay:
		m.helpOverlay = helpOverlay{}
	}
}

// enterState runs setup hooks when entering a state.
func (m *Model) enterState(state SessionState, prev SessionState) {
	switch state {
	case stateListing:
		m.refreshActiveList()
	case stateHome:
		m.refreshActiveList()
		if preview := m.activePanel().StaticPreview(); preview != "" {
			m.preview.SetContent(preview)
		}
	case stateConfirming:
		m.lastState = prev
	case stateBindingAgent:
		m.lastState = prev
		m.resizeComponents()
		m.syncBindHint()
	case stateInstalling:
		m.lastState = prev
	case stateInspecting:
		m.lastState = prev
	case stateFilteringAgent:
		m.lastState = prev
	case stateCommandPalette:
		m.lastState = prev
	case stateHelpOverlay:
		m.lastState = prev
	}
	m.clearError()
	m.updateFooterForState(state)
}

// updateFooterForState sets the footer context based on the current state.
func (m *Model) updateFooterForState(state SessionState) {
	switch state {
	case stateListing:
		m.setFooterContext(fmt.Sprintf("%d %s · agents: %s", m.activePanel().Count(), m.activePanel().CountLabel(), m.agentDisplay()))
	case stateHome:
		m.setFooterContext("home")
	case stateSearching:
		// footer set by search handler
	case stateInstalling:
		if m.install.flow != nil {
			m.syncInstallHint()
		}
	case stateBindingAgent:
		m.syncBindHint()
	default:
		// keep whatever was set
	}
}
