package app

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func (m *Model) handleMutationCompleted(msg mutationCompletedMsg) (tea.Model, tea.Cmd) {
	return m.applyMutationResult(msg)
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.resizeComponents()
	m.cmdPalette.resizeInput()
	return m, m.syncSelectionPreview()
}

func (m *Model) handleMouseDispatch(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
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
	if m.scan.Pending == 0 && m.status == "loading" {
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
	var listCmd, previewCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	m.preview, previewCmd = m.preview.Update(msg)
	return m, tea.Batch(listCmd, previewCmd)
}
