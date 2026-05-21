package filter

import (
	tea "github.com/charmbracelet/bubbletea"

	statefiltering "github.com/JoeHe0x/skill-man/internal/app/state/filtering"
)

// Feature owns the agent-filter overlay (open, render, key routing).
type Feature struct {
	host Host
}

// New returns an agent-filter feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string { return "agentFilter" }

func (f *Feature) Active() bool { return f.host.IsFilteringAgent() }

func (f *Feature) Init() tea.Cmd { return nil }

func (f *Feature) View(width, height int) string { return "" }

// Open enters agent-filter mode.
func (f *Feature) Open() (tea.Model, tea.Cmd) {
	return Open(f.host)
}

// RenderMainOverlay renders the filter dialog in the main area.
func (f *Feature) RenderMainOverlay() string {
	return RenderMainOverlay(f.host)
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	_, cmd := statefiltering.HandleUpdate(f.host, msg)
	return cmd, true
}
