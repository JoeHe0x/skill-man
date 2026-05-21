package bind

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

func TestChoicesToListItemsMapsOneToOne(t *testing.T) {
	t.Parallel()

	choices := []usecasebind.Choice{
		{Agent: agent.Agent{ID: "cursor"}, Scope: extension.ScopeProject, ConfigPath: "/p/.cursor/mcp.json", Desired: true},
		{Agent: agent.Agent{ID: "codex"}, Scope: extension.ScopeGlobal, ConfigPath: "/h/.codex/config.toml", Desired: false},
		{Agent: agent.Agent{ID: "windsurf"}, Scope: extension.ScopeGlobal, ConfigPath: "/h/windsurf/mcp_config.json", Desired: true},
	}
	items := ChoicesToListItems(choices, "/proj", "/home/joe")
	if len(items) != len(choices) {
		t.Fatalf("got %d items, want %d (1:1 mapping)", len(items), len(choices))
	}
}

func TestChoicesToListItemsUsesSkillDirMeta(t *testing.T) {
	t.Parallel()

	group := agent.AgentBySkillsDir(".agents/skills")
	if len(group) == 0 {
		t.Fatal("no agents for .agents/skills")
	}
	choices := []usecasebind.Choice{{
		Agents:   group,
		SkillDir: ".agents/skills",
		Agent:    usecasebind.DisplayAgent(group),
		Desired:  true,
	}}
	items := ChoicesToListItems(choices, t.TempDir(), "/home/joe")
	li, ok := items[0].(panel.Item)
	if !ok {
		t.Fatal("expected panel.Item")
	}
	if li.Meta != ".agents/skills" {
		t.Fatalf("meta=%q want .agents/skills", li.Meta)
	}
}

func TestChoicesToListItemsPreservesScope(t *testing.T) {
	t.Parallel()

	choices := []usecasebind.Choice{
		{
			Agent:      agent.Agent{Name: "Codex", ID: "codex"},
			Scope:      extension.ScopeGlobal,
			ConfigPath: "/home/joe/.codex/config.toml",
			Desired:    true,
		},
	}
	items := ChoicesToListItems(choices, t.TempDir(), "/home/joe")
	li, ok := items[0].(panel.Item)
	if !ok {
		t.Fatal("expected panel.Item")
	}
	if li.Meta != "codex" {
		t.Fatalf("meta=%q, want codex", li.Meta)
	}
}
