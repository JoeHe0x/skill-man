package fallback

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

func HandleMutationCompleted(h Host, msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	return h.ApplyMutationResult(msg)
}

func HandleWindowResize(h Host, msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	h.SetWindowSize(msg.Width, msg.Height)
	h.ResizeComponents()
	h.ResizePaletteInput()
	return h.TeaModel(), h.SyncSelectionPreview()
}

func HandleMouse(h Host, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return h.HandleMouse(msg)
}

func HandlePreviewLoaded(h Host, msg panel.PreviewLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Gen != h.PreviewGeneration() || msg.Tab != h.ActiveTab() {
		return h.TeaModel(), nil
	}
	if msg.Err != nil {
		h.SetPreviewError(msg.Err.Error())
	} else {
		h.SetPreviewBody(msg.Content)
		h.SetPreviewContent(msg.Content)
	}
	h.ClearStaleLoadingIfIdle()
	return h.TeaModel(), nil
}

func HandleSpinnerTick(h Host, msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	cmd := h.SpinnerTick(msg)
	if h.InstallWizardSearching() {
		return h.TeaModel(), tea.Batch(cmd, h.InstallHandleUIMsg(msg))
	}
	return h.TeaModel(), cmd
}

func HandleProgressFrame(h Host, msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	if cmd, ok := h.InstallHandleBackgroundFrame(msg); ok {
		return h.TeaModel(), cmd
	}
	return h.TeaModel(), nil
}

func HandleReselectMCP(h Host, msg uimsg.ReselectMCP) (tea.Model, tea.Cmd) {
	if h.SelectMCPByName(msg.Name) {
		return h.TeaModel(), tea.Batch(h.FlashFooter("selected MCP "+msg.Name), h.SyncSelectionPreview())
	}
	return h.TeaModel(), nil
}

func HandleReselectSkill(h Host, msg uimsg.ReselectSkill) (tea.Model, tea.Cmd) {
	if h.SelectSkillByName(msg.Name) {
		return h.TeaModel(), tea.Batch(h.FlashFooter("selected "+msg.Name), h.SyncSelectionPreview())
	}
	return h.TeaModel(), nil
}

func HandleFallthrough(h Host, msg tea.Msg) (tea.Model, tea.Cmd) {
	listCmd, previewCmd := h.MainFallthrough(msg)
	return h.TeaModel(), tea.Batch(listCmd, previewCmd)
}
