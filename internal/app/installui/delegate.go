package installui

import (
	"fmt"
	"io"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type itemDelegate struct {
	styles theme.Styles
	height int
}

func newItemDelegate(styles theme.Styles) *itemDelegate {
	return &itemDelegate{styles: styles, height: 1}
}

func (d *itemDelegate) SetHeight(h int) {
	if h < 1 {
		h = 1
	}
	d.height = h
}

func (d *itemDelegate) Height() int {
	if d.height < 1 {
		return 1
	}
	return d.height
}

func (d *itemDelegate) Spacing() int                            { return 0 }
func (d *itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(Row)
	if !ok {
		return
	}
	selected := index == m.Index()
	prefix := "  "
	titleStyle := d.styles.ItemTitle
	if selected {
		prefix = "› "
		titleStyle = d.styles.ItemSelected
	}
	titlePart := titleStyle.Render(entry.Title)
	descPart := d.styles.ItemDesc.Render(entry.Desc)
	line := prefix + titlePart + "  " + descPart
	if entry.Meta != "" {
		line += "  " + d.styles.ItemMeta.Render(entry.Meta)
	}
	fmt.Fprint(w, line)
}

func truncate(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	if limit == 1 {
		return "…"
	}
	return string(runes[:limit-1]) + "…"
}
