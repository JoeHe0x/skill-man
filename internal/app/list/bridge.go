package list

import (
	"path/filepath"
	"strings"

	bubbleslist "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/render"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// BridgeHost exposes list/preview selection helpers on the app Model.
type BridgeHost interface {
	ListPane() *Pane
	ActiveTab() panel.Tab
	ActivePanel() panel.Panel
	AppWidth() int
	Styles() theme.Styles
	TransitionTo(session.State) bool
	RefreshActiveList()
	FindSkillByName(string) (*skilldomain.Skill, bool)
}

// PanelToListItems converts panel rows for bubbles/list.
func PanelToListItems(items []panel.Item) []bubbleslist.Item {
	out := make([]bubbleslist.Item, len(items))
	for i := range items {
		out[i] = items[i]
	}
	return out
}

// SetMainItemsFromPanel converts panel rows and applies them to the main list.
func SetMainItemsFromPanel(h BridgeHost, items []panel.Item) {
	h.ListPane().SetMainItems(PanelToListItems(items))
}

// SyncSelectionPreview refreshes the preview pane for the current main-list selection.
func SyncSelectionPreview(h BridgeHost) tea.Cmd {
	p := h.ListPane()
	selected, ok := p.SelectedItem()
	if !ok {
		p.Preview.SetContent(h.Styles().EmptyPreview.Render("No selection"))
		return nil
	}

	if selected.Kind == panel.ItemCommand {
		p.PreviewBody = service.RenderCommandPreview(
			selected.Command.Name,
			selected.Command.Usage,
			selected.Command.Summary,
			selected.Command.Implemented,
		)
		p.Preview.SetContent(p.PreviewBody)
		return nil
	}

	width := p.PreviewWidth(max(40, h.AppWidth()/2))
	return panel.SyncPreviewCmd(h.ActivePanel(), selected, width, &p.PreviewGen)
}

// PreviewFileCmd loads markdown preview for a file path in the inspect tree.
func PreviewFileCmd(h BridgeHost, path string) tea.Cmd {
	p := h.ListPane()
	width := p.PreviewWidth(max(40, h.AppWidth()/2))
	p.PreviewGen++
	gen := p.PreviewGen
	tab := h.ActiveTab()
	return func() tea.Msg {
		dummy := skilldomain.Skill{
			BaseExtension: extension.BaseExtension{
				Name:       filepath.Base(path),
				ConfigPath: path,
			},
		}
		md, err := service.PreviewMarkdown(dummy)
		if err != nil {
			return panel.PreviewLoadedMsg{Tab: tab, Err: err, Gen: gen}
		}
		content, err := render.Markdown(md, width)
		return panel.PreviewLoadedMsg{Tab: tab, Content: content, Err: err, Gen: gen}
	}
}

// SelectSkillByName switches to listing and selects the named skill row.
func SelectSkillByName(h BridgeHost, name string) bool {
	skill, ok := h.FindSkillByName(name)
	if !ok {
		return false
	}

	h.TransitionTo(session.Listing)
	h.RefreshActiveList()
	p := h.ListPane()
	for idx, item := range p.Main.Items() {
		li, ok := item.(panel.Item)
		if ok && li.Kind == panel.ItemSkill && strings.EqualFold(li.Skill.GetName(), skill.GetName()) {
			p.Main.Select(idx)
			break
		}
	}
	return true
}

// SelectMCPByName selects the named MCP row on the active list.
func SelectMCPByName(h BridgeHost, name string) bool {
	h.RefreshActiveList()
	p := h.ListPane()
	for idx, item := range p.Main.Items() {
		li, ok := item.(panel.Item)
		if !ok || li.Kind != panel.ItemMCP {
			continue
		}
		if strings.EqualFold(li.MCPKey, name) ||
			strings.EqualFold(li.MCP.GetName(), name) ||
			strings.EqualFold(li.MCP.ConfigKey, name) {
			p.Main.Select(idx)
			return true
		}
	}
	return false
}
