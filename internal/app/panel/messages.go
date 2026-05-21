package panel

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// ScannedMsg delivers panel scan results to the TUI (one message type per batch).
type ScannedMsg struct {
	Gen     uint64
	Tab     Tab
	Skills  []*skilldomain.Skill
	Servers []*mcpdomain.Server
	Err     error
}

// SkillsScan builds a skills-tab scan result (Gen is set by the scan coordinator).
func SkillsScan(skills []*skilldomain.Skill, err error) ScannedMsg {
	return ScannedMsg{Tab: TabSkills, Skills: skills, Err: err}
}

// MCPScan builds an MCP-tab scan result (Gen is set by the scan coordinator).
func MCPScan(servers []*mcpdomain.Server, err error) ScannedMsg {
	return ScannedMsg{Tab: TabMCP, Servers: servers, Err: err}
}

// PreviewLoadedMsg delivers async preview content.
type PreviewLoadedMsg struct {
	Tab     Tab
	Content string
	Err     error
	Gen     int
}
