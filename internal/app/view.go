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

	leftPanel := m.styles.panel.Width(leftWidth).MaxHeight(leftHeight).Render(
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

	rightPanel := m.styles.panel.Width(rightWidth).MaxHeight(rightHeight).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.panelTitle.Render("Preview"),
			previewContent,
		),
	)

	var out string
	if m.shouldStack() {
		out = lipgloss.JoinVertical(lipgloss.Left, leftPanel, rightPanel)
	} else {
		out = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}

	// Render the bind dialog as an overlay instead of inside the left pane if it's open
	if m.state == stateBindingAgent {
		out = m.renderBindDialogArea(out, mainHeight)
	}

	return clipLines(out, mainHeight)
}

func (m *Model) renderBindDialogArea(base string, mainHeight int) string {
	leftWidth, _, _, _ := m.paneSizesFor(mainHeight)

	// Determine dynamic width based on item contents
	maxWidth := len("Bind Agents")
	subtitle := "Select agents to bind to"
	if m.bindingMCP != nil {
		subtitle = "Bind MCP server to agents"
	}
	if len(subtitle) > maxWidth {
		maxWidth = len(subtitle)
	}

	for _, item := range m.agentList.Items() {
		if li, ok := item.(listItem); ok {
			// Title and desc are rendered on one line with "  " between them
			itemLen := len("  ") + len(li.title) + len("  ") + len(li.desc)
			if itemLen > maxWidth {
				maxWidth = itemLen
			}
		}
	}

	dialogWidth := maxWidth + 8 // 4 for inner padding, 4 for list padding
	if dialogWidth > leftWidth-2 {
		dialogWidth = leftWidth - 2
	}
	if dialogWidth < 20 {
		dialogWidth = 20
	}

	// Dynamic height based on items
	numItems := len(m.agentList.Items())
	listHeight := numItems + 2 // padding

	dialogHeight := listHeight + 8 // dialog frame padding
	dialogHeight = min(max(10, dialogHeight), min(mainHeight-2, 28))
	listHeight = dialogHeight - 8
	if listHeight < 2 {
		listHeight = 2
	}

	innerWidth := dialogWidth - 4
	m.agentList.SetSize(innerWidth, listHeight)

	body := lipgloss.JoinVertical(lipgloss.Left,
		m.styles.panelTitle.Render("Bind Agents"),
		m.styles.hint.Render(subtitle),
		m.agentList.View(),
	)

	dialog := m.styles.modal.Width(dialogWidth).Render(body)
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, dialog)
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
