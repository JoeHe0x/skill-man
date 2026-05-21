package app

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, consumed := m.dispatchToFeatures(msg); consumed {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tea.KeyMsg:
		return m.dispatchKey(msg)

	case tea.MouseMsg:
		return m.handleMouseDispatch(msg)

	case panel.ScannedMsg:
		return m.handleScanned(msg)

	case panel.PreviewLoadedMsg:
		return m.handlePreviewLoaded(msg)

	case uimsg.MutationCompleted:
		return m.handleMutationCompleted(msg)

	case uimsg.ReselectMCP:
		return m.handleReselectMCP(msg)

	case uimsg.ReselectSkill:
		return m.handleReselectSkill(msg)

	case footerFlashTimeoutMsg:
		return m.handleFooterFlashTimeout(msg)

	case theme.DetectedMsg:
		return m.handleThemeDetected(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)

	case progress.FrameMsg:
		return m.handleProgressFrame(msg)
	}

	return m.handleFallthroughMsg(msg)
}

func (m *Model) dispatchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == stateInstalling && m.install.WizardOpen() {
		return m.handleInstallingUpdate(msg)
	}
	if m.prompt.Active() {
		return m, nil // consumed by promptFeature in dispatchToFeatures
	}
	if m.listFilterActive() {
		return m.handleListFilterKeys(msg)
	}
	return m.handleKeyMsg(msg)
}
