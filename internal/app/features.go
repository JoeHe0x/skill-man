package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// inspectFeature wraps the skill file-tree inspection flow.

type inspectFeature struct {
	m *Model
}

func (f *inspectFeature) Name() string                  { return "inspect" }
func (f *inspectFeature) Active() bool                  { return f.m.state == stateInspecting }
func (f *inspectFeature) Init() tea.Cmd                 { return nil }
func (f *inspectFeature) View(width, height int) string { return "" }

func (f *inspectFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleInspectingKeys(msg)
		return cmd, true
	}
	return nil, false
}

// agentFilterFeature wraps the agent filter dialog.

type agentFilterFeature struct {
	m *Model
}

func (f *agentFilterFeature) Name() string                  { return "agentFilter" }
func (f *agentFilterFeature) Active() bool                  { return f.m.state == stateFilteringAgent }
func (f *agentFilterFeature) Init() tea.Cmd                 { return nil }
func (f *agentFilterFeature) View(width, height int) string { return "" }

func (f *agentFilterFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleAgentFilterUpdate(msg)
		return cmd, true
	}
	return nil, false
}
