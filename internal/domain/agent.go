package domain

import (
	"path/filepath"
	"slices"
	"strings"
)

type Agent struct {
	Name      string
	ID        string
	SkillsDir string // relative path from project/home root to skills directory, e.g. ".claude/skills"
}

func DefaultAgents() []Agent {
	return []Agent{
		{Name: "AiderDesk", ID: "aider-desk", SkillsDir: ".aider-desk/skills"},
		{Name: "Amp", ID: "amp", SkillsDir: ".agents/skills"},
		{Name: "Kimi Code CLI", ID: "kimi-cli", SkillsDir: ".agents/skills"},
		{Name: "Replit", ID: "replit", SkillsDir: ".agents/skills"},
		{Name: "Universal", ID: "universal", SkillsDir: ".agents/skills"},
		{Name: "Antigravity", ID: "antigravity", SkillsDir: ".agents/skills"},
		{Name: "Augment", ID: "augment", SkillsDir: ".augment/skills"},
		{Name: "IBM Bob", ID: "bob", SkillsDir: ".bob/skills"},
		{Name: "Claude Code", ID: "claude-code", SkillsDir: ".claude/skills"},
		{Name: "OpenClaw", ID: "openclaw", SkillsDir: "skills"},
		{Name: "Cline", ID: "cline", SkillsDir: ".agents/skills"},
		{Name: "Dexto", ID: "dexto", SkillsDir: ".agents/skills"},
		{Name: "Warp", ID: "warp", SkillsDir: ".agents/skills"},
		{Name: "CodeArts Agent", ID: "codearts-agent", SkillsDir: ".codeartsdoer/skills"},
		{Name: "CodeBuddy", ID: "codebuddy", SkillsDir: ".codebuddy/skills"},
		{Name: "Codemaker", ID: "codemaker", SkillsDir: ".codemaker/skills"},
		{Name: "Code Studio", ID: "codestudio", SkillsDir: ".codestudio/skills"},
		{Name: "Codex", ID: "codex", SkillsDir: ".agents/skills"},
		{Name: "Command Code", ID: "command-code", SkillsDir: ".commandcode/skills"},
		{Name: "Continue", ID: "continue", SkillsDir: ".continue/skills"},
		{Name: "Cortex Code", ID: "cortex", SkillsDir: ".cortex/skills"},
		{Name: "Crush", ID: "crush", SkillsDir: ".crush/skills"},
		{Name: "Cursor", ID: "cursor", SkillsDir: ".agents/skills"},
		{Name: "Deep Agents", ID: "deepagents", SkillsDir: ".agents/skills"},
		{Name: "Devin for Terminal", ID: "devin", SkillsDir: ".devin/skills"},
		{Name: "Droid", ID: "droid", SkillsDir: ".factory/skills"},
		{Name: "Firebender", ID: "firebender", SkillsDir: ".agents/skills"},
		{Name: "ForgeCode", ID: "forgecode", SkillsDir: ".forge/skills"},
		{Name: "Gemini CLI", ID: "gemini-cli", SkillsDir: ".agents/skills"},
		{Name: "GitHub Copilot", ID: "github-copilot", SkillsDir: ".agents/skills"},
		{Name: "Goose", ID: "goose", SkillsDir: ".goose/skills"},
		{Name: "Hermes Agent", ID: "hermes-agent", SkillsDir: ".hermes/skills"},
		{Name: "Junie", ID: "junie", SkillsDir: ".junie/skills"},
		{Name: "iFlow CLI", ID: "iflow-cli", SkillsDir: ".iflow/skills"},
		{Name: "Kilo Code", ID: "kilo", SkillsDir: ".kilocode/skills"},
		{Name: "Kiro CLI", ID: "kiro-cli", SkillsDir: ".kiro/skills"},
		{Name: "Kode", ID: "kode", SkillsDir: ".kode/skills"},
		{Name: "MCPJam", ID: "mcpjam", SkillsDir: ".mcpjam/skills"},
		{Name: "Mistral Vibe", ID: "mistral-vibe", SkillsDir: ".vibe/skills"},
		{Name: "Mux", ID: "mux", SkillsDir: ".mux/skills"},
		{Name: "OpenCode", ID: "opencode", SkillsDir: ".agents/skills"},
		{Name: "OpenHands", ID: "openhands", SkillsDir: ".openhands/skills"},
		{Name: "Pi", ID: "pi", SkillsDir: ".pi/skills"},
		{Name: "Qoder", ID: "qoder", SkillsDir: ".qoder/skills"},
		{Name: "Qwen Code", ID: "qwen-code", SkillsDir: ".qwen/skills"},
		{Name: "Rovo Dev", ID: "rovodev", SkillsDir: ".rovodev/skills"},
		{Name: "Roo Code", ID: "roo", SkillsDir: ".roo/skills"},
		{Name: "Tabnine CLI", ID: "tabnine-cli", SkillsDir: ".tabnine/agent/skills"},
		{Name: "Trae", ID: "trae", SkillsDir: ".trae/skills"},
		{Name: "Trae CN", ID: "trae-cn", SkillsDir: ".trae/skills"},
		{Name: "Windsurf", ID: "windsurf", SkillsDir: ".windsurf/skills"},
		{Name: "Zencoder", ID: "zencoder", SkillsDir: ".zencoder/skills"},
		{Name: "Neovate", ID: "neovate", SkillsDir: ".neovate/skills"},
		{Name: "Pochi", ID: "pochi", SkillsDir: ".pochi/skills"},
		{Name: "AdaL", ID: "adal", SkillsDir: ".adal/skills"},
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
		if a.SkillsDir == dir {
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
		if !seen[a.SkillsDir] {
			seen[a.SkillsDir] = true
			dirs = append(dirs, a.SkillsDir)
		}
	}
	slices.Sort(dirs)
	return dirs
}

// MatchesAgent checks if this skill path belongs to the given agent.
func MatchesAgent(skillPath, root string, agent Agent) bool {
	skillDir := filepath.Join(root, agent.SkillsDir)
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
			projectDir := filepath.Join(projectRoot, a.SkillsDir)
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
			globalDir := filepath.Join(home, a.SkillsDir)
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
