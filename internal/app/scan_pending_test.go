package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestNoteScanCompleted_waitsForAllPanels(t *testing.T) {
	m := New("/tmp", "/home/test")
	m.scanGen = 1
	m.scansPending = 2
	m.status = "loading"

	_ = m.noteScanCompleted(1)
	if m.status != "loading" {
		t.Fatalf("after first scan: status = %q, want loading", m.status)
	}

	_ = m.noteScanCompleted(1)
	if m.status != "ready" {
		t.Fatalf("after second scan: status = %q, want ready", m.status)
	}
}

func TestNoteScanCompleted_ignoresStaleGeneration(t *testing.T) {
	m := New("/tmp", "/home/test")
	m.scanGen = 2
	m.scansPending = 2
	m.status = "loading"

	_ = m.noteScanCompleted(1)
	if m.scansPending != 2 {
		t.Fatalf("scansPending = %d, want 2 (stale gen ignored)", m.scansPending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}

func TestHandleMCPScanned_doesNotEndLoadingBeforeSkillsScan(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.scanGen = 1
	m.scansPending = 2
	m.status = "loading"
	m.activeTab = panel.TabSkills

	updated, _ := m.handleMCPScanned(panel.MCPScannedMsg{Gen: 1})
	m = mustModel(t, updated)

	if m.status != "loading" {
		t.Fatalf("status = %q, want loading until skills scan completes", m.status)
	}
	if m.scansPending != 1 {
		t.Fatalf("scansPending = %d, want 1", m.scansPending)
	}
}

func TestHandleSkillsScanned_ignoresStaleGeneration(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.scanGen = 2
	m.scansPending = 2
	m.status = "loading"

	updated, _ := m.handleSkillsScanned(panel.SkillsScannedMsg{Gen: 1})
	m = mustModel(t, updated)

	if m.scansPending != 2 {
		t.Fatalf("scansPending = %d, want 2", m.scansPending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}

func TestBeginScanAllCmd_resetsPendingOnRapidReload(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	_ = m.beginScanAllCmd() // gen 1
	m.scanGen = 1
	m.scansPending = 1 // one panel still in flight from first batch

	_ = m.beginScanAllCmd() // gen 2, full batch
	if m.scanGen != 2 {
		t.Fatalf("scanGen = %d, want 2", m.scanGen)
	}
	if m.scansPending != 2 {
		t.Fatalf("scansPending = %d, want 2", m.scansPending)
	}

	// Stale completion from gen 1 must not clear loading.
	updated, _ := m.handleMCPScanned(panel.MCPScannedMsg{Gen: 1})
	m = mustModel(t, updated)
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading after stale MCP scan", m.status)
	}

	updated, _ = m.handleSkillsScanned(panel.SkillsScannedMsg{Gen: 2})
	m = mustModel(t, updated)
	if m.scansPending != 1 {
		t.Fatalf("scansPending = %d, want 1", m.scansPending)
	}

	updated, _ = m.handleMCPScanned(panel.MCPScannedMsg{Gen: 2})
	m = mustModel(t, updated)
	if m.status != "ready" {
		t.Fatalf("status = %q, want ready", m.status)
	}
}

func TestNew_startsInLoadingState(t *testing.T) {
	m := New("/tmp", "/home/test")
	if m.scansPending != 0 {
		t.Fatalf("scansPending = %d, want 0 before Init", m.scansPending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}
