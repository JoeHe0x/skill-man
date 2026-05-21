package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
)

func TestHelpOverlay_opensWithoutChangingList(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{BaseExtension: extension.BaseExtension{Name: "keep-me"}}}, nil))
	m.refreshActiveList()
	before := len(m.list.Items())

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mustModel(t, updated)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyF1})
	m = mustModel(t, updated)
	if m.state != stateHelpOverlay {
		t.Fatalf("expected help overlay state, got %v", m.state)
	}
	if len(m.list.Items()) != before {
		t.Fatalf("help overlay should not replace list items: before=%d after=%d", before, len(m.list.Items()))
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mustModel(t, updated)
	if m.state == stateHelpOverlay {
		t.Fatal("expected help overlay closed after Esc")
	}
}
