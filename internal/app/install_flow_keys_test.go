package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInstallDialogEnterDoesNotMatchCancel(t *testing.T) {
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	if key.Matches(enter, keys.Cancel) {
		t.Fatal("Enter must not match Cancel — breaks install dialog search")
	}
	if !key.Matches(enter, keys.Enter) {
		t.Fatal("Enter should match keys.Enter")
	}
}
