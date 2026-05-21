package app

import (
	"path/filepath"
	"strings"

	bubbleslist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/render"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

func (m *Model) setMainListItems(items []bubbleslist.Item) {
	m.SetMainItems(items)
}

func (m *Model) setAgentListItems(items []bubbleslist.Item) {
	m.SetOverlayItems(items)
}

func (m *Model) selectedListItem() (panel.Item, bool) {
	return m.SelectedItem()
}

func (m *Model) syncSelectionPreview() tea.Cmd {
	selected, ok := m.selectedListItem()
	if !ok {
		m.Preview.SetContent(m.styles.EmptyPreview.Render("No selection"))
		return nil
	}

	if selected.Kind == panel.ItemCommand {
		m.PreviewBody = service.RenderCommandPreview(
			selected.Command.Name,
			selected.Command.Usage,
			selected.Command.Summary,
			selected.Command.Implemented,
		)
		m.Preview.SetContent(m.PreviewBody)
		return nil
	}

	width := m.PreviewWidth(max(40, m.width/2))
	return panel.SyncPreviewCmd(m.activePanel(), selected, width, &m.PreviewGen)
}

func (m *Model) previewFileCmd(path string) tea.Cmd {
	width := m.PreviewWidth(max(40, m.width/2))
	m.PreviewGen++
	gen := m.PreviewGen
	return func() tea.Msg {
		dummy := skilldomain.Skill{
			BaseExtension: extension.BaseExtension{
				Name:       filepath.Base(path),
				ConfigPath: path,
			},
		}
		md, err := service.PreviewMarkdown(dummy)
		if err != nil {
			return panel.PreviewLoadedMsg{Tab: m.activeTab, Err: err, Gen: gen}
		}
		content, err := render.Markdown(md, width)
		return panel.PreviewLoadedMsg{Tab: m.activeTab, Content: content, Err: err, Gen: gen}
	}
}

func (m *Model) selectSkillByName(name string) bool {
	skill, ok := m.findSkillByName(name)
	if !ok {
		return false
	}

	m.transitionTo(stateListing)
	m.refreshActiveList()
	for idx, item := range m.Main.Items() {
		li, ok := item.(panel.Item)
		if ok && li.Kind == panel.ItemSkill && strings.EqualFold(li.Skill.GetName(), skill.GetName()) {
			m.Main.Select(idx)
			break
		}
	}
	return true
}

func (m *Model) selectMCPByName(name string) bool {
	m.refreshActiveList()
	for idx, item := range m.Main.Items() {
		li, ok := item.(panel.Item)
		if !ok || li.Kind != panel.ItemMCP {
			continue
		}
		if strings.EqualFold(li.MCPKey, name) ||
			strings.EqualFold(li.MCP.GetName(), name) ||
			strings.EqualFold(li.MCP.ConfigKey, name) {
			m.Main.Select(idx)
			return true
		}
	}
	return false
}
