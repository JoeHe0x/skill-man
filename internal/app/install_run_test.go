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

	m.install.flow.installing = true
	m.install.flow.selected.Name = "demo-skill"
	m.install.flow.quitPending = false
	m.install.cancel = func() {}

	updated, _ = m.handleInstallQuitAttempt()
	m = mustModel(t, updated)
	if !m.install.flow.quitPending {
		t.Fatal("first Esc should set quitPending")
	}
	if m.install.cancel == nil {
		t.Fatal("expected active install cancel func")
	}

	updated, _ = m.handleInstallQuitAttempt()
	m = mustModel(t, updated)
	if m.install.flow.installing {
		t.Fatal("second Esc should stop installing state")
	}
}
