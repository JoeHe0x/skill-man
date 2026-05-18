package app

func (m *Model) shouldStack() bool {
	return m.width < 80
}

func (m *Model) mainAreaSize() (int, int) {
	contentWidth := m.width - 2
	contentHeight := m.height

	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	footerHeight := 2

	mainHeight := contentHeight - m.headerHeight() - footerHeight - 2
	if mainHeight < 6 {
		mainHeight = 6
	}

	return contentWidth, mainHeight
}

func (m *Model) paneSizes() (int, int, int, int) {
	contentWidth, mainHeight := m.mainAreaSize()
	if m.shouldStack() {
		topHeight := mainHeight / 2
		bottomHeight := mainHeight - topHeight
		return contentWidth, topHeight, contentWidth, bottomHeight
	}

	leftWidth := (contentWidth * 35) / 100
	if contentWidth < 120 {
		leftWidth = (contentWidth * 40) / 100
	}
	rightWidth := contentWidth - leftWidth
	return leftWidth, mainHeight, rightWidth, mainHeight
}
