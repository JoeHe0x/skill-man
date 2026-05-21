package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
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

func TestPromptLifecycle(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	if m.prompt.Active() {
		t.Fatal("expected inactive prompt initially")
	}

	updated, cmd := m.showFindPrompt()
	m2 := mustModel(t, updated)
	if !m2.prompt.Active() {
		t.Fatal("expected prompt after showFindPrompt")
	}
	if m2.prompt.PromptLabel() != "Find" {
		t.Fatalf("expected prompt label 'Find', got %q", m2.prompt.PromptLabel())
	}
	if cmd == nil {
		t.Fatal("expected blink cmd from showPrompt")
	}

	// Esc cancels prompt
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m2.Update(msg)
	m3 := mustModel(t, updated)
	if m3.prompt.Active() {
		t.Fatal("expected prompt inactive after esc")
	}
}

func TestExtensionTabSwitch(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{BaseExtension: extension.BaseExtension{Name: "test-skill"}}}, nil))
	if m.activeTab != panel.TabSkills {
		t.Fatalf("expected initial tab Skills, got %v", m.activeTab)
	}

	cmd := m.setActiveTab(panel.TabMCP)
	if cmd != nil {
		t.Fatal("expected static preview cmd when switching to MCP with no servers")
	}
	if m.activeTab != panel.TabMCP {
		t.Fatalf("expected MCP tab, got %v", m.activeTab)
	}

	cmd = m.switchExtensionTab(false)
	if m.activeTab != panel.TabSkills {
		t.Fatalf("expected Skills tab after Tab, got %v", m.activeTab)
	}
	if cmd == nil {
		t.Fatal("expected preview sync cmd when returning to Skills with items")
	}
}

func TestPromptEnterExecutes(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{BaseExtension: extension.BaseExtension{Name: "my-query-skill"}}}, nil))
	m.refreshActiveList()

	updated, _ := m.showInitPrompt()
	m2 := mustModel(t, updated)

	for _, r := range "my-skill" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updated, _ := m2.Update(msg)
		m2 = mustModel(t, updated)
	}
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m2.Update(msg)
	m3 := mustModel(t, updated)

	if m3.prompt.Active() {
		t.Fatal("expected prompt inactive after enter")
	}
}
