package panel

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// SkillsScannedMsg delivers skill scan results to the TUI.
type SkillsScannedMsg struct {
	Skills []*skilldomain.Skill
	Err    error
}

// MCPScannedMsg delivers MCP scan results to the TUI.
type MCPScannedMsg struct {
	Servers []*mcpdomain.Server
	Err     error
}

// PreviewLoadedMsg delivers async preview content.
type PreviewLoadedMsg struct {
	Tab     Tab
	Content string
	Err     error
	Gen     int
}
