package app

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type itemDelegate struct {
	styles styles
	height int
}

func newItemDelegate(styles styles) *itemDelegate {
	return &itemDelegate{styles: styles, height: 3}
}

func (d *itemDelegate) SetHeight(h int) {
	if h < 3 {
		h = 3
	}
	d.height = h
}

func (d *itemDelegate) Height() int {
	if d.height < 3 {
		return 3
	}
	return d.height
}

func (d *itemDelegate) Spacing() int                            { return 0 }
func (d *itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(listItem)
	if !ok {
		return
	}

	width := max(8, m.Width())
	selected := index == m.Index()
	prefix := "  "
	titleStyle := d.styles.itemTitle
	if selected {
		prefix = "› "
		titleStyle = d.styles.itemSelected
	}

	title := truncate(prefix+entry.title, width)
	meta := truncate("  "+entry.meta, width)

	var lines []string
	lines = append(lines, titleStyle.Render(title))

	if len(entry.detailLines) > 0 {
		for _, detail := range entry.detailLines {
			lines = append(lines, d.styles.itemBinding.Render(truncate(detail, width)))
		}
		desc := truncate("  "+entry.desc, width)
		lines = append(lines, d.styles.itemDesc.Render(desc))
	} else {
		desc := truncate("  "+entry.desc, width)
		lines = append(lines, d.styles.itemDesc.Render(desc))
	}

	lines = append(lines, d.styles.itemMeta.Render(meta))

	rendered := lipgloss.JoinVertical(lipgloss.Left, lines...)
	fmt.Fprint(w, rendered)
}

func listHeightForItems(items []list.Item) int {
	h := 3
	for _, it := range items {
		li, ok := it.(listItem)
		if !ok || len(li.detailLines) == 0 {
			continue
		}
		need := 2 + len(li.detailLines)
		if need > h {
			h = need
		}
	}
	return h
}
