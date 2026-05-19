package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInstallQuitAttempt_requiresConfirmation(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mustModel(t, updated)
	updated, _ = m.startInstallFlow()
	m = mustModel(t, updated)

	m.installFlow.installing = true
	m.installFlow.selected.Name = "demo-skill"
	m.installFlow.quitPending = false
	m.installCancel = func() {}

	updated, _ = m.handleInstallQuitAttempt()
	m = mustModel(t, updated)
	if !m.installFlow.quitPending {
		t.Fatal("first Esc should set quitPending")
	}
	if m.installCancel == nil {
		t.Fatal("expected active install cancel func")
	}

	updated, _ = m.handleInstallQuitAttempt()
	m = mustModel(t, updated)
	if m.installFlow.installing {
		t.Fatal("second Esc should stop installing state")
	}
}
