package bind

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	skillservice "github.com/JoeHe0x/skill-man/internal/service/skill"
)

func TestMcpTargetBound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	codexPath := filepath.Join(home, ".codex", "config.toml")

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "filesystem",
			ConfigPath: cursorPath,
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
	}

	target := servicemcp.BindTarget{
		Agent:      agent.Agent{ID: "cursor"},
		Scope:      extension.ScopeGlobal,
		ConfigPath: cursorPath,
	}

	t.Run("empty bindings falls back to server agents", func(t *testing.T) {
		if !MCPTargetBound(srv, target) {
			t.Fatal("expected bound via server-level agents")
		}
	})

	t.Run("agent id in binding", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: codexPath,
				Scope:      extension.ScopeGlobal,
				ConfigKey:  "filesystem",
				Agents:     []string{"codex"},
			}},
		}
		codexTarget := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "codex"},
			Scope:      extension.ScopeGlobal,
			ConfigPath: codexPath,
		}
		if !MCPTargetBound(s, codexTarget) {
			t.Fatal("expected bound via binding agents list")
		}
	})

	t.Run("path clean normalizes slashes", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: cursorPath + "/",
				Scope:      extension.ScopeGlobal,
				ConfigKey:  "filesystem",
			}},
		}
		if !MCPTargetBound(s, target) {
			t.Fatal("expected bound via cleaned config path")
		}
	})

	t.Run("empty bindings slice must not use stale aggregated agents", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			BaseExtension: extension.BaseExtension{
				ConfigPath: cursorPath,
				Scope:      extension.ScopeProject,
				Agents:     []string{"cursor", "codex", "windsurf"},
			},
			Bindings: []mcpdomain.Binding{},
		}
		windsurfTarget := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "windsurf"},
			Scope:      extension.ScopeProject,
			ConfigPath: filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"),
		}
		if MCPTargetBound(s, windsurfTarget) {
			t.Fatal("empty Bindings with stale srv.Agents must not mark windsurf bound")
		}
	})

	t.Run("path match requires agent when binding lists agents", func(t *testing.T) {
		windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: windsurfPath,
				ConfigKey:  "filesystem",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"codex"},
			}},
		}
		target := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "windsurf"},
			Scope:      extension.ScopeProject,
			ConfigPath: windsurfPath,
		}
		if MCPTargetBound(s, target) {
			t.Fatal("must not mark windsurf bound when only codex is listed for that config file")
		}
	})

	t.Run("bound via config path regardless of scope label", func(t *testing.T) {
		windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: windsurfPath,
				ConfigKey:  "filesystem",
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"windsurf"},
			}},
		}
		target := servicemcp.BindTarget{
			Agent:      agent.Agent{ID: "windsurf"},
			Scope:      extension.ScopeProject,
			ConfigPath: windsurfPath,
		}
		if !MCPTargetBound(s, target) {
			t.Fatal("expected bound when server exists in windsurf config for windsurf agent")
		}
	})

	t.Run("different path not bound", func(t *testing.T) {
		s := &mcpdomain.Server{
			ConfigKey: "filesystem",
			Bindings: []mcpdomain.Binding{{
				ConfigPath: codexPath,
				Scope:      extension.ScopeGlobal,
				ConfigKey:  "filesystem",
			}},
		}
		if MCPTargetBound(s, target) {
			t.Fatal("expected not bound for unrelated config path")
		}
	})
}

func TestMCPBindChoicesReflectAllMembers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	codexPath := filepath.Join(home, ".codex", "config.toml")

	members := []*mcpdomain.Server{
		{
			BaseExtension: extension.BaseExtension{
				Name:       "server-filesystem",
				ConfigPath: cursorPath,
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"cursor"},
			},
			ConfigKey: "filesystem",
			Command:   "npx",
		},
		{
			BaseExtension: extension.BaseExtension{
				Name:       "server-filesystem",
				ConfigPath: codexPath,
				Scope:      extension.ScopeGlobal,
				Agents:     []string{"codex"},
			},
			ConfigKey: "filesystem",
			Command:   "npx",
		},
	}

	b := NewBinder(nil, servicemcp.NewManager(), root, home)
	choices := b.NewMCPChoices(members)
	byID := map[string]bool{}
	for _, c := range choices {
		if c.Scope != extension.ScopeGlobal {
			continue
		}
		switch c.Agent.ID {
		case "cursor":
			byID["cursor"] = c.Initial
		case "codex":
			byID["codex"] = c.Initial
		case "windsurf":
			byID["windsurf"] = c.Initial
		}
	}
	if !byID["cursor"] {
		t.Fatal("cursor global should be bound (member at cursor path)")
	}
	if !byID["codex"] {
		t.Fatal("codex global should be bound (member at codex path)")
	}
	if byID["windsurf"] {
		t.Fatal("windsurf should not be bound when no member at windsurf path")
	}
}

func TestMCPBindChoicesOneRowPerTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:   "server-filesystem",
			Agents: []string{"cursor"},
		},
		ConfigKey: "filesystem",
	}
	b := NewBinder(nil, servicemcp.NewManager(), root, home)
	choices := b.NewMCPChoices([]*mcpdomain.Server{srv})
	targets := servicemcp.ListBindTargets(root, home)
	if len(choices) != len(targets) {
		t.Fatalf("expected %d MCP bind targets, got %d choices", len(targets), len(choices))
	}
	if len(choices) < 4 {
		t.Fatalf("expected at least 4 MCP bind targets (multi-scope), got %d", len(choices))
	}
	seen := map[string]bool{}
	for _, c := range choices {
		key := c.Agent.ID + "|" + string(c.Scope) + "|" + c.ConfigPath
		if seen[key] {
			t.Fatalf("duplicate target: %s", key)
		}
		seen[key] = true
	}
}

func TestApplyMCPBindChoicesMultipleAgents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(cursorPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeMCPJSON(t, cursorPath, "filesystem", "npx", []string{"-y", "pkg"})

	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{
			Name:       "server-filesystem",
			ConfigPath: cursorPath,
			Scope:      extension.ScopeGlobal,
			Agents:     []string{"cursor"},
		},
		ConfigKey: "filesystem",
		Command:   "npx",
		Args:      []string{"-y", "pkg"},
	}

	mgr := servicemcp.NewManager()
	b := NewBinder(nil, mgr, root, home)
	choices := b.NewMCPChoices([]*mcpdomain.Server{srv})
	for i := range choices {
		switch {
		case choices[i].Agent.ID == "codex" && choices[i].Scope == extension.ScopeGlobal:
			choices[i].Desired = true
		case choices[i].Agent.ID == "windsurf":
			choices[i].Desired = true
		}
	}

	if err := b.ApplyMCP(srv, choices); err != nil {
		t.Fatalf("apply: %v", err)
	}

	codexPath := filepath.Join(home, ".codex", "config.toml")
	if _, err := os.Stat(codexPath); err != nil {
		t.Fatalf("codex config: %v", err)
	}
	windsurfPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	if _, err := os.Stat(windsurfPath); err != nil {
		t.Fatalf("windsurf config: %v", err)
	}
}

