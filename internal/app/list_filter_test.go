package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
)

func TestStartListFilter(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{BaseExtension: extension.BaseExtension{Name: "demo-skill"}}}, nil))
	m.refreshActiveList()

	updated, cmd := m.startListFilter()
	m2 := mustModel(t, updated)
	if m2.Main.FilterState() != list.Filtering {
		t.Fatalf("expected list.Filtering, got %v", m2.Main.FilterState())
	}
	if cmd == nil {
		t.Fatal("expected blink cmd for filter input")
	}
}

func TestListFilterTyping(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{
			{BaseExtension: extension.BaseExtension{Name: "alpha-only"}},
			{BaseExtension: extension.BaseExtension{Name: "zzz-other"}},
		}, nil))
	m.refreshActiveList()

	_, _ = m.startListFilter()
	for _, r := range "alpha" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updated, _ := m.Update(msg)
		m = mustModel(t, updated)
	}
	if m.Main.FilterValue() != "alpha" {
		t.Fatalf("expected filter value alpha, got %q", m.Main.FilterValue())
	}
	if m.Main.FilterState() != list.Filtering && m.Main.FilterState() != list.FilterApplied {
		t.Fatalf("unexpected filter state %v", m.Main.FilterState())
	}
}
