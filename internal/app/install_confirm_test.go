package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestInstallConfirm_requiresPaths(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mustModel(t, updated)
	updated, _ = m.startInstallFlow()
	m = mustModel(t, updated)

	m.installFlow.step = installStepConfirm
	m.installFlow.selected.Name = "demo"
	m.installFlow.targets = newInstallDirChoices([]string{"all"})
	for i := range m.installFlow.targets {
		m.installFlow.targets[i].desired = false
	}

	updated, _ = m.handleInstallConfirmKeys(tea.KeyMsg{Type: tea.KeyEnter})
	m = mustModel(t, updated)
	if m.installFlow.step != installStepAgents {
		t.Fatalf("expected back to agents step when no path selected, got %v", m.installFlow.step)
	}
}

func TestInstallConfirm_summary(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mustModel(t, updated)
	m.installFlow = &installFlow{
		step: installStepConfirm,
		selected: domaininstall.Candidate{
			Name:   "demo",
			Source: "owner/pkg@demo",
		},
		targets: newInstallDirChoices([]string{"all"}),
	}
	out := m.renderInstallConfirm(60)
	if !strings.Contains(out, "demo") {
		t.Fatalf("summary should include skill name: %q", out)
	}
}
