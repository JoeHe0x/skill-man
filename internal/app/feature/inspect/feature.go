package inspect

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	stateinspect "github.com/JoeHe0x/skill-man/internal/app/state/inspect"
)

// Feature owns skill file-tree inspect (enter + key routing).
type Feature struct {
	host Host
}

// New returns an inspect feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string { return "inspect" }

func (f *Feature) Active() bool { return f.host.IsInspecting() }

func (f *Feature) Init() tea.Cmd { return nil }

func (f *Feature) View(width, height int) string { return "" }

// EnterFromItem opens inspect for the given list item.
func (f *Feature) EnterFromItem(item panel.Item) (tea.Model, tea.Cmd) {
	return EnterFromItem(f.host, item)
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		_, cmd := stateinspect.HandleKeys(f.host, key)
		return cmd, true
	}
	return nil, false
}
