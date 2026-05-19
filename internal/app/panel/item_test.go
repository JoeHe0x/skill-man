package panel

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

func TestMCPListKeyLevelOnly(t *testing.T) {
	t.Parallel()

	servers := []*mcpdomain.Server{
		{
			BaseExtension: extension.BaseExtension{
				Name:       "server-filesystem",
				ConfigPath: "/home/joe/.cursor/mcp.json",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"cursor"},
			},
			ConfigKey: "filesystem",
		},
		{
			BaseExtension: extension.BaseExtension{
				Name:       "server-filesystem",
				ConfigPath: "/home/joe/.codex/config.toml",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"codex"},
			},
			ConfigKey: "filesystem",
		},
		{
			BaseExtension: extension.BaseExtension{
				Name:       "server-everything",
				ConfigPath: "/proj/.cursor/mcp.json",
				Scope:      extension.ScopeProject,
				Agents:     []string{"cursor"},
			},
			ConfigKey: "my-test-server",
		},
	}

	items := mcpListItems(servers, []string{"all"}, "/home/joe")
	if len(items) != 2 {
		t.Fatalf("expected 2 key rows, got %d", len(items))
	}
	for _, it := range items {
		if it.Kind != ItemMCP {
			t.Fatalf("expected only MCP key rows, got kind %v title %q", it.Kind, it.Title)
		}
		if len(it.MCPMembers) == 0 {
			t.Fatal("key row must carry members for preview")
		}
	}
	if items[0].MCPKey != "filesystem" || len(items[0].MCPMembers) != 2 {
		t.Fatalf("filesystem group: key=%q members=%d", items[0].MCPKey, len(items[0].MCPMembers))
	}
}

func TestRenderKeyPreviewShowsPlacements(t *testing.T) {
	t.Parallel()

	members := []*mcpdomain.Server{
		{
			BaseExtension: extension.BaseExtension{
				ConfigPath: "/home/joe/.cursor/mcp.json",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"cursor"},
			},
			ConfigKey: "filesystem",
			Command:   "npx",
		},
	}
	out, err := servicemcp.RenderKeyPreview("filesystem", members, "/home/joe", 80)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Locations") || !strings.Contains(strings.ToLower(out), "cursor") {
		t.Fatalf("preview should list agent placements: %q", out)
	}
}
