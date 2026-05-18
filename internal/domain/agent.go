package domain

import (
	"path/filepath"
	"slices"
	"strings"
)

type EntityType string

const (
	EntitySkill EntityType = "skill"
)

type Agent struct {
	Name       string
	ID         string
	SkillsDir  string // deprecated: use EntityDirs[EntitySkill]
	EntityDirs map[EntityType]string
}

func DefaultAgents() []Agent {
	return []Agent{
		{Name: "AiderDesk", ID: "aider-desk", SkillsDir: ".aider-desk/skills", EntityDirs: map[EntityType]string{EntitySkill: ".aider-desk/skills"}},
		{Name: "Amp", ID: "amp", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Kimi Code CLI", ID: "kimi-cli", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Replit", ID: "replit", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Universal", ID: "universal", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Antigravity", ID: "antigravity", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Augment", ID: "augment", SkillsDir: ".augment/skills", EntityDirs: map[EntityType]string{EntitySkill: ".augment/skills"}},
		{Name: "IBM Bob", ID: "bob", SkillsDir: ".bob/skills", EntityDirs: map[EntityType]string{EntitySkill: ".bob/skills"}},
		{Name: "Claude Code", ID: "claude-code", SkillsDir: ".claude/skills", EntityDirs: map[EntityType]string{EntitySkill: ".claude/skills"}},
		{Name: "OpenClaw", ID: "openclaw", SkillsDir: "skills", EntityDirs: map[EntityType]string{EntitySkill: "skills"}},
		{Name: "Cline", ID: "cline", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Dexto", ID: "dexto", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Warp", ID: "warp", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "CodeArts Agent", ID: "codearts-agent", SkillsDir: ".codeartsdoer/skills", EntityDirs: map[EntityType]string{EntitySkill: ".codeartsdoer/skills"}},
		{Name: "CodeBuddy", ID: "codebuddy", SkillsDir: ".codebuddy/skills", EntityDirs: map[EntityType]string{EntitySkill: ".codebuddy/skills"}},
		{Name: "Codemaker", ID: "codemaker", SkillsDir: ".codemaker/skills", EntityDirs: map[EntityType]string{EntitySkill: ".codemaker/skills"}},
		{Name: "Code Studio", ID: "codestudio", SkillsDir: ".codestudio/skills", EntityDirs: map[EntityType]string{EntitySkill: ".codestudio/skills"}},
		{Name: "Codex", ID: "codex", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Command Code", ID: "command-code", SkillsDir: ".commandcode/skills", EntityDirs: map[EntityType]string{EntitySkill: ".commandcode/skills"}},
		{Name: "Continue", ID: "continue", SkillsDir: ".continue/skills", EntityDirs: map[EntityType]string{EntitySkill: ".continue/skills"}},
		{Name: "Cortex Code", ID: "cortex", SkillsDir: ".cortex/skills", EntityDirs: map[EntityType]string{EntitySkill: ".cortex/skills"}},
		{Name: "Crush", ID: "crush", SkillsDir: ".crush/skills", EntityDirs: map[EntityType]string{EntitySkill: ".crush/skills"}},
		{Name: "Cursor", ID: "cursor", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Deep Agents", ID: "deepagents", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Devin for Terminal", ID: "devin", SkillsDir: ".devin/skills", EntityDirs: map[EntityType]string{EntitySkill: ".devin/skills"}},
		{Name: "Droid", ID: "droid", SkillsDir: ".factory/skills", EntityDirs: map[EntityType]string{EntitySkill: ".factory/skills"}},
		{Name: "Firebender", ID: "firebender", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "ForgeCode", ID: "forgecode", SkillsDir: ".forge/skills", EntityDirs: map[EntityType]string{EntitySkill: ".forge/skills"}},
		{Name: "Gemini CLI", ID: "gemini-cli", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "GitHub Copilot", ID: "github-copilot", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "Goose", ID: "goose", SkillsDir: ".goose/skills", EntityDirs: map[EntityType]string{EntitySkill: ".goose/skills"}},
		{Name: "Hermes Agent", ID: "hermes-agent", SkillsDir: ".hermes/skills", EntityDirs: map[EntityType]string{EntitySkill: ".hermes/skills"}},
		{Name: "Junie", ID: "junie", SkillsDir: ".junie/skills", EntityDirs: map[EntityType]string{EntitySkill: ".junie/skills"}},
		{Name: "iFlow CLI", ID: "iflow-cli", SkillsDir: ".iflow/skills", EntityDirs: map[EntityType]string{EntitySkill: ".iflow/skills"}},
		{Name: "Kilo Code", ID: "kilo", SkillsDir: ".kilocode/skills", EntityDirs: map[EntityType]string{EntitySkill: ".kilocode/skills"}},
		{Name: "Kiro CLI", ID: "kiro-cli", SkillsDir: ".kiro/skills", EntityDirs: map[EntityType]string{EntitySkill: ".kiro/skills"}},
		{Name: "Kode", ID: "kode", SkillsDir: ".kode/skills", EntityDirs: map[EntityType]string{EntitySkill: ".kode/skills"}},
		{Name: "MCPJam", ID: "mcpjam", SkillsDir: ".mcpjam/skills", EntityDirs: map[EntityType]string{EntitySkill: ".mcpjam/skills"}},
		{Name: "Mistral Vibe", ID: "mistral-vibe", SkillsDir: ".vibe/skills", EntityDirs: map[EntityType]string{EntitySkill: ".vibe/skills"}},
		{Name: "Mux", ID: "mux", SkillsDir: ".mux/skills", EntityDirs: map[EntityType]string{EntitySkill: ".mux/skills"}},
		{Name: "OpenCode", ID: "opencode", SkillsDir: ".agents/skills", EntityDirs: map[EntityType]string{EntitySkill: ".agents/skills"}},
		{Name: "OpenHands", ID: "openhands", SkillsDir: ".openhands/skills", EntityDirs: map[EntityType]string{EntitySkill: ".openhands/skills"}},
		{Name: "Pi", ID: "pi", SkillsDir: ".pi/skills", EntityDirs: map[EntityType]string{EntitySkill: ".pi/skills"}},
		{Name: "Qoder", ID: "qoder", SkillsDir: ".qoder/skills", EntityDirs: map[EntityType]string{EntitySkill: ".qoder/skills"}},
		{Name: "Qwen Code", ID: "qwen-code", SkillsDir: ".qwen/skills", EntityDirs: map[EntityType]string{EntitySkill: ".qwen/skills"}},
		{Name: "Rovo Dev", ID: "rovodev", SkillsDir: ".rovodev/skills", EntityDirs: map[EntityType]string{EntitySkill: ".rovodev/skills"}},
		{Name: "Roo Code", ID: "roo", SkillsDir: ".roo/skills", EntityDirs: map[EntityType]string{EntitySkill: ".roo/skills"}},
		{Name: "Tabnine CLI", ID: "tabnine-cli", SkillsDir: ".tabnine/agent/skills", EntityDirs: map[EntityType]string{EntitySkill: ".tabnine/agent/skills"}},
		{Name: "Trae", ID: "trae", SkillsDir: ".trae/skills", EntityDirs: map[EntityType]string{EntitySkill: ".trae/skills"}},
		{Name: "Trae CN", ID: "trae-cn", SkillsDir: ".trae/skills", EntityDirs: map[EntityType]string{EntitySkill: ".trae/skills"}},
		{Name: "Windsurf", ID: "windsurf", SkillsDir: ".windsurf/skills", EntityDirs: map[EntityType]string{EntitySkill: ".windsurf/skills"}},
		{Name: "Zencoder", ID: "zencoder", SkillsDir: ".zencoder/skills", EntityDirs: map[EntityType]string{EntitySkill: ".zencoder/skills"}},
		{Name: "Neovate", ID: "neovate", SkillsDir: ".neovate/skills", EntityDirs: map[EntityType]string{EntitySkill: ".neovate/skills"}},
		{Name: "Pochi", ID: "pochi", SkillsDir: ".pochi/skills", EntityDirs: map[EntityType]string{EntitySkill: ".pochi/skills"}},
		{Name: "AdaL", ID: "adal", SkillsDir: ".adal/skills", EntityDirs: map[EntityType]string{EntitySkill: ".adal/skills"}},
	}
}

func AgentByID(id string) (Agent, bool) {
	for _, a := range DefaultAgents() {
		if a.ID == id {
			return a, true
		}
	}
	return Agent{}, false
}

func AgentBySkillsDir(dir string) []Agent {
	var agents []Agent
	for _, a := range DefaultAgents() {
		if a.EntityDirs[EntitySkill] == dir {
			agents = append(agents, a)
		}
	}
	return agents
}

// UniqueSkillDirs returns deduplicated skill directory paths for the given agents.
func UniqueSkillDirs(agents []Agent) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, a := range agents {
		if !seen[a.EntityDirs[EntitySkill]] {
			seen[a.EntityDirs[EntitySkill]] = true
			dirs = append(dirs, a.EntityDirs[EntitySkill])
		}
	}
	slices.Sort(dirs)
	return dirs
}

// MatchesAgent checks if this skill path belongs to the given agent.
func MatchesAgent(skillPath, root string, agent Agent) bool {
	skillDir := filepath.Join(root, agent.EntityDirs[EntitySkill])
	rel, err := filepath.Rel(skillDir, skillPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// ResolveAgentIDs returns all agent IDs whose skill dir contains the given path.
func ResolveAgentIDs(skillDir, projectRoot, home string) []string {
	var ids []string
	skillName := filepath.Base(skillDir)

	resolvedSkill, err := filepath.EvalSymlinks(skillDir)
	if err != nil {
		resolvedSkill = skillDir
	}

	for _, a := range DefaultAgents() {
		if projectRoot != "" {
			projectDir := filepath.Join(projectRoot, a.EntityDirs[EntitySkill])
			if rel, err := filepath.Rel(projectDir, skillDir); err == nil && !strings.HasPrefix(rel, "..") {
				ids = append(ids, a.ID)
				continue
			}
			targetPath := filepath.Join(projectDir, skillName)
			if resolvedTarget, err := filepath.EvalSymlinks(targetPath); err == nil {
				if resolvedTarget == resolvedSkill {
					ids = append(ids, a.ID)
					continue
				}
			}
		}
		if home != "" {
			globalDir := filepath.Join(home, a.EntityDirs[EntitySkill])
			if rel, err := filepath.Rel(globalDir, skillDir); err == nil && !strings.HasPrefix(rel, "..") {
				ids = append(ids, a.ID)
				continue
			}
			targetPath := filepath.Join(globalDir, skillName)
			if resolvedTarget, err := filepath.EvalSymlinks(targetPath); err == nil {
				if resolvedTarget == resolvedSkill {
					ids = append(ids, a.ID)
				}
			}
		}
	}
	return ids
}
