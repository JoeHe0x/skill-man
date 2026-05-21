package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
)

func TestOpenCommandPalette(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, cmd := m.openCommandPalette()
	m2 := mustModel(t, updated)
	if m2.state != stateCommandPalette {
		t.Fatalf("expected stateCommandPalette, got %v", m2.state)
	}
	if m2.cmdPalette.ui == nil {
		t.Fatal("expected palette model")
	}
	if cmd == nil {
		t.Fatal("expected blink cmd")
	}
	if len(m2.cmdPalette.ui.filtered) == 0 {
		t.Fatal("expected palette matches")
	}
}

func TestPaletteFuzzyFilter(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	_, _ = m.openCommandPalette()
	for _, r := range "reload" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updated, _ := m.Update(msg)
		m = mustModel(t, updated)
	}
	if len(m.cmdPalette.ui.filtered) == 0 {
		t.Fatal("expected matches for reload")
	}
	top := m.cmdPalette.ui.all[m.cmdPalette.ui.filtered[0]]
	if !strings.Contains(strings.ToLower(top.title), "reload") {
		t.Fatalf("expected reload on top, got %q", top.title)
	}
}

func TestPaletteEscCloses(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.state = stateListing
	_, _ = m.openCommandPalette()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := mustModel(t, updated)
	if m2.cmdPalette.ui != nil {
		t.Fatal("expected palette closed")
	}
	if m2.state != stateListing {
		t.Fatalf("expected stateListing after esc, got %v", m2.state)
	}
}

func TestPaletteRunInspectWhenSkillSelected(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{BaseExtension: extension.BaseExtension{Name: "demo"}}}, nil))
	m.refreshActiveList()
	m.state = stateListing

	_, _ = m.openCommandPalette()
	// filter to inspect
	for _, r := range "inspect" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mustModel(t, updated)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := mustModel(t, updated)
	if m2.cmdPalette.ui != nil {
		t.Fatal("palette should close after run")
	}
}
