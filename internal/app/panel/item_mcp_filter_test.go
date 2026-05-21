package panel

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestMcpListItems_usesBindingAgentsForFilter(t *testing.T) {
	t.Parallel()

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{Name: "fs"},
		ConfigKey:     "filesystem",
		Bindings: []mcpdomain.Binding{
			{Agents: []string{"cursor"}, ConfigPath: "/home/joe/.cursor/mcp.json"},
		},
	}
	srv.SyncAggregatedFields()

	items := mcpListItems([]*mcpdomain.Server{srv}, []string{"cursor"}, "/home/joe")
	if len(items) != 1 || items[0].Kind != ItemMCP {
		t.Fatalf("expected 1 MCP key row, got %d (%v)", len(items), items)
	}

	filtered := mcpListItems([]*mcpdomain.Server{srv}, []string{"codex"}, "/home/joe")
	if len(filtered) != 1 || filtered[0].Kind != ItemMessage {
		t.Fatalf("expected filter placeholder, got %v", filtered)
	}
	if VisibleListCount(filtered) != 0 {
		t.Fatalf("VisibleListCount = %d, want 0", VisibleListCount(filtered))
	}
}
