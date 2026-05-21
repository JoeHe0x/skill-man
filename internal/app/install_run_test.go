package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestStartInstallSelected_returnsToListingWithBackgroundJob(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.startInstallFlow()
	m = mustModel(t, updated)
	if !m.install.WizardOpen() {
		t.Fatal("expected install flow")
	}

	m.install.PrepareWizardSelected(install.Candidate{
		Name:   "demo",
		Source: "owner/repo@demo",
	})

	updated, cmd := m.install.StartSelected([]string{"cursor"}, extension.ScopeProject)
	m = mustModel(t, updated)

	if m.state != stateListing {
		t.Fatalf("expected listing after starting install, got %v", m.state)
	}
	if m.install.WizardOpen() {
		t.Fatal("wizard should close when background install starts")
	}
	if !m.install.BackgroundActive() {
		t.Fatal("expected background progress job")
	}
	if cmd == nil {
		t.Fatal("expected progress + install commands")
	}
}
