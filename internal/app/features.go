package app

import tea "github.com/charmbracelet/bubbletea"

type inspectFeature struct {
	host inspectHost
}

func (f *inspectFeature) Name() string                  { return "inspect" }
func (f *inspectFeature) Active() bool                  { return f.host.IsInspecting() }
func (f *inspectFeature) Init() tea.Cmd                 { return nil }
func (f *inspectFeature) View(width, height int) string { return "" }

func (f *inspectFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		_, cmd := f.host.HandleInspectingKeys(key)
		return cmd, true
	}
	return nil, false
}

type agentFilterFeature struct {
	host agentFilterHost
}

func (f *agentFilterFeature) Name() string                  { return "agentFilter" }
func (f *agentFilterFeature) Active() bool                  { return f.host.IsFilteringAgent() }
func (f *agentFilterFeature) Init() tea.Cmd                 { return nil }
func (f *agentFilterFeature) View(width, height int) string { return "" }

func (f *agentFilterFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	_, cmd := f.host.HandleAgentFilterUpdate(msg)
	return cmd, true
}
