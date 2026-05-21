package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/list"
	statenspect "github.com/JoeHe0x/skill-man/internal/app/state/inspect"
)

func (m *Model) handleInspectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statenspect.HandleKeys(m, msg)
}

func (m *Model) SyncSelectionPreview() tea.Cmd {
	return m.syncSelectionPreview()
}

func (m *Model) PreviewFileCmd(path string) tea.Cmd {
	return m.previewFileCmd(path)
}

func (m *Model) PreviewUpdate(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.Preview, cmd = m.Preview.Update(msg)
	return cmd
}

func (m *Model) TreeUpdate(msg tea.Msg) (list.FileTree, tea.Cmd) {
	var cmd tea.Cmd
	m.Tree, cmd = m.Tree.Update(msg)
	return m.Tree, cmd
}

func (m *Model) TreeSelected() list.TreeNode {
	return m.Tree.SelectedNode()
}

var _ statenspect.Host = (*Model)(nil)
