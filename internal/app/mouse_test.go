package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPaneFromMouse_sideBySide(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = mustModel(t, updated)

	_, mainY := m.mainAreaOrigin()
	_, mainH := m.mainAreaSize()
	leftW, _, _, _ := m.paneSizesFor(mainH)

	if pane, ok := m.paneFromMouse(2, mainY+2); !ok || pane != focusPaneList {
		t.Fatalf("left click: got pane=%v ok=%v", pane, ok)
	}
	if pane, ok := m.paneFromMouse(2+leftW, mainY+2); !ok || pane != focusPanePreview {
		t.Fatalf("right click: got pane=%v ok=%v", pane, ok)
	}
}

func TestPaneFromMouse_stacked(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 70, Height: 40})
	m = mustModel(t, updated)

	if !m.shouldStack() {
		t.Fatal("expected stacked layout below width 80")
	}

	_, mainY := m.mainAreaOrigin()
	_, mainH := m.mainAreaSize()
	_, topH, _, _ := m.paneSizesFor(mainH)

	if pane, ok := m.paneFromMouse(5, mainY+1); !ok || pane != focusPaneList {
		t.Fatalf("top click: got pane=%v ok=%v", pane, ok)
	}
	if pane, ok := m.paneFromMouse(5, mainY+topH); !ok || pane != focusPanePreview {
		t.Fatalf("bottom click: got pane=%v ok=%v", pane, ok)
	}
}

func TestHandleMouseMsg_setsFocusedPane(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = mustModel(t, updated)
	m.focusedPane = focusPaneList

	_, mainY := m.mainAreaOrigin()
	_, mainH := m.mainAreaSize()
	leftW, _, _, _ := m.paneSizesFor(mainH)

	click := tea.MouseMsg{
		X:      2 + leftW,
		Y:      mainY + 3,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	updated, _ = m.Update(click)
	m = mustModel(t, updated)
	if m.focusedPane != focusPanePreview {
		t.Fatalf("expected preview focus, got %v", m.focusedPane)
	}
}

func TestHandleMouseMsg_ignoredDuringPalette(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = mustModel(t, updated)
	m.focusedPane = focusPaneList
	m.state = stateCommandPalette
	m.cmdPalette.ui = &commandPalette{}

	_, mainY := m.mainAreaOrigin()
	click := tea.MouseMsg{X: 50, Y: mainY + 3, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	updated, _ = m.Update(click)
	m = mustModel(t, updated)
	if m.focusedPane != focusPaneList {
		t.Fatalf("palette open: focus should not change, got %v", m.focusedPane)
	}
}
