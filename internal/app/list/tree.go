package list

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// TreeNode is one row in the inspect file tree.
type TreeNode struct {
	Path  string
	Name  string
	IsDir bool
	Depth int
}

type treeItem struct {
	TreeNode
}

func (i treeItem) FilterValue() string { return i.Name }
func (i treeItem) Title() string       { return i.Name }
func (i treeItem) Description() string { return "" }

type treeDelegate struct {
	styles theme.Styles
}

func (d treeDelegate) Height() int                             { return 1 }
func (d treeDelegate) Spacing() int                            { return 0 }
func (d treeDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d treeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	node, ok := item.(treeItem)
	if !ok {
		return
	}

	indent := strings.Repeat("  ", node.Depth)
	icon := "  "
	if node.IsDir {
		icon = "> "
	}

	text := indent + icon + node.Name

	if index == m.Index() {
		fmt.Fprint(w, d.styles.ItemSelected.Render("› "+text))
	} else {
		fmt.Fprint(w, d.styles.ItemTitle.Render("  "+text))
	}
}

// FileTree is the inspect-mode file browser.
type FileTree struct {
	list     list.Model
	rootPath string
	expanded map[string]bool
	styles   theme.Styles
}

// NewFileTree returns an empty file tree list.
func NewFileTree(styles theme.Styles) FileTree {
	l := list.New([]list.Item{}, treeDelegate{styles: styles}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return FileTree{
		list:     l,
		expanded: make(map[string]bool),
		styles:   styles,
	}
}

func (m *FileTree) SetStyles(s theme.Styles) {
	m.styles = s
	m.list.SetDelegate(treeDelegate{styles: s})
}

func (m *FileTree) SetRoot(root string) {
	m.rootPath = root
	m.expanded = make(map[string]bool)
	m.expanded[root] = true
	m.refreshItems()
}

func (m *FileTree) refreshItems() {
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

			items = append(items, treeItem{TreeNode: TreeNode{
				Path:  path,
				Name:  e.Name(),
				IsDir: isDir,
				Depth: depth,
			}})

			if isDir && m.expanded[path] {
				walk(path, depth+1)
			}
		}
	}

	walk(m.rootPath, 0)
	m.list.SetItems(items)
}

func (m *FileTree) Update(msg tea.Msg) (FileTree, tea.Cmd) {
	keys := uikeys.Default
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter, keys.Toggle, keys.Left, keys.Right):
			item := m.list.SelectedItem()
			if node, ok := item.(treeItem); ok && node.IsDir {
				switch {
				case key.Matches(msg, keys.Left):
					m.expanded[node.Path] = false
				case key.Matches(msg, keys.Right):
					m.expanded[node.Path] = true
				default:
					m.expanded[node.Path] = !m.expanded[node.Path]
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

func (m *FileTree) View() string {
	return m.list.View()
}

func (m *FileTree) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// SelectedNode returns the currently selected tree row.
func (m *FileTree) SelectedNode() TreeNode {
	if i, ok := m.list.SelectedItem().(treeItem); ok {
		return i.TreeNode
	}
	return TreeNode{}
}
