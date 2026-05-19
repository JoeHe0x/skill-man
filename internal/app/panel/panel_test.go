package panel

import (
	"context"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestMCPPanelScanAndList(t *testing.T) {
	t.Parallel()

	panel := NewMCPPanel(MCPDeps{
		Scan: func(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]*mcpdomain.Server, error) {
			_ = ctx
			_ = projectRoot
			_ = home
			_ = agents
			return []*mcpdomain.Server{{
				BaseExtension: extension.BaseExtension{Name: "filesystem"},
				ConfigKey:     "filesystem",
			}}, nil
		},
	})

	if !panel.ApplyScan(MCPScannedMsg{Servers: []*mcpdomain.Server{
		{BaseExtension: extension.BaseExtension{Name: "filesystem"}, ConfigKey: "filesystem"},
	}}) {
		t.Fatal("expected ApplyScan to accept MCPScannedMsg")
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
