package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestParseClaudeJSONProjectScoped(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	project := filepath.Join(home, "my-repo")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	config := `{
  "mcpServers": {
    "user-global": { "command": "echo", "args": ["global"] }
  },
  "projects": {
    "` + filepath.ToSlash(project) + `": {
      "mcpServers": {
        "project-local": { "command": "echo", "args": ["local"] }
      }
    },
    "` + filepath.ToSlash(filepath.Join(home, "other")) + `": {
      "mcpServers": {
        "other-project": { "command": "echo", "args": ["other"] }
      }
    }
  }
}`
	configPath := filepath.Join(home, ".claude.json")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	servers, err := ParseClaudeJSON(configPath, project, home)
	if err != nil {
		t.Fatalf("ParseClaudeJSON: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected user + matching project servers, got %d", len(servers))
	}

	byName := map[string]extension.Scope{}
	for _, srv := range servers {
		byName[srv.GetName()] = srv.GetScope()
	}
	if byName["user-global"] != extension.ScopeGlobal {
		t.Fatalf("expected user-global scope global, got %s", byName["user-global"])
	}
	if byName["project-local"] != extension.ScopeProject {
		t.Fatalf("expected project-local scope project, got %s", byName["project-local"])
	}
	if _, ok := byName["other-project"]; ok {
		t.Fatal("did not expect servers from other projects")
	}
}
