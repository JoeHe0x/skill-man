package app

import tea "github.com/charmbracelet/bubbletea"

// mainAreaOrigin returns the top-left cell of the main (dual-pane) content area
// in terminal coordinates, accounting for doc horizontal padding.
func (m *Model) mainAreaOrigin() (x0, y0 int) {
	headerH, _ := m.chromeHeights()
	return 1, headerH
}

// paneFromMouse maps a terminal mouse position to list or preview pane when the
// point lies inside the main content region. Returns false for header/footer hits.
func (m *Model) paneFromMouse(x, y int) (focusPane, bool) {
	if m.width == 0 || m.height == 0 {
		return focusPaneList, false
	}
	x0, y0 := m.mainAreaOrigin()
	contentW, mainH := m.mainAreaSize()
	relX := x - x0
	relY := y - y0
	if relX < 0 || relY < 0 || relX >= contentW || relY >= mainH {
		return focusPaneList, false
	}

	leftW, leftH, _, _ := m.paneSizesFor(mainH)
	if m.shouldStack() {
		if relY < leftH {
			return focusPaneList, true
		}
		return focusPanePreview, true
	}
	if relX < leftW {
		return focusPaneList, true
	}
	return focusPanePreview, true
}

func (m *Model) mouseFocusEnabled() bool {
	if m.prompt.Active() {
		return false
	}
	switch m.state {
	case stateInstalling, stateConfirming, stateCommandPalette, stateFilteringAgent, stateHelpOverlay:
		return false
	default:
		return true
	}
}

func (m *Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if !m.mouseFocusEnabled() {
		return m, nil
	}

	pane, ok := m.paneFromMouse(msg.X, msg.Y)
	if !ok {
		return m, nil
	}

	m.focusedPane = pane

	if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown ||
		msg.Button == tea.MouseButtonWheelLeft || msg.Button == tea.MouseButtonWheelRight {
		if pane == focusPanePreview {
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if msg.Button != tea.MouseButtonLeft && msg.Button != tea.MouseButtonNone {
		return m, nil
	}

	return m, nil
}