func TestSkillBindChoicesGroupSharedDir(t *testing.T) {
	t.Parallel()

	skill := &skilldomain.Skill{
		BaseExtension: extension.BaseExtension{
			Name:   "demo",
			Agents: []string{"cursor"},
		},
	}
	b := NewBinder(nil, servicemcp.NewManager(), t.TempDir(), "/home/joe")
	choices := b.NewSkillChoices(skill)

	var shared *Choice
	for i := range choices {
		if choices[i].SkillDir == ".agents/skills" {
			shared = &choices[i]
			break
		}
	}
	if shared == nil {
		t.Fatal("missing .agents/skills bind row")
	}
	if len(shared.Agents) < 10 {
		t.Fatalf(".agents/skills group has %d agents, want many", len(shared.Agents))
	}
	if shared.Initial {
		t.Fatal("row should be unchecked when only one agent in the group is bound")
	}
	if shared.Desired != shared.Initial {
		t.Fatal("desired should match initial")
	}
	idx := -1
	for i, c := range choices {
		if c.SkillDir == ".agents/skills" {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("should find row by skillDir meta")
	}
}

func TestApplySkillBindChoicesAllAgentsInSharedDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := t.TempDir()
	skillDir := filepath.Join(root, ".skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := &skilldomain.Skill{
		BaseExtension: extension.BaseExtension{
			Name:       "demo",
			Path:       skillDir,
			ConfigPath: filepath.Join(skillDir, "SKILL.md"),
			Scope:      extension.ScopeProject,
		},
	}

	group := agent.AgentBySkillsDir(".agents/skills")
	if len(group) < 2 {
		t.Fatal("expected multiple agents sharing .agents/skills")
	}

	choices := []Choice{{
		Agents:   group,
		SkillDir: ".agents/skills",
		Agent:    DisplayAgent(group),
		Desired:  true,
	}}

	mgr := manager.NewManager[*skilldomain.Skill](skillservice.SkillScanStrategy{})
	b := NewBinder(mgr, servicemcp.NewManager(), root, home)
	ctx := context.Background()
	if err := b.ApplySkill(ctx, skill, choices); err != nil {
		t.Fatalf("bind group: %v", err)
	}

	link := filepath.Join(root, ".agents", "skills", "demo")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("expected symlink at shared dir: %v", err)
	}

	choices[0].Desired = false
	if err := b.ApplySkill(ctx, skill, choices); err != nil {
		t.Fatalf("unbind group: %v", err)
	}
	if _, err := os.Lstat(link); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("symlink should be removed after unbinding the whole group")
	}
}

func TestApplySkillBindUnbindRelocatesPrimaryOutOfSharedDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := t.TempDir()
	skillDir := filepath.Join(root, ".agents", "skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	skill := &skilldomain.Skill{
		BaseExtension: extension.BaseExtension{
			Name:       "demo",
			Path:       skillDir,
			ConfigPath: filepath.Join(skillDir, "SKILL.md"),
			Scope:      extension.ScopeProject,
		},
	}

	group := agent.AgentBySkillsDir(".agents/skills")
	choices := []Choice{{
		Agents:   group,
		SkillDir: ".agents/skills",
		Agent:    DisplayAgent(group),
		Desired:  false,
	}}

	mgr := manager.NewManager[*skilldomain.Skill](skillservice.SkillScanStrategy{})
	b := NewBinder(mgr, servicemcp.NewManager(), root, home)
	if err := b.ApplySkill(context.Background(), skill, choices); err != nil {
		t.Fatalf("unbind group: %v", err)
	}
	moved := filepath.Join(root, ".skills", "demo")
	if _, err := os.Stat(moved); err != nil {
		t.Fatalf("skill should move to .skills: %v", err)
	}
	if _, err := os.Stat(skillDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("skill should no longer live under .agents/skills")
	}
}

func TestSkillDirGroupBoundOnDiskSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := t.TempDir()
	skillDir := filepath.Join(root, ".skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(linkDir, "demo")
	rel, err := filepath.Rel(linkDir, skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rel, link); err != nil {
		t.Fatal(err)
	}

	skill := &skilldomain.Skill{
		BaseExtension: extension.BaseExtension{
			Path:  skillDir,
			Scope: extension.ScopeProject,
		},
	}
	rep, ok := agent.AgentByID("cursor")
	if !ok {
		t.Fatal("cursor agent missing")
	}
	if !SkillDirGroupBoundOnDisk(skill, rep, root, home) {
		t.Fatal("expected bound via shared-dir symlink")
	}
}

func TestBindChoicesTogglePreservesOthers(t *testing.T) {
	t.Parallel()

	choices := []Choice{
		{Agent: agent.Agent{Name: "Cursor", ID: "cursor"}, Scope: extension.ScopeGlobal, Initial: true, Desired: true},
		{Agent: agent.Agent{Name: "Codex", ID: "codex"}, Scope: extension.ScopeGlobal, Initial: false, Desired: false},
	}
	choices[1].Desired = true

	if !choices[0].Desired || !choices[1].Desired {
		t.Fatal("expected both desired after toggling codex")
	}
}

func writeMCPJSON(t *testing.T, path, key, command string, args []string) {
	t.Helper()
	content := `{"mcpServers":{"` + key + `":{"command":"` + command + `","args":[`
	for i, a := range args {
		if i > 0 {
			content += ","
		}
		content += `"` + a + `"`
	}
	content += `]}}}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
