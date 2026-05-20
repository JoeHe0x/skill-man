package installui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

func TestConfirm_requiresPaths(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider()})
	m.step = stepConfirm
	m.selected = domaininstall.Candidate{Name: "demo", Source: "owner/repo@demo"}
	m.targets = newDirChoices([]string{"all"})
	for i := range m.targets {
		m.targets[i].desired = false
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected hint cmd")
	}
	if _, ok := cmd().(HintMsg); !ok {
		t.Fatalf("expected HintMsg, got %T", cmd())
	}
	if next.step != stepPaths {
		t.Fatalf("expected paths step, got %v", next.step)
	}
}

func TestInstallQuitAttempt_requiresConfirmation(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider()})
	m.installing = true
	m.selected = domaininstall.Candidate{Name: "demo-skill"}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !next.QuitPending() {
		t.Fatal("first Esc should set quitPending")
	}
	next2, cmd := next.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next2.Installing() {
		t.Fatal("second Esc should clear installing")
	}
	if _, ok := cmd().(CancelInstallMsg); !ok {
		t.Fatalf("expected CancelInstallMsg, got %T", cmd())
	}
}

func TestConfirm_render(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider()})
	m.step = stepConfirm
	m.selected = domaininstall.Candidate{Name: "demo", Source: "owner/repo@demo"}
	m.targets = newDirChoices([]string{"adal"})
	for i := range m.targets {
		m.targets[i].desired = true
	}
	out := m.renderConfirm(64)
	if !strings.Contains(out, "demo") {
		t.Fatalf("expected skill name in confirm: %q", out)
	}
	if !strings.Contains(out, "Not installed yet") {
		t.Fatalf("expected pending hint in confirm: %q", out)
	}
	if strings.Contains(out, "agent(s): adal") {
		t.Fatalf("should show agent names not raw ids: %q", out)
	}
}
