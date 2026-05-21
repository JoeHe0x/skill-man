package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestNoteScanCompleted_waitsForAllPanels(t *testing.T) {
	m := New("/tmp", "/home/test")
	m.scan.Gen = 1
	m.scan.Pending = 2
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
	m.scan.Gen = 2
	m.scan.Pending = 2
	m.status = "loading"

	_ = m.noteScanCompleted(1)
	if m.scan.Pending != 2 {
		t.Fatalf("scan.Pending = %d, want 2 (stale gen ignored)", m.scan.Pending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}

func TestHandleScanned_doesNotEndLoadingBeforeSkillsScan(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.scan.Gen = 1
	m.scan.Pending = 2
	m.status = "loading"
	m.activeTab = panel.TabSkills

	msg := panel.MCPScan(nil, nil)
	msg.Gen = 1
	updated, _ := m.handleScanned(msg)
	m = mustModel(t, updated)

	if m.status != "loading" {
		t.Fatalf("status = %q, want loading until skills scan completes", m.status)
	}
	if m.scan.Pending != 1 {
		t.Fatalf("scan.Pending = %d, want 1", m.scan.Pending)
	}
}

func TestHandleScanned_ignoresStaleGeneration(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.scan.Gen = 2
	m.scan.Pending = 2
	m.status = "loading"

	msg := panel.SkillsScan(nil, nil)
	msg.Gen = 1
	updated, _ := m.handleScanned(msg)
	m = mustModel(t, updated)

	if m.scan.Pending != 2 {
		t.Fatalf("scan.Pending = %d, want 2", m.scan.Pending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}

func TestBeginScanAllCmd_resetsPendingOnRapidReload(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	_ = m.beginScanAllCmd() // gen 1
	m.scan.Gen = 1
	m.scan.Pending = 1 // one panel still in flight from first batch

	_ = m.beginScanAllCmd() // gen 2, full batch
	if m.scan.Gen != 2 {
		t.Fatalf("scan.Gen = %d, want 2", m.scan.Gen)
	}
	if m.scan.Pending != 2 {
		t.Fatalf("scan.Pending = %d, want 2", m.scan.Pending)
	}

	staleMCP := panel.MCPScan(nil, nil)
	staleMCP.Gen = 1
	updated, _ := m.handleScanned(staleMCP)
	m = mustModel(t, updated)
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading after stale MCP scan", m.status)
	}

	skills := panel.SkillsScan(nil, nil)
	skills.Gen = 2
	updated, _ = m.handleScanned(skills)
	m = mustModel(t, updated)
	if m.scan.Pending != 1 {
		t.Fatalf("scan.Pending = %d, want 1", m.scan.Pending)
	}

	mcp := panel.MCPScan(nil, nil)
	mcp.Gen = 2
	updated, _ = m.handleScanned(mcp)
	m = mustModel(t, updated)
	if m.status != "ready" {
		t.Fatalf("status = %q, want ready", m.status)
	}
}

func TestNew_startsInLoadingState(t *testing.T) {
	m := New("/tmp", "/home/test")
	if m.scan.Pending != 0 {
		t.Fatalf("scan.Pending = %d, want 0 before Init", m.scan.Pending)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading", m.status)
	}
}
