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
	main := m.renderMainArea()
	footer := m.renderFooter()

	body := lipgloss.JoinVertical(lipgloss.Left, header, main, footer)
	if m.state == stateConfirming && m.pending != nil {
		body = m.renderModalOverlay(body)
	}
	return m.styles.doc.Render(body)
}

func (m *Model) renderHeader() string {
	if m.width >= 80 && m.height >= 24 {
		return m.renderFullHeader()
	}
	return m.renderCompactHeader()
}

func (m *Model) renderFullHeader() string {
	styledLogo := m.styles.logo.Render(asciiLogo)

	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.statusBarDim.Render("scope: project"),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("agents: %s", m.agentDisplay())),
		m.styles.statusBarSep.Render(" │ "),
		m.styles.statusBarDim.Render(fmt.Sprintf("skills: %d", len(m.skills))),
		m.styles.statusBarSep.Render(" │ "),
		m.statusView(),
	)
	paddedStatus := m.styles.statusBar.Render(statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, styledLogo, paddedStatus)
}

func (m *Model) renderCompactHeader() string {
	line1 := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.header.Render("skill-man v0.1"),
		"  ",
		m.styles.headerDim.Render(fmt.Sprintf("agents: %s", m.agentDisplay())),
		"  ",
		m.styles.headerDim.Render(fmt.Sprintf("skills: %d", len(m.skills))),
		"  ",
		m.statusView(),
	)

	return lipgloss.JoinVertical(lipgloss.Left, line1)
}

func (m *Model) renderMainArea() string {
	leftWidth, leftHeight, rightWidth, rightHeight := m.paneSizes()

	leftInnerWidth := max(8, leftWidth-4)
	leftInnerHeight := max(3, leftHeight-2)
	rightInnerWidth := max(8, rightWidth-4)
	rightInnerHeight := max(3, rightHeight-2)

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

	leftPanel := m.styles.panel.Width(leftWidth - 2).Height(leftInnerHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.panelTitle.Render(m.leftPanelTitle()),
			leftContent,
		),
	)

	previewContent := mutablePreview.View()
	if strings.TrimSpace(previewContent) == "" {
		previewContent = m.styles.emptyPreview.Render("Nothing to preview.")
	}

	rightPanel := m.styles.panel.Width(rightWidth - 2).Height(rightInnerHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.panelTitle.Render("Preview"),
			previewContent,
		),
	)

	if m.shouldStack() {
		return lipgloss.JoinVertical(lipgloss.Left, leftPanel, rightPanel)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
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

func (m *Model) renderHintFooter() string {
	hint := m.hint
	if m.errMsg != "" {
		hint = m.errMsg
	}
	return m.styles.hint.Render(truncate(hint, max(20, m.width-6)))
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
	modalText := fmt.Sprintf(
		"Are you sure?\n\nYou are about to remove:\n[%s]\n\nPress 'y' to confirm, 'n' to abort.",
		m.pending.skillName,
	)

	box := m.styles.modalDanger.Width(min(52, max(36, m.width/2))).Render(modalText)
	return lipgloss.Place(m.width-2, m.height-2, lipgloss.Center, lipgloss.Center, box, lipgloss.WithWhitespaceChars(" "))
}

func (m *Model) leftPanelTitle() string {
	switch m.state {
	case stateHome, stateListing:
		return "Skills"
	case stateSearching:
		return "Search Results"
	case stateViewingHelp:
		return "Commands"
	case stateBindingAgent:
		return "Bind Agents"
	case stateInspecting:
		return "Files"
	default:
		return "Skills"
	}
}

func (m *Model) resizeComponents() {
	leftWidth, leftHeight, rightWidth, rightHeight := m.paneSizes()
	m.list.SetSize(max(8, leftWidth-4), max(3, leftHeight-3))
	m.agentList.SetSize(max(8, leftWidth-4), max(3, leftHeight-3))
	m.preview.Width = max(8, rightWidth-4)
	m.preview.Height = max(3, rightHeight-3)
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
