package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestBindChoicesToListItemsMapsOneToOne(t *testing.T) {
	t.Parallel()

	choices := []agentBindChoice{
		{agent: agent.Agent{ID: "cursor"}, scope: extension.ScopeProject, configPath: "/p/.cursor/mcp.json", desired: true},
		{agent: agent.Agent{ID: "codex"}, scope: extension.ScopeGlobal, configPath: "/h/.codex/config.toml", desired: false},
		{agent: agent.Agent{ID: "windsurf"}, scope: extension.ScopeGlobal, configPath: "/h/windsurf/mcp_config.json", desired: true},
	}
	items := bindChoicesToListItems(choices, "/proj", "/home/joe")
	if len(items) != len(choices) {
		t.Fatalf("got %d items, want %d (1:1 mapping)", len(items), len(choices))
	}
}

func TestBindChoicesToListItemsUsesSkillDirMeta(t *testing.T) {
	t.Parallel()

	group := agent.AgentBySkillsDir(".agents/skills")
	if len(group) == 0 {
		t.Fatal("no agents for .agents/skills")
	}
	choices := []agentBindChoice{{
		agents:   group,
		skillDir: ".agents/skills",
		agent:    skillBindDisplayAgent(group),
		desired:  true,
	}}
	items := bindChoicesToListItems(choices, t.TempDir(), "/home/joe")
	li, ok := items[0].(listItem)
	if !ok {
		t.Fatal("expected listItem")
	}
	if li.meta != ".agents/skills" {
		t.Fatalf("meta=%q want .agents/skills", li.meta)
	}
}

func TestBindChoicesToListItemsPreservesScope(t *testing.T) {
	t.Parallel()

	choices := []agentBindChoice{
		{
			agent:      agent.Agent{Name: "Codex", ID: "codex"},
			scope:      extension.ScopeGlobal,
			configPath: "/home/joe/.codex/config.toml",
			desired:    true,
		},
	}
	items := bindChoicesToListItems(choices, t.TempDir(), "/home/joe")
	li, ok := items[0].(listItem)
	if !ok {
		t.Fatal("expected listItem")
	}
	if li.meta != "codex" {
		t.Fatalf("meta=%q, want codex", li.meta)
	}
}
