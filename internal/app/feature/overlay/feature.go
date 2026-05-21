package overlay

import (
	tea "github.com/charmbracelet/bubbletea"
)

// InspectHost exposes skill inspect flow needs from the app Model.
type InspectHost interface {
	IsInspecting() bool
	HandleInspectingKeys(tea.KeyMsg) (tea.Model, tea.Cmd)
	TeaModel() tea.Model
}

// InspectFeature routes keys while inspecting a skill tree.
type InspectFeature struct {
	host InspectHost
}

func NewInspect(host InspectHost) *InspectFeature {
	return &InspectFeature{host: host}
}

func (f *InspectFeature) Name() string                  { return "inspect" }
func (f *InspectFeature) Active() bool                  { return f.host.IsInspecting() }
func (f *InspectFeature) Init() tea.Cmd                 { return nil }
func (f *InspectFeature) View(width, height int) string { return "" }

func (f *InspectFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		_, cmd := f.host.HandleInspectingKeys(key)
		return cmd, true
	}
	return nil, false
}
