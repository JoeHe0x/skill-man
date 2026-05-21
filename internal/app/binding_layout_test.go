package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestBindingMainAreaFitsAndUsesCompactList(t *testing.T) {
	m := New("/mnt/c/Code/skill-man", "/home/joe")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = mustModel(t, updated)

	m.state = stateBindingAgent
	m.bind.session.agents = m.binder.NewMCPChoices(nil)
	m.setAgentListItems(bindChoicesToListItems(m.bind.session.agents, m.cwd, m.home))

	_, wantMainH := m.mainAreaSize()
	main := m.renderMainAreaSized(wantMainH)
	if h := lipgloss.Height(main); h > wantMainH {
		t.Fatalf("bind main area %d > budget %d", h, wantMainH)
	}
	if m.agentList.ShowPagination() {
		t.Fatal("agent list pagination should be disabled in bind UI")
	}
	if m.agentListDelegate.Height() != 1 {
		t.Fatalf("delegate height %d want 1 for bind rows", m.agentListDelegate.Height())
	}
}
