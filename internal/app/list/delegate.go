package list

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/strutil"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

// Delegate renders panel list rows.
type Delegate struct {
	styles theme.Styles
	height int
}

// NewDelegate returns a list item delegate with default row height.
func NewDelegate(styles theme.Styles) *Delegate {
	return &Delegate{styles: styles, height: 3}
}

// Styles returns the delegate's theme (for tests).
func (d *Delegate) Styles() theme.Styles { return d.styles }

func (d *Delegate) SetHeight(h int) {
	if h < 1 {
		h = 1
	}
	d.height = h
}

func (d *Delegate) Height() int {
	if d.height < 1 {
		return 1
	}
	return d.height
}

func (d *Delegate) Spacing() int                            { return 0 }
func (d *Delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d *Delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
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

	title := strutil.Truncate(prefix+entry.Title, width)
	meta := strutil.Truncate("  "+entry.Meta, width)

	var lines []string
	lines = append(lines, titleStyle.Render(title))

	if len(entry.DetailLines) > 0 {
		for _, detail := range entry.DetailLines {
			lines = append(lines, d.styles.ItemBinding.Render(strutil.Truncate(detail, width)))
		}
		desc := strutil.Truncate("  "+entry.Desc, width)
		lines = append(lines, d.styles.ItemDesc.Render(desc))
	} else if entry.Kind == panel.ItemMessage {
		titleStyle := d.styles.ItemTitle
		if selected {
			titleStyle = d.styles.ItemSelected
		}

		descStyle := d.styles.ItemDesc
		descText := entry.Desc

		titlePart := titleStyle.Render(entry.Title)
		descPart := descStyle.Render(descText)

		lines[0] = prefix + titlePart + "  " + descPart
		lines = lines[:1]

		rendered := lipgloss.JoinVertical(lipgloss.Left, lines...)
		fmt.Fprint(w, rendered)
		return
	} else {
		desc := strutil.Truncate("  "+entry.Desc, width)
		lines = append(lines, d.styles.ItemDesc.Render(desc))
	}

	lines = append(lines, d.styles.ItemMeta.Render(meta))

	rendered := lipgloss.JoinVertical(lipgloss.Left, lines...)
	fmt.Fprint(w, rendered)
}
