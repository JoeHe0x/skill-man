package installui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

func TestPaths_requiresSelection(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider()})
	m = m.WithSelected(domaininstall.Candidate{Name: "demo", Source: "owner/repo@demo"})
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

func TestPaths_enterStartsInstall(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider(), AgentIDs: []string{"all"}})
	m = m.WithSelected(domaininstall.Candidate{Name: "demo", Source: "owner/repo@demo"})
	if len(selectedAgentIDs(m.targets)) == 0 {
		m.targets[0].desired = true
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected install request cmd")
	}
	if _, ok := cmd().(RequestInstallMsg); !ok {
		t.Fatalf("expected RequestInstallMsg, got %T", cmd())
	}
}

func TestPaths_render(t *testing.T) {
	m := New(Config{Styles: theme.NewStyles(true), Provider: serviceinstall.NewSkillsCLIProvider()})
	m = m.WithSelected(domaininstall.Candidate{Name: "demo", Source: "owner/repo@demo"})
	out := m.renderPaths(64, 10)
	if !strings.Contains(out, "demo") {
		t.Fatalf("expected skill name in paths panel: %q", out)
	}
	if !strings.Contains(out, "Install paths") {
		t.Fatalf("expected paths title: %q", out)
	}
}
