package app

import (
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

func (m *Model) transitionTo(target SessionState) bool {
	if m.state == target {
		return true
	}
	if !session.CanTransition(m.state, target) {
		return false
	}
	m.exitState(m.state)
	prev := m.state
	m.state = target
	m.enterState(target, prev)
	return true
}

func (m *Model) exitState(state SessionState) {
	switch state {
	case stateInstalling:
		m.install.ClearWizard()
	case stateBindingAgent:
		m.bind.Clear()
	case stateConfirming:
		m.confirm.Clear()
	case stateFilteringAgent:
		// nothing to clean
	case stateInspecting:
		// nothing to clean
	case stateCommandPalette:
		m.cmdPalette.Close()
	case stateHelpOverlay:
		m.helpScreen.Clear()
	}
}

func (m *Model) enterState(state SessionState, prev SessionState) {
	switch state {
	case stateListing:
		m.clearStaleLoadingIfIdle()
		m.refreshActiveList()
	case stateHome:
		m.refreshActiveList()
		if preview := m.activePanel().StaticPreview(); preview != "" {
			m.Preview.SetContent(preview)
		}
	case stateConfirming:
		m.lastState = prev
		m.confirm.BeginRemoveConfirm()
	case stateBindingAgent:
		m.lastState = prev
		m.resizeComponents()
		m.bind.SyncHint()
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

func (m *Model) updateFooterForState(state SessionState) {
	switch state {
	case stateListing:
		if m.status == "loading" {
			m.setFooterContext(m.scanLoadingLabel())
		} else {
			m.setFooterContext(m.footerStatsLine())
		}
	case stateHome:
		if m.status == "loading" {
			m.setFooterContext(m.scanLoadingLabel())
		} else {
			m.setFooterContext("home")
		}
	case stateSearching:
		// footer set by search handler
	case stateInstalling:
		if m.install.Active() {
			m.install.SyncHint()
		}
	case stateBindingAgent:
		m.bind.SyncHint()
	default:
		// keep whatever was set
	}
}
