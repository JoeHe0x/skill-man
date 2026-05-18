package mcp

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func TestListTitleShowsMergeBadge(t *testing.T) {
	t.Parallel()

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{Name: "server-filesystem"},
		Bindings: []mcpdomain.Binding{
			{ConfigPath: "/home/joe/.cursor/mcp.json"},
			{ConfigPath: "/home/joe/.codex/config.toml"},
		},
	}
	title := ListTitle(srv)
	if !strings.Contains(title, "[×2 merged]") {
		t.Fatalf("unexpected title: %s", title)
	}
}

func TestListBindingDetailLines(t *testing.T) {
	t.Parallel()

	home := "/home/joe"
	srv := &mcpdomain.Server{
		Bindings: []mcpdomain.Binding{
			{
				ConfigPath: "/home/joe/.cursor/mcp.json",
				ConfigKey:  "filesystem",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"cursor"},
			},
		},
	}
	if lines := ListBindingDetailLines(srv, home); lines != nil {
		t.Fatalf("expected nil for single binding, got %v", lines)
	}

	srv.Bindings = append(srv.Bindings, mcpdomain.Binding{
		ConfigPath: "/mnt/c/Code/skill-man/.cursor/mcp.json",
		ConfigKey:  "filesystem",
		Scope:      extension.ScopeProject,
		Agents:     []string{"cursor"},
	})
	lines := ListBindingDetailLines(srv, home)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "cursor") || !strings.Contains(lines[0], "global") {
		t.Fatalf("unexpected line 0: %s", lines[0])
	}
}
