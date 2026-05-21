package app

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestEmptyMCPList_statusBarHidden_footerZero(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.activeTab = panel.TabMCP
	m.agentIDs = []string{"all"}
	m.status = "ready"
	m.panels.Get(panel.TabMCP).ApplyScan(panel.MCPScan(nil, nil))

	m.refreshActiveList()

	if m.Main.ShowStatusBar() {
		t.Fatal("expected status bar hidden for empty MCP placeholder")
	}
	if len(m.Main.Items()) != 1 {
		t.Fatalf("expected 1 placeholder row, got %d", len(m.Main.Items()))
	}
	footer := m.footerStatsLine()
	if !strings.Contains(footer, "0 mcp") {
		t.Fatalf("footer should show 0 mcp, got %q", footer)
	}
	if !strings.Contains(footer, "agents: all") {
		t.Fatalf("footer should show agents: all, got %q", footer)
	}
}
