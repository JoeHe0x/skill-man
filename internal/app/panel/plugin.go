package panel

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// SkillProvider is optionally implemented by a Panel to expose scanned skills.
type SkillProvider interface {
	Skills() []*skilldomain.Skill
}

// MCPProvider is optionally implemented by a Panel to expose scanned MCP servers.
type MCPProvider interface {
	Servers() []*mcpdomain.Server
}
