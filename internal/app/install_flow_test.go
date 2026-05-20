package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestStartInstallFlowSkillsTab(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, cmd := m.startInstallFlow()
	m2 := mustModel(t, updated)
	if m2.install.flow == nil {
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
	if m2.install.flow != nil {
		t.Fatal("expected no install flow on MCP tab")
	}
}
