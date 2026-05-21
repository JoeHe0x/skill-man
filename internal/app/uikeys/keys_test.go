package uikeys

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestEnterDoesNotMatchCancel(t *testing.T) {
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	if key.Matches(enter, Default.Cancel) {
		t.Fatal("Enter must not match Cancel — breaks install dialog search")
	}
	if !key.Matches(enter, Default.Enter) {
		t.Fatal("Enter should match Default.Enter")
	}
}
