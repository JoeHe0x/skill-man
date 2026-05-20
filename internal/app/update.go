package app

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
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

	case panel.SkillsScannedMsg:
		return m.handleSkillsScanned(msg)

	case panel.MCPScannedMsg:
		return m.handleMCPScanned(msg)

	case panel.PreviewLoadedMsg:
		return m.handlePreviewLoaded(msg)

	case mutationCompletedMsg:
		return m.handleMutationCompleted(msg)

	case reselectMCPMsg:
		return m.handleReselectMCP(msg)

	case reselectSkillMsg:
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
	// Installing during progress-bar phase: installFeature does not consume
	// KeyMsgs (it lets them pass so the user can quit).  All other feature
	// states are consumed by dispatchToFeatures.
	if m.state == stateInstalling && m.install.flow != nil {
		return m.handleInstallingUpdate(msg)
	}
	if m.prompt != nil {
		return m.handlePromptKeys(msg)
	}
	if m.listFilterActive() {
		return m.handleListFilterKeys(msg)
	}
	return m.handleKeyMsg(msg)
}
