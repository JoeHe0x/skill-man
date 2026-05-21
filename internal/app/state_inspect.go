package app

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

func (m *Model) handleInspectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.transitionTo(stateListing)
		return m, tea.Batch(m.flashFooter("Returned to skill list"), m.syncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd
	}

	oldSelected := m.tree.SelectedItem()
	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	newSelected := m.tree.SelectedItem()

	if newSelected.path != "" && newSelected.path != oldSelected.path && !newSelected.isDir {
		return m, tea.Batch(cmd, m.previewFileCmd(newSelected.path))
	}

	return m, cmd
}

func (m *Model) previewFileCmd(path string) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	m.previewGen++
	gen := m.previewGen
	return func() tea.Msg {
		dummy := skilldomain.Skill{
			BaseExtension: extension.BaseExtension{
				Name:       filepath.Base(path),
				ConfigPath: path,
			},
		}
		content, err := service.RenderSkillPreview(dummy, width)
		return panel.PreviewLoadedMsg{Tab: m.activeTab, Content: content, Err: err, Gen: gen}
	}
}
