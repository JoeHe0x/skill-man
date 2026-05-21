package bind

import (
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// Choice tracks desired vs initial bind state for one row in the bind UI.
// Skill rows group every agent that shares a skills directory (e.g. .agents/skills).
type Choice struct {
	Agent      agent.Agent // display label; MCP rows use this agent only
	Agents     []agent.Agent
	SkillDir   string // skill rows: EntityDirs[EntitySkill] for index lookup
	Scope      extension.Scope
	ConfigPath string // MCP only; destination config file for this row
	Initial    bool
	Desired    bool
}

func groupAgents(c Choice) []agent.Agent {
	if len(c.Agents) > 0 {
		return c.Agents
	}
	return []agent.Agent{c.Agent}
}
