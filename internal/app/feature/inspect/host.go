package inspect

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/list"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

// Host exposes skill inspect flow needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	IsInspecting() bool
	ActivePanel() panel.Panel
	PreviewWidth() int
	AppWidth() int
	PreviewGenPtr() *int
	TransitionTo(session.State) bool
	SetFooterContext(string)
	SetTreeRoot(path string)
	TreeSelected() list.TreeNode
	FlashFooter(string) tea.Cmd
	SyncSelectionPreview() tea.Cmd
	PreviewUpdate(tea.KeyMsg) tea.Cmd
	TreeUpdate(tea.Msg) (list.FileTree, tea.Cmd)
	PreviewFileCmd(path string) tea.Cmd
}
