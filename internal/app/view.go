package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	headerH := lipgloss.Height(header)
	footerH := lipgloss.Height(footer)
	mainH := max(6, m.height-headerH-footerH)

	main := m.renderMainAreaSized(mainH)
	if m.state == stateInstalling && m.installFlow != nil {
		main = clipLines(m.renderInstallDialogArea(), mainH)
	}
	if m.state == stateFilteringAgent {
		main = clipLines(m.renderAgentFilterDialogArea(), mainH)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
	if m.state == stateConfirming && m.pending != nil {
		body = m.renderModalOverlay(body)
	}
	if m.state == stateCommandPalette && m.palette != nil {
		body = m.renderPaletteOverlay(body)
	}
	if m.state == stateHelpOverlay {
		body = m.renderHelpOverlay(body)
	}
	return m.styles.doc.Render(body)
}

func (m *Model) renderMainAreaSized(mainHeight int) string {
	leftWidth, leftHeight, rightWidth, rightHeight := m.paneSizesFor(mainHeight)

	leftInnerWidth, leftInnerHeight := panelInnerSize(leftWidth, leftHeight)
	rightInnerWidth, rightInnerHeight := panelInnerSize(rightWidth, rightHeight)

	var leftContent string
	if m.state == stateInspecting {
		m.tree.SetSize(leftInnerWidth, leftInnerHeight)
		leftContent = m.tree.View()
	} else {
		mutableList := m.list
		if m.state == stateBindingAgent {
			mutableList = m.agentList
		}
		mutableList.SetSize(leftInnerWidth, leftInnerHeight)
		leftContent = mutableList.View()
	}

	mutablePreview := m.preview
	mutablePreview.Width = rightInnerWidth
	mutablePreview.Height = rightInnerHeight

	leftStyle, leftTitleStyle := m.panelStyles(focusPaneList)
	leftPanel := leftStyle.Width(leftWidth).MaxHeight(leftHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			leftTitleStyle.Render(m.leftPanelTitle()),
			leftContent,
		),
	)

	previewContent := mutablePreview.View()
	if strings.TrimSpace(previewContent) == "" {
		previewContent = m.styles.emptyPreview.Render("Nothing to preview.")
	}

	rightStyle, rightTitleStyle := m.panelStyles(focusPanePreview)
	rightPanel := rightStyle.Width(rightWidth).MaxHeight(rightHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			rightTitleStyle.Render("Preview"),
			previewContent,
		),
	)

	var out string
	if m.shouldStack() {
		out = lipgloss.JoinVertical(lipgloss.Left, leftPanel, rightPanel)
	} else {
		out = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}
	return clipLines(out, mainHeight)
}

func (m *Model) renderFooter() string {
	var content string
	if m.prompt != nil {
		content = m.renderPromptFooter()
	} else {
		content = m.renderHintFooter()
	}
	return m.styles.footer.Render(content)
}

func (m *Model) panelStyles(pane focusPane) (panel, title lipgloss.Style) {
	if m.focusedPane == pane {
		return m.styles.panelFocused, m.styles.panelTitleFocus
	}
	return m.styles.panelBlur, m.styles.panelTitleBlur
}

func (m *Model) renderHintFooter() string {
	var lines []string
	switch {
	case m.errMsg != "":
		lines = append(lines, m.styles.statusError.Render(truncate(m.errMsg, max(20, m.width-6))))
	case m.footerFlash != "":
		lines = append(lines, m.styles.footerFlash.Render(truncate(m.footerFlash, max(20, m.width-6))))
	case m.footerContext != "":
		lines = append(lines, m.styles.footerContext.Render(truncate(m.footerContext, max(20, m.width-6))))
	}
	lines = append(lines, m.renderHelpFooter())
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) renderPromptFooter() string {
	label := m.styles.hintBold.Render(m.prompt.label + ": ")
	input := m.prompt.input.View()
	helpLine := m.styles.hint.Render("Enter=confirm  Esc=cancel")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, label, input),
		helpLine,
	)
}

func (m *Model) renderModalOverlay(base string) string {
	target := m.pending.skillName
	if m.pending.mcpName != "" {
		target = "MCP " + m.pending.mcpName
	}
	modalText := fmt.Sprintf(
		"Are you sure?\n\nYou are about to remove:\n[%s]\n\nPress 'y' to confirm, 'n' or Esc to abort.",
		target,
	)

	box := m.styles.modalDanger.Width(min(52, max(36, m.width/2))).Render(modalText)
	return lipgloss.Place(m.width-2, m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) leftPanelTitle() string {
	return m.activePanel().PanelTitle(appViewState(m.state))
}

func (m *Model) resizeComponents() {
	leftWidth, leftHeight, rightWidth, rightHeight := m.paneSizes()
	lw, lh := panelInnerSize(leftWidth, leftHeight)
	rw, rh := panelInnerSize(rightWidth, rightHeight)
	m.list.SetSize(lw, lh)
	m.agentList.SetSize(lw, lh)
	m.preview.Width = rw
	m.preview.Height = rh
}

func truncate(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	if limit == 1 {
		return "…"
	}
	return string(runes[:limit-1]) + "…"
}
