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
}

func newItemDelegate(styles styles) itemDelegate {
	return itemDelegate{styles: styles}
}

func (d itemDelegate) Height() int                             { return 3 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
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
	desc := truncate("  "+entry.desc, width)
	meta := truncate("  "+entry.meta, width)

	rendered := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		d.styles.itemDesc.Render(desc),
		d.styles.itemMeta.Render(meta),
	)
	fmt.Fprint(w, rendered)
}
