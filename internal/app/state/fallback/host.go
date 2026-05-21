package fallback

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

// Host exposes non-key message handlers for the app Model.
type Host interface {
	TeaModel() tea.Model

	ApplyMutationResult(uimsg.MutationCompleted) (tea.Model, tea.Cmd)
	SetWindowSize(int, int)
	ResizeComponents()
	ResizePaletteInput()
	SyncSelectionPreview() tea.Cmd
	HandleMouse(tea.MouseMsg) (tea.Model, tea.Cmd)

	PreviewGeneration() int
	ActiveTab() panel.Tab
	SetPreviewError(string)
	SetPreviewBody(string)
	SetPreviewContent(string)
	ClearStaleLoadingIfIdle()

	SpinnerTick(spinner.TickMsg) tea.Cmd
	InstallWizardSearching() bool
	InstallHandleUIMsg(tea.Msg) tea.Cmd
	InstallHandleBackgroundFrame(progress.FrameMsg) (tea.Cmd, bool)

	SelectMCPByName(string) bool
	SelectSkillByName(string) bool
	FlashFooter(string) tea.Cmd

	MainFallthrough(tea.Msg) (tea.Cmd, tea.Cmd)
}
