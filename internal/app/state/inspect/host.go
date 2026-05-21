package inspect

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/list"
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

// Host exposes inspect-mode needs from the app Model.
type Host interface {
	TeaModel() tea.Model
	TransitionTo(session.State) bool
	FlashFooter(string) tea.Cmd
	SyncSelectionPreview() tea.Cmd
	PreviewUpdate(tea.KeyMsg) tea.Cmd
	TreeUpdate(tea.Msg) (list.FileTree, tea.Cmd)
	TreeSelected() list.TreeNode
	PreviewFileCmd(path string) tea.Cmd
}
