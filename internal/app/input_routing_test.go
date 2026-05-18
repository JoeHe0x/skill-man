package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"skill-man/internal/commands"
	"skill-man/internal/domain/skill"
)

func TestCtrlLTriggersList(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()

	updated, _ := m.handleList()
	result := mustModel(t, updated)
	if result.state != stateListing {
		t.Fatalf("expected stateListing, got %v", result.state)
	}
}

func TestCtrlFShowsFindPrompt(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()

	updated, _ := m.showFindPrompt()
	result := mustModel(t, updated)
	if result.prompt == nil {
		t.Fatal("expected prompt after showFindPrompt")
	}
	if result.prompt.label != "Find" {
		t.Fatalf("expected 'Find' label, got %q", result.prompt.label)
	}
}

func TestCtrlDShowsAddPrompt(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.showAddPrompt()
	result := mustModel(t, updated)
	if result.prompt == nil {
		t.Fatal("expected prompt after showAddPrompt")
	}
	if result.prompt.label != "Add source" {
		t.Fatalf("expected 'Add source' label, got %q", result.prompt.label)
	}
}

func TestCtrlNShowsInitPrompt(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.showInitPrompt()
	result := mustModel(t, updated)
	if result.prompt == nil {
		t.Fatal("expected prompt after showInitPrompt")
	}
	if result.prompt.label != "Init name" {
		t.Fatalf("expected 'Init name' label, got %q", result.prompt.label)
	}
}

func TestCtrlRTriggersReload(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.handleReload()
	result := mustModel(t, updated)
	if result.status != "loading" {
		t.Fatalf("expected status 'loading', got %q", result.status)
	}
}

func TestCtrlACyclesAgent(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.handleCycleAgent()
	result := mustModel(t, updated)
	if result.agentDisplay() == "all" {
		t.Fatal("expected agent to cycle away from 'all' on first ctrl+a")
	}
}

func TestEscReturnsToHome(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()
	m.state = stateListing
	m.errMsg = "some error"

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEsc})
	result := mustModel(t, updated)
	if result.state != stateHome {
		t.Fatalf("expected stateHome after esc, got %v", result.state)
	}
	if result.errMsg != "" {
		t.Fatal("expected errMsg to be cleared")
	}
}

func TestF1ShowsHelp(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()

	updated, _ := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyF1})
	result := mustModel(t, updated)
	if result.state != stateViewingHelp {
		t.Fatalf("expected stateViewingHelp after F1, got %v", result.state)
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command on ctrl+c")
	}
}

func TestDownKeyNavigatesList(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()
	m.setSkillItems([]skill.Skill{{Name: "a"}, {Name: "b"}})

	_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyDown})
	_ = cmd // may be nil for single-item or bottom, just verify no panic
}

func TestUpKeyNavigatesList(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.registry = commands.NewRegistry()
	m.setSkillItems([]skill.Skill{{Name: "a"}, {Name: "b"}})

	_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyUp})
	_ = cmd
}

func TestPromptEscCancels(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.showFindPrompt()
	m2 := mustModel(t, updated)

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m2.Update(msg)
	result := mustModel(t, updated)
	if result.prompt != nil {
		t.Fatal("expected prompt to be dismissed after esc")
	}
}

func TestPromptEnterExecutesAction(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.showInitPrompt()
	m2 := mustModel(t, updated)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m2.Update(msg)
	result := mustModel(t, updated)
	if result.prompt != nil {
		t.Fatal("expected prompt to be dismissed after enter")
	}
	if result.status != "loading" {
		t.Fatalf("expected status 'loading' after init, got %q", result.status)
	}
}

func TestPromptBlocksUnrelatedKeys(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.showFindPrompt()
	m2 := mustModel(t, updated)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, _ = m2.Update(msg)
	result := mustModel(t, updated)
	if result.state != stateHome {
		t.Fatalf("expected stateHome (key consumed by prompt), got %v", result.state)
	}
	if result.prompt == nil {
		t.Fatal("prompt should still be active")
	}
}
