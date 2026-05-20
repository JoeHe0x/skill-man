package app

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

type itemDelegate struct {
	styles theme.Styles
	height int
}

func newItemDelegate(styles theme.Styles) *itemDelegate {
	return &itemDelegate{styles: styles, height: 3}
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
	entry, ok := item.(panel.Item)
	if !ok {
		return
	}

	width := max(8, m.Width())
	selected := index == m.Index()
	prefix := "  "
	titleStyle := d.styles.ItemTitle
	if selected {
		prefix = "› "
		titleStyle = d.styles.ItemSelected
	}

	title := truncate(prefix+entry.Title, width)
	meta := truncate("  "+entry.Meta, width)

	var lines []string
	lines = append(lines, titleStyle.Render(title))

	if len(entry.DetailLines) > 0 {
		for _, detail := range entry.DetailLines {
			lines = append(lines, d.styles.ItemBinding.Render(truncate(detail, width)))
		}
		desc := truncate("  "+entry.Desc, width)
		lines = append(lines, d.styles.ItemDesc.Render(desc))
	} else if entry.Kind == panel.ItemMessage {
		titleStyle := d.styles.ItemTitle
		if selected {
			titleStyle = d.styles.ItemSelected
		}

		descStyle := d.styles.ItemDesc
		// Strip the extra indent since we are inlining
		descText := entry.Desc

		titlePart := titleStyle.Render(entry.Title)
		descPart := descStyle.Render(descText)

		// Note: truncation on ansi strings might not work correctly if truncate strips ansi.
		// It's safer to truncate before adding color, or trust the container to hide overflow.
		lines[0] = prefix + titlePart + "  " + descPart

		// Ensure only 1 line height for panel.ItemMessage (the inline title/desc)
		lines = lines[:1]

		rendered := lipgloss.JoinVertical(lipgloss.Left, lines...)
		fmt.Fprint(w, rendered)
		return
	} else {
		desc := truncate("  "+entry.Desc, width)
		lines = append(lines, d.styles.ItemDesc.Render(desc))
	}

	lines = append(lines, d.styles.ItemMeta.Render(meta))

	rendered := lipgloss.JoinVertical(lipgloss.Left, lines...)
	fmt.Fprint(w, rendered)
}

func listHeightForItems(items []list.Item) int {
	h := 3

	allMessages := len(items) > 0
	for _, it := range items {
		if li, ok := it.(panel.Item); !ok || li.Kind != panel.ItemMessage {
			allMessages = false
			break
		}
	}

	if allMessages {
		return 1
	}

	for _, it := range items {
		li, ok := it.(panel.Item)
		if !ok || len(li.DetailLines) == 0 {
			continue
		}
		need := 2 + len(li.DetailLines)
		if need > h {
			h = need
		}
	}
	return h
}
