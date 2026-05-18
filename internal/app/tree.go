package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type treeItem struct {
	path     string
	name     string
	isDir    bool
	depth    int
	isLast   bool // Not strictly necessary for simple indentation, but good for drawing lines
	selected bool // updated by delegate
}

func (i treeItem) FilterValue() string { return i.name }
func (i treeItem) Title() string       { return i.name }
func (i treeItem) Description() string { return "" }

type treeDelegate struct {
	styles styles
}

func (d treeDelegate) Height() int                             { return 1 }
func (d treeDelegate) Spacing() int                            { return 0 }
func (d treeDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d treeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	node, ok := item.(treeItem)
	if !ok {
		return
	}

	indent := strings.Repeat("  ", node.depth)
	icon := "📄 "
	if node.isDir {
		icon = "📁 "
	}

	text := indent + icon + node.name

	if index == m.Index() {
		fmt.Fprint(w, d.styles.itemSelected.Render("› "+text))
	} else {
		fmt.Fprint(w, d.styles.itemTitle.Render("  "+text))
	}
}

type fileTreeModel struct {
	list     list.Model
	rootPath string
	expanded map[string]bool
	styles   styles
}

func newFileTreeModel(styles styles) fileTreeModel {
	l := list.New([]list.Item{}, treeDelegate{styles: styles}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return fileTreeModel{
		list:     l,
		expanded: make(map[string]bool),
		styles:   styles,
	}
}

func (m *fileTreeModel) setRoot(root string) {
	m.rootPath = root
	m.expanded = make(map[string]bool)
	m.expanded[root] = true // root is always expanded
	m.refreshItems()
}

func (m *fileTreeModel) refreshItems() {
	if m.rootPath == "" {
		return
	}

	var items []list.Item

	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.Name() == ".git" {
				continue
			}
			path := filepath.Join(dir, e.Name())
			isDir := e.IsDir()

			items = append(items, treeItem{
				path:  path,
				name:  e.Name(),
				isDir: isDir,
				depth: depth,
			})

			if isDir && m.expanded[path] {
				walk(path, depth+1)
			}
		}
	}

	walk(m.rootPath, 0)
	m.list.SetItems(items)
}

func (m *fileTreeModel) Update(msg tea.Msg) (fileTreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "space", "right", "left"))):
			item := m.list.SelectedItem()
			if node, ok := item.(treeItem); ok && node.isDir {
				// Toggle
				if key.Matches(msg, key.NewBinding(key.WithKeys("left"))) {
					m.expanded[node.path] = false
				} else if key.Matches(msg, key.NewBinding(key.WithKeys("right"))) {
					m.expanded[node.path] = true
				} else {
					m.expanded[node.path] = !m.expanded[node.path]
				}
				m.refreshItems()
				return *m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

func (m *fileTreeModel) View() string {
	return m.list.View()
}

func (m *fileTreeModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

func (m *fileTreeModel) SelectedItem() treeItem {
	if i, ok := m.list.SelectedItem().(treeItem); ok {
		return i
	}
	return treeItem{}
}
