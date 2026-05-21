package list

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// Pane owns the main list, overlay list (bind / agent filter), preview, and inspect tree.
type Pane struct {
	Main        list.Model
	MainDel     *Delegate
	Agent       list.Model
	AgentDel    *Delegate
	Preview     viewport.Model
	PreviewBody string
	PreviewGen  int
	Tree        FileTree
}

// NewPane builds list/preview/tree widgets with default welcome content.
func NewPane(styles theme.Styles) Pane {
	mainDel := NewDelegate(styles)
	main := list.New([]list.Item{}, mainDel, 0, 0)
	main.SetShowTitle(false)
	main.SetShowStatusBar(false)
	main.SetFilteringEnabled(false)
	main.SetShowHelp(false)
	main.DisableQuitKeybindings()

	agentDel := NewDelegate(styles)
	agent := list.New([]list.Item{}, agentDel, 0, 0)
	agent.SetShowTitle(false)
	agent.SetShowStatusBar(false)
	agent.SetShowPagination(false)
	agent.SetFilteringEnabled(false)
	agent.SetShowHelp(false)
	agent.DisableQuitKeybindings()

	preview := viewport.New(0, 0)
	preview.SetContent(WelcomePreview)

	return Pane{
		Main:        main,
		MainDel:     mainDel,
		Agent:       agent,
		AgentDel:    agentDel,
		Preview:     preview,
		PreviewBody: WelcomePreview,
		Tree:        NewFileTree(styles),
	}
}

// ConfigureKeys wires bubbletea keymaps from uikeys.
func (p *Pane) ConfigureKeys() {
	keys := uikeys.Default
	p.Main.KeyMap.CursorUp = keys.Up
	p.Main.KeyMap.CursorDown = keys.Down
	p.Main.KeyMap.NextPage = keys.PgDown
	p.Main.KeyMap.PrevPage = keys.PgUp
	p.Agent.KeyMap.CursorUp = keys.Up
	p.Agent.KeyMap.CursorDown = keys.Down
	p.Agent.KeyMap.NextPage = keys.PgDown
	p.Agent.KeyMap.PrevPage = keys.PgUp
	p.Preview.KeyMap.PageUp = keys.PgUp
	p.Preview.KeyMap.PageDown = keys.PgDown
}

func (p *Pane) ApplyTheme(styles theme.Styles) {
	if p.MainDel != nil {
		p.MainDel.styles = styles
	}
	if p.AgentDel != nil {
		p.AgentDel.styles = styles
	}
	p.Tree.SetStyles(styles)
	p.Main.SetDelegate(p.MainDel)
	p.Agent.SetDelegate(p.AgentDel)
}

func (p *Pane) Resize(lw, lh, rw, rh int) {
	p.Main.SetSize(lw, lh)
	p.Agent.SetSize(lw, lh)
	p.Preview.Width = rw
	p.Preview.Height = rh
}

func (p *Pane) SetMainItems(items []list.Item) {
	p.MainDel.SetHeight(HeightForItems(items))
	p.Main.SetItems(items)
	p.Main.SetShowStatusBar(visibleItemCount(items) > 0)
}

func (p *Pane) SetOverlayItems(items []list.Item) {
	p.AgentDel.SetHeight(HeightForItems(items))
	p.Agent.SetItems(items)
}

func (p *Pane) SelectedItem() (panel.Item, bool) {
	item := p.Main.SelectedItem()
	if item == nil {
		return panel.Item{}, false
	}
	li, ok := item.(panel.Item)
	return li, ok
}

func visibleItemCount(items []list.Item) int {
	if len(items) == 0 {
		return 0
	}
	panelItems := make([]panel.Item, 0, len(items))
	for _, it := range items {
		if pi, ok := it.(panel.Item); ok {
			panelItems = append(panelItems, pi)
		}
	}
	return panel.VisibleListCount(panelItems)
}

// PreviewWidth returns preview width or fallback when unset.
func (p *Pane) PreviewWidth(fallback int) int {
	w := p.Preview.Width
	if w == 0 {
		return fallback
	}
	return w
}
