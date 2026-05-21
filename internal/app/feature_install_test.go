package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestStartInstallFlowSkillsTab(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, cmd := m.startInstallFlow()
	m2 := mustModel(t, updated)
	if !m2.install.WizardOpen() {
		t.Fatal("expected install flow on skills tab")
	}
	if m2.state != stateInstalling {
		t.Fatalf("expected stateInstalling, got %v", m2.state)
	}
	if cmd == nil {
		t.Fatal("expected blink cmd for search input")
	}
}

func TestStartInstallFlowMCPDisabled(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	_ = m.setActiveTab(panel.TabMCP)
	updated, cmd := m.startInstallFlow()
	m2 := mustModel(t, updated)
	if cmd == nil {
		t.Fatal("expected flash cmd when MCP install unavailable")
	}
	if m2.install.WizardOpen() {
		t.Fatal("expected no install flow on MCP tab")
	}
}

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
