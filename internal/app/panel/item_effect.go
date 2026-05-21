package panel

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// InspectEffect describes how to open inspection for a list item.
type InspectEffect struct {
	SkillPath  string
	MCPKey     string
	MCPMembers []*mcpdomain.Server
}

// DisableEffect describes a disable/enable toggle target.
type DisableEffect struct {
	Skill      *skilldomain.Skill
	MCPMembers []*mcpdomain.Server
}

// RemoveEffect describes a remove confirmation target.
type RemoveEffect struct {
	Skill      *skilldomain.Skill
	MCPName    string
	MCPMembers []*mcpdomain.Server
}

// BindEffect describes starting the agent bind flow.
type BindEffect struct {
	Skill      *skilldomain.Skill
	MCPKey     string
	MCPMembers []*mcpdomain.Server
}

// UpdateEffect describes a skill update target.
type UpdateEffect struct {
	Skill *skilldomain.Skill
}

// MCPConfigKey returns the MCP config key for this row, if any.
func (i Item) MCPConfigKey() string {
	if i.MCPKey != "" {
		return i.MCPKey
	}
	if i.MCP != nil && i.MCP.ConfigKey != "" {
		return i.MCP.ConfigKey
	}
	return ""
}

func (i Item) InspectEffect() (InspectEffect, bool) {
	switch i.Kind {
	case ItemSkill:
		if i.Skill == nil {
			return InspectEffect{}, false
		}
		return InspectEffect{SkillPath: i.Skill.Path}, true
	case ItemMCP:
		if len(i.MCPMembers) == 0 {
			return InspectEffect{}, false
		}
		return InspectEffect{MCPKey: i.MCPConfigKey(), MCPMembers: i.MCPMembers}, true
	default:
		return InspectEffect{}, false
	}
}

func (i Item) DisableEffect() (DisableEffect, bool) {
	switch i.Kind {
	case ItemSkill:
		if i.Skill == nil {
			return DisableEffect{}, false
		}
		return DisableEffect{Skill: i.Skill}, true
	case ItemMCP:
		if len(i.MCPMembers) == 0 {
			return DisableEffect{}, false
		}
		return DisableEffect{MCPMembers: i.MCPMembers}, true
	default:
		return DisableEffect{}, false
	}
}

func (i Item) RemoveEffect() (RemoveEffect, bool) {
	switch i.Kind {
	case ItemSkill:
		if i.Skill == nil {
			return RemoveEffect{}, false
		}
		return RemoveEffect{Skill: i.Skill}, true
	case ItemMCP:
		name := i.MCPConfigKey()
		if name == "" && i.MCP != nil {
			name = i.MCP.ConfigKey
		}
		return RemoveEffect{MCPName: name, MCPMembers: i.MCPMembers}, true
	default:
		return RemoveEffect{}, false
	}
}

func (i Item) BindEffect() (BindEffect, bool) {
	switch i.Kind {
	case ItemSkill:
		if i.Skill == nil {
			return BindEffect{}, false
		}
		return BindEffect{Skill: i.Skill}, true
	case ItemMCP:
		if len(i.MCPMembers) == 0 && i.MCPKey == "" {
			return BindEffect{}, false
		}
		return BindEffect{MCPKey: i.MCPConfigKey(), MCPMembers: i.MCPMembers}, true
	default:
		return BindEffect{}, false
	}
}

func (i Item) UpdateEffect() (UpdateEffect, bool) {
	if i.Kind == ItemSkill && i.Skill != nil {
		return UpdateEffect{Skill: i.Skill}, true
	}
	return UpdateEffect{}, false
}
