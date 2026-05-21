package app

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	statefallback "github.com/JoeHe0x/skill-man/internal/app/state/fallback"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

func (m *Model) handleMutationCompleted(msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	return statefallback.HandleMutationCompleted(m, msg)
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleWindowResize(m, msg)
}

func (m *Model) handleMouseDispatch(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleMouse(m, msg)
}

func (m *Model) handlePreviewLoaded(msg panel.PreviewLoadedMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandlePreviewLoaded(m, msg)
}

func (m *Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleSpinnerTick(m, msg)
}

func (m *Model) handleProgressFrame(msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleProgressFrame(m, msg)
}

func (m *Model) handleReselectMCP(msg uimsg.ReselectMCP) (tea.Model, tea.Cmd) {
	return statefallback.HandleReselectMCP(m, msg)
}

func (m *Model) handleReselectSkill(msg uimsg.ReselectSkill) (tea.Model, tea.Cmd) {
	return statefallback.HandleReselectSkill(m, msg)
}

func (m *Model) handleFallthroughMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	return statefallback.HandleFallthrough(m, msg)
}

func (m *Model) ApplyMutationResult(msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	return m.applyMutationResult(msg)
}

func (m *Model) SetWindowSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) ResizeComponents() { m.resizeComponents() }

func (m *Model) ResizePaletteInput() { m.cmdPalette.ResizeInput() }

func (m *Model) HandleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return m.handleMouseMsg(msg)
}

func (m *Model) PreviewGeneration() int { return m.PreviewGen }

func (m *Model) SetPreviewError(err string) {
	m.Preview.SetContent("Preview failed:\n\n" + err)
}

func (m *Model) SetPreviewBody(body string) { m.PreviewBody = body }

func (m *Model) ClearStaleLoadingIfIdle() { m.clearStaleLoadingIfIdle() }

// clearStaleLoadingIfIdle resets status after non-scan work (e.g. inspect file preview)
// when no panel scan batch is in flight.
func (m *Model) clearStaleLoadingIfIdle() {
	if m.scan.Pending == 0 && m.status == "loading" {
		m.status = "ready"
		m.updateFooterForState(m.state)
	}
}

func (m *Model) SpinnerTick(msg spinner.TickMsg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m *Model) InstallWizardSearching() bool {
	return m.state == stateInstalling && m.install.WizardOpen() && m.install.WizardSearching()
}

func (m *Model) InstallHandleUIMsg(msg tea.Msg) tea.Cmd {
	_, cmd := m.install.HandleUIMsg(msg)
	return cmd
}

func (m *Model) InstallHandleBackgroundFrame(msg progress.FrameMsg) (tea.Cmd, bool) {
	return m.install.HandleBackgroundFrame(msg)
}

func (m *Model) SelectMCPByName(name string) bool { return m.selectMCPByName(name) }

func (m *Model) SelectSkillByName(name string) bool { return m.selectSkillByName(name) }

func (m *Model) MainFallthrough(msg tea.Msg) (tea.Cmd, tea.Cmd) {
	var listCmd, previewCmd tea.Cmd
	m.Main, listCmd = m.Main.Update(msg)
	m.Preview, previewCmd = m.Preview.Update(msg)
	return listCmd, previewCmd
}

var _ statefallback.Host = (*Model)(nil)
