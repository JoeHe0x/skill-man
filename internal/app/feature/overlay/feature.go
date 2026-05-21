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

// AgentFilterHost exposes agent filter overlay needs from the app Model.
type AgentFilterHost interface {
	IsFilteringAgent() bool
	HandleAgentFilterUpdate(tea.Msg) (tea.Model, tea.Cmd)
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

// AgentFilterFeature routes messages while filtering agents.
type AgentFilterFeature struct {
	host AgentFilterHost
}

func NewAgentFilter(host AgentFilterHost) *AgentFilterFeature {
	return &AgentFilterFeature{host: host}
}

func (f *AgentFilterFeature) Name() string                  { return "agentFilter" }
func (f *AgentFilterFeature) Active() bool                  { return f.host.IsFilteringAgent() }
func (f *AgentFilterFeature) Init() tea.Cmd                 { return nil }
func (f *AgentFilterFeature) View(width, height int) string { return "" }

func (f *AgentFilterFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	_, cmd := f.host.HandleAgentFilterUpdate(msg)
	return cmd, true
}
