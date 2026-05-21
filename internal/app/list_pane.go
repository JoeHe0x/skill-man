package app

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// listPane owns the main list, overlay list (bind / agent filter), preview, and inspect tree.
type listPane struct {
	list              list.Model
	listDelegate      *itemDelegate
	agentList         list.Model // overlay: bind flow or agent filter
	agentListDelegate *itemDelegate
	preview           viewport.Model
	previewBody       string
	previewGen        int
	tree              fileTreeModel
}

func newListPane(styles theme.Styles) listPane {
	mainDelegate := newItemDelegate(styles)
	skillList := list.New([]list.Item{}, mainDelegate, 0, 0)
	skillList.SetShowTitle(false)
	skillList.SetShowStatusBar(false)
	skillList.SetFilteringEnabled(false)
	skillList.SetShowHelp(false)
	skillList.DisableQuitKeybindings()

	agentDelegate := newItemDelegate(styles)
	agentList := list.New([]list.Item{}, agentDelegate, 0, 0)
	agentList.SetShowTitle(false)
	agentList.SetShowStatusBar(false)
	agentList.SetShowPagination(false)
	agentList.SetFilteringEnabled(false)
	agentList.SetShowHelp(false)
	agentList.DisableQuitKeybindings()

	preview := viewport.New(0, 0)
	preview.SetContent(welcomePreview)

	return listPane{
		list:              skillList,
		listDelegate:      mainDelegate,
		agentList:         agentList,
		agentListDelegate: agentDelegate,
		preview:           preview,
		previewBody:       welcomePreview,
		tree:              newFileTreeModel(styles),
	}
}

func (p *listPane) configureKeys() {
	p.list.KeyMap.CursorUp = keys.Up
	p.list.KeyMap.CursorDown = keys.Down
	p.list.KeyMap.NextPage = keys.PgDown
	p.list.KeyMap.PrevPage = keys.PgUp
	p.agentList.KeyMap.CursorUp = keys.Up
	p.agentList.KeyMap.CursorDown = keys.Down
	p.agentList.KeyMap.NextPage = keys.PgDown
	p.agentList.KeyMap.PrevPage = keys.PgUp
	p.preview.KeyMap.PageUp = keys.PgUp
	p.preview.KeyMap.PageDown = keys.PgDown
}

func (p *listPane) applyTheme(styles theme.Styles) {
	if p.listDelegate != nil {
		p.listDelegate.styles = styles
	}
	if p.agentListDelegate != nil {
		p.agentListDelegate.styles = styles
	}
	p.tree.setStyles(styles)
	p.list.SetDelegate(p.listDelegate)
	p.agentList.SetDelegate(p.agentListDelegate)
}

func (p *listPane) resize(lw, lh, rw, rh int) {
	p.list.SetSize(lw, lh)
	p.agentList.SetSize(lw, lh)
	p.preview.Width = rw
	p.preview.Height = rh
}

func (p *listPane) setMainItems(items []list.Item) {
	p.listDelegate.SetHeight(listHeightForItems(items))
	p.list.SetItems(items)
	p.list.SetShowStatusBar(visiblePanelListCount(items) > 0)
}

func (p *listPane) setOverlayItems(items []list.Item) {
	p.agentListDelegate.SetHeight(listHeightForItems(items))
	p.agentList.SetItems(items)
}

func (p *listPane) selectedItem() (panel.Item, bool) {
	item := p.list.SelectedItem()
	if item == nil {
		return panel.Item{}, false
	}
	li, ok := item.(panel.Item)
	return li, ok
}

func (p *listPane) previewWidth(fallback int) int {
	w := p.preview.Width
	if w == 0 {
		return fallback
	}
	return w
}

func (m *Model) setMainListItems(items []list.Item) {
	m.listPane.setMainItems(items)
}

func (m *Model) setAgentListItems(items []list.Item) {
	m.listPane.setOverlayItems(items)
}

func (m *Model) selectedListItem() (panel.Item, bool) {
	return m.listPane.selectedItem()
}

func (m *Model) syncSelectionPreview() tea.Cmd {
	selected, ok := m.selectedListItem()
	if !ok {
		m.preview.SetContent(m.styles.EmptyPreview.Render("No selection"))
		return nil
	}

	if selected.Kind == panel.ItemCommand {
		m.previewBody = service.RenderCommandPreview(
			selected.Command.Name,
			selected.Command.Usage,
			selected.Command.Summary,
			selected.Command.Implemented,
		)
		m.preview.SetContent(m.previewBody)
		return nil
	}

	width := m.listPane.previewWidth(max(40, m.width/2))
	return m.activePanel().SyncPreview(selected, width, &m.previewGen)
}

func (m *Model) previewFileCmd(path string) tea.Cmd {
	width := m.listPane.previewWidth(max(40, m.width/2))
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

func (m *Model) selectSkillByName(name string) bool {
	skill, ok := m.findSkillByName(name)
	if !ok {
		return false
	}

	m.transitionTo(stateListing)
	m.refreshActiveList()
	for idx, item := range m.list.Items() {
		li, ok := item.(panel.Item)
		if ok && li.Kind == panel.ItemSkill && strings.EqualFold(li.Skill.GetName(), skill.GetName()) {
			m.list.Select(idx)
			break
		}
	}
	return true
}

func (m *Model) selectMCPByName(name string) bool {
	m.refreshActiveList()
	for idx, item := range m.list.Items() {
		li, ok := item.(panel.Item)
		if !ok || li.Kind != panel.ItemMCP {
			continue
		}
		if strings.EqualFold(li.MCPKey, name) ||
			strings.EqualFold(li.MCP.GetName(), name) ||
			strings.EqualFold(li.MCP.ConfigKey, name) {
			m.list.Select(idx)
			return true
		}
	}
	return false
}
