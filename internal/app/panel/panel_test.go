package panel

import (
	"context"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
)

func TestMCPPanelScanAndList(t *testing.T) {
	t.Parallel()

	panel := NewMCPPanel()

	if !panel.ApplyScan(MCPScan([]*mcpdomain.Server{
		{BaseExtension: extension.BaseExtension{Name: "filesystem"}, ConfigKey: "filesystem"},
	}, nil)) {
		t.Fatal("expected ApplyScan to accept MCP scan result")
	}

	items := panel.ListItems([]string{"all"})
	if len(items) != 1 {
		t.Fatalf("expected 1 MCP key row, got %d", len(items))
	}
	if items[0].Kind != ItemMCP || items[0].MCPKey != "filesystem" {
		t.Fatalf("expected MCP key row, got %+v", items[0])
	}
}

func TestTabCycle(t *testing.T) {
	t.Parallel()

	if TabSkills.Next() != TabMCP || TabMCP.Next() != TabSkills {
		t.Fatal("unexpected tab cycle forward")
	}
	if TabMCP.Prev() != TabSkills || TabSkills.Prev() != TabMCP {
		t.Fatal("unexpected tab cycle backward")
	}
}

func TestRegistryProviderAccess(t *testing.T) {
	t.Parallel()

	mgr := manager.NewManager[*skilldomain.Skill](nil)
	reg := NewRegistry(
		NewSkillPanel(mgr),
		NewMCPPanel(),
	)

	// Provider accessors work even before scan (return nil slice).
	_ = reg.Skills()
	_ = reg.MCPServers()
	if reg.Get(TabSkills) == nil {
		t.Fatal("expected TabSkills panel")
	}
	if reg.Get(TabMCP) == nil {
		t.Fatal("expected TabMCP panel")
	}
}

func TestScanAllCmd(t *testing.T) {
	t.Parallel()

	// Use a test-only scan function.
	panel := NewMCPPanel()
	reg := NewRegistry(panel)
	cmd := reg.ScanAllCmd("/tmp", "/home", agent.DefaultAgents())
	if cmd == nil {
		t.Fatal("expected non-nil ScanAllCmd")
	}
	_ = context.Background()
}
