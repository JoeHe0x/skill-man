package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func (m *Model) handleMutationCompleted(msg mutationCompletedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.reportError(msg.err)
		m.updateFooterForState(m.state)
		return m, m.beginScanAllCmd()
	}
	m.clearError()
	m.status = "ready"
	m.updateFooterForState(m.state)
	var flashCmd tea.Cmd
	if msg.message != "" {
		flashCmd = m.flashFooter(msg.message)
	}
	if msg.selectName != "" {
		if msg.targetTab == panel.TabMCP {
			return m, tea.Batch(flashCmd, tea.Sequence(
				m.beginScanAllCmd(),
				func() tea.Msg { return reselectMCPMsg{name: msg.selectName} },
			))
		}
		return m, tea.Batch(flashCmd, tea.Sequence(
			m.beginScanAllCmd(),
			func() tea.Msg { return reselectSkillMsg{name: msg.selectName} },
		))
	}
	return m, tea.Batch(flashCmd, m.beginScanAllCmd())
}

func (m *Model) handleInstallCompleted(msg installCompletedMsg) (tea.Model, tea.Cmd) {
	m.install.bg = nil
	m.clearInstallFlow()
	if m.state == stateInstalling {
		m.transitionTo(stateListing)
	}
	if msg.err != nil {
		m.reportError(msg.err)
		return m, m.beginScanAllCmd()
	}
	m.clearError()
	m.status = "ready"
	return m, tea.Batch(
		m.flashFooter(fmt.Sprintf("✓ Installed %s — back in skill list", msg.name)),
		tea.Sequence(
			m.beginScanAllCmd(),
			func() tea.Msg { return reselectSkillMsg{name: msg.name} },
		),
	)
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.resizeComponents()
	if m.palette != nil {
		m.palette.input.Width = paletteInputWidth(m.contentWidth())
	}
	return m, m.syncSelectionPreview()
}

func (m *Model) handleMouseDispatch(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.state == stateHelpOverlay {
		return m.handleHelpOverlayMouse(msg)
	}
	return m.handleMouseMsg(msg)
}

func (m *Model) handlePreviewLoaded(msg panel.PreviewLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Gen != m.previewGen || msg.Tab != m.activeTab {
		return m, nil
	}
	if msg.Err != nil {
		m.preview.SetContent("Preview failed:\n\n" + msg.Err.Error())
	} else {
		m.previewBody = msg.Content
		m.preview.SetContent(msg.Content)
	}
	m.clearStaleLoadingIfIdle()
	return m, nil
}

// clearStaleLoadingIfIdle resets status after non-scan work (e.g. inspect file preview)
// when no panel scan batch is in flight.
func (m *Model) clearStaleLoadingIfIdle() {
	if m.scansPending == 0 && m.status == "loading" {
		m.status = "ready"
		m.updateFooterForState(m.state)
	}
}

func (m *Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	if m.state == stateInstalling && m.install.flow != nil && m.install.flow.Searching() {
		next, flowCmd := m.install.flow.Update(msg)
		m.install.flow = &next
		return m, tea.Batch(cmd, flowCmd)
	}
	return m, cmd
}

func (m *Model) handleProgressFrame(msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	if m.install.bg != nil {
		if cmd, ok := m.install.bg.handleFrame(msg); ok {
			return m, cmd
		}
	}
	return m, nil
}

// --- helpers ---

func mcpKeyDisabled(members []*mcpdomain.Server) bool {
	if len(members) == 0 {
		return false
	}
	for _, srv := range members {
		if !srv.AggregatedDisabled() {
			return false
		}
	}
	return true
}

func (m *Model) selectedListItem() (panel.Item, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return panel.Item{}, false
	}
	li, ok := item.(panel.Item)
	return li, ok
}

func mcpKeyFromListItem(li panel.Item) string {
	if li.MCPKey != "" {
		return li.MCPKey
	}
	if li.MCP != nil && li.MCP.ConfigKey != "" {
		return li.MCP.ConfigKey
	}
	return ""
}

func (m *Model) handleSkillsScanned(msg panel.SkillsScannedMsg) (tea.Model, tea.Cmd) {
	if msg.Gen != m.scanGen {
		return m, nil
	}
	if msg.Err != nil {
		m.reportError(msg.Err)
		return m, m.noteScanCompleted(msg.Gen)
	}
	m.panels.Get(panel.TabSkills).ApplyScan(msg)
	m.clearError()
	if m.state == stateInstalling && m.install.flow != nil {
		return m, m.noteScanCompleted(msg.Gen)
	}
	var cmds []tea.Cmd
	if m.activeTab == panel.TabSkills && (m.state == stateHome || m.state == stateListing || m.state == stateSearching) {
		m.refreshActiveList()
		cmds = append(cmds, m.syncSelectionPreview())
	}
	cmds = append(cmds, m.noteScanCompleted(msg.Gen))
	return m, tea.Batch(cmds...)
}

func (m *Model) handleMCPScanned(msg panel.MCPScannedMsg) (tea.Model, tea.Cmd) {
	if msg.Gen != m.scanGen {
		return m, nil
	}
	if msg.Err != nil {
		m.reportError(msg.Err)
		return m, m.noteScanCompleted(msg.Gen)
	}
	if !m.panels.Get(panel.TabMCP).ApplyScan(msg) {
		return m, m.noteScanCompleted(msg.Gen)
	}
	m.clearError()
	var cmds []tea.Cmd
	if m.activeTab == panel.TabMCP && (m.state == stateHome || m.state == stateListing || m.state == stateSearching) {
		m.refreshActiveList()
		cmds = append(cmds, m.syncSelectionPreview())
	}
	cmds = append(cmds, m.noteScanCompleted(msg.Gen))
	return m, tea.Batch(cmds...)
}

func (m *Model) handleReselectMCP(msg reselectMCPMsg) (tea.Model, tea.Cmd) {
	if m.selectMCPByName(msg.name) {
		return m, tea.Batch(m.flashFooter("selected MCP "+msg.name), m.syncSelectionPreview())
	}
	return m, nil
}

func (m *Model) handleReselectSkill(msg reselectSkillMsg) (tea.Model, tea.Cmd) {
	if m.selectSkillByName(msg.name) {
		return m, tea.Batch(m.flashFooter("selected "+msg.name), m.syncSelectionPreview())
	}
	return m, nil
}

func (m *Model) handleFallthroughMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == stateInstalling && m.install.flow != nil {
		model, cmd := m.handleInstallingUpdate(msg)
		m.syncInstallHint()
		return model, cmd
	}
	var listCmd, previewCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.preview, previewCmd = m.preview.Update(msg)
	return m, tea.Batch(listCmd, previewCmd)
}
