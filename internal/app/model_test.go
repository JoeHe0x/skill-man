package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"skill-man/internal/domain/skill"
)

func mustModel(t *testing.T, m tea.Model) *Model {
	t.Helper()
	switch mm := m.(type) {
	case *Model:
		return mm
	default:
		t.Fatalf("expected model or *model, got %T", m)
		return nil
	}
}

func TestAgentCyclingAllToFirst(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.skills = []skill.Skill{{Name: "test-skill"}}

	if m.agentDisplay() != "all" {
		t.Fatalf("expected initial filter 'all', got %q", m.agentDisplay())
	}

	updated, _ := m.handleCycleAgent()
	m2 := mustModel(t, updated)
	if m2.agentDisplay() == "all" {
		t.Fatal("expected agent to cycle away from 'all'")
	}
}

func TestAgentCyclingBackToAll(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.skills = []skill.Skill{{Name: "test-skill"}}

	cycles := len(m.allAgents) + 1
	for i := 0; i < cycles; i++ {
		updated, _ := m.handleCycleAgent()
		m = mustModel(t, updated)
	}

	if m.agentDisplay() != "all" {
		t.Fatalf("expected to cycle back to 'all', got %q", m.agentDisplay())
	}
}

func TestPromptLifecycle(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	if m.prompt != nil {
		t.Fatal("expected nil prompt initially")
	}

	updated, cmd := m.showFindPrompt()
	m2 := mustModel(t, updated)
	if m2.prompt == nil {
		t.Fatal("expected prompt after showFindPrompt")
	}
	if m2.prompt.label != "Find" {
		t.Fatalf("expected prompt label 'Find', got %q", m2.prompt.label)
	}
	if cmd == nil {
		t.Fatal("expected blink cmd from showPrompt")
	}

	// Esc cancels prompt
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m2.Update(msg)
	m3 := mustModel(t, updated)
	if m3.prompt != nil {
		t.Fatal("expected prompt to be nil after esc")
	}
}

func TestPromptEnterExecutes(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.showFindPrompt()
	m2 := mustModel(t, updated)

	// Type "my-query" then press Enter
	for _, r := range "my-query" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updated, _ := m2.Update(msg)
		m2 = mustModel(t, updated)
	}
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m2.Update(msg)
	m3 := mustModel(t, updated)

	if m3.prompt != nil {
		t.Fatal("expected prompt to be nil after enter")
	}
	if m3.state != stateSearching {
		t.Fatalf("expected stateSearching after find, got %v", m3.state)
	}
}
