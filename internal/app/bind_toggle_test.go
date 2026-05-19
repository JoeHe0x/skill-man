package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestBindChoiceIndexMatchesAgentAndScope(t *testing.T) {
	t.Parallel()

	choices := []agentBindChoice{
		{agent: agent.Agent{ID: "cursor"}, scope: extension.ScopeProject, configPath: "/p/.cursor/mcp.json", desired: true},
		{agent: agent.Agent{ID: "codex"}, scope: extension.ScopeGlobal, configPath: "/h/.codex/config.toml", desired: false},
		{agent: agent.Agent{ID: "windsurf"}, scope: extension.ScopeGlobal, configPath: "/h/windsurf/mcp_config.json", desired: true},
	}
	if got := bindChoiceIndex(choices, "codex", extension.ScopeGlobal, "/h/.codex/config.toml"); got != 1 {
		t.Fatalf("bindChoiceIndex = %d, want 1", got)
	}
	if got := bindChoiceIndex(choices, "windsurf", extension.ScopeProject, "/h/windsurf/mcp_config.json"); got != -1 {
		t.Fatalf("bindChoiceIndex = %d, want -1 for missing scope", got)
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
	if li.meta != "codex" || li.bindScope != extension.ScopeGlobal || li.configPath != "/home/joe/.codex/config.toml" {
		t.Fatalf("meta=%q bindScope=%q configPath=%q", li.meta, li.bindScope, li.configPath)
	}
}
