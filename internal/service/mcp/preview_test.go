package mcp

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestRenderPreviewRendersMarkdown(t *testing.T) {
	t.Parallel()

	srv := mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: "/home/joe/.cursor/mcp.json",
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
	}

	out, err := RenderPreview(srv, 80)
	if err != nil {
		t.Fatalf("RenderPreview: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected non-empty preview")
	}
	if strings.Contains(out, "# MCP:") {
		t.Fatalf("expected glamour-rendered output, got raw markdown: %q", out[:min(80, len(out))])
	}
	if !strings.Contains(out, "server-filesystem") {
		t.Fatalf("expected server name in preview: %q", out[:min(120, len(out))])
	}
}
