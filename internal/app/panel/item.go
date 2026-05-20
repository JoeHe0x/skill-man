package panel

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/JoeHe0x/skill-man/internal/commands"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// ItemKind classifies list rows built by panels.
type ItemKind int

const (
	ItemCommand ItemKind = iota
	ItemSkill
	ItemMCP
	ItemMessage
)

// Item is the panel-neutral list row model used directly in the app layer.
type Item struct {
	Kind        ItemKind
	Title       string
	Desc        string
	Meta        string
	DetailLines []string
	Command     commands.Spec
	Skill       *skilldomain.Skill
	MCP         *mcpdomain.Server   // representative instance for mutations
	MCPKey      string              // config key (level-1 list selection)
	MCPMembers  []*mcpdomain.Server // all config files using this key
}

func (i Item) FilterValue() string {
	if i.Kind == ItemMCP && len(i.MCPMembers) > 0 {
		return mcpKeyFilterValue(i.MCPKey, i.MCPMembers)
	}
	parts := []string{i.Title, i.Desc, i.Meta}
	parts = append(parts, i.DetailLines...)
	return strings.ToLower(strings.Join(parts, " "))
}

// --- polymorphic action targets ---

// InspectTarget describes what to inspect for a selected item.
type InspectTarget struct {
	Kind       string // "skill", "mcp"
	SkillPath  string
	MCPKey     string
	MCPMembers []*mcpdomain.Server
}

// DisableTarget describes what to toggle disable for a selected item.
type DisableTarget struct {
	Kind       string
	Skill      *skilldomain.Skill
	MCPMembers []*mcpdomain.Server
}

// RemoveTarget describes what to remove for a selected item.
type RemoveTarget struct {
	Kind       string
	Skill      *skilldomain.Skill
	MCPName    string
	MCPMembers []*mcpdomain.Server
}

// BindTarget describes what to bind for a selected item.
type BindTarget struct {
	Kind       string
	Skill      *skilldomain.Skill
	MCPKey     string
	MCPMembers []*mcpdomain.Server
}

// UpdateTarget describes what to update for a selected item.
type UpdateTarget struct {
	Kind  string
	Skill *skilldomain.Skill
}

func (i Item) CanInspect() bool {
	return i.Kind == ItemSkill || (i.Kind == ItemMCP && len(i.MCPMembers) > 0)
}
func (i Item) CanDisable() bool {
	return i.Kind == ItemSkill || (i.Kind == ItemMCP && len(i.MCPMembers) > 0)
}
func (i Item) CanRemove() bool {
	return i.Kind == ItemSkill || (i.Kind == ItemMCP && len(i.MCPMembers) > 0)
}
func (i Item) CanBind() bool {
	return i.Kind == ItemSkill || (i.Kind == ItemMCP && (len(i.MCPMembers) > 0 || i.MCPKey != ""))
}
func (i Item) CanUpdate() bool { return i.Kind == ItemSkill }

func (i Item) InspectTarget() InspectTarget {
	switch i.Kind {
	case ItemSkill:
		return InspectTarget{Kind: "skill", SkillPath: i.Skill.Path}
	case ItemMCP:
		return InspectTarget{Kind: "mcp", MCPKey: i.MCPKey, MCPMembers: i.MCPMembers}
	}
	return InspectTarget{}
}

func (i Item) DisableTarget() DisableTarget {
	switch i.Kind {
	case ItemSkill:
		return DisableTarget{Kind: "skill", Skill: i.Skill}
	case ItemMCP:
		return DisableTarget{Kind: "mcp", MCPMembers: i.MCPMembers}
	}
	return DisableTarget{}
}

func (i Item) RemoveTarget() RemoveTarget {
	switch i.Kind {
	case ItemSkill:
		return RemoveTarget{Kind: "skill", Skill: i.Skill}
	case ItemMCP:
		name := i.MCPKey
		if name == "" && i.MCP != nil {
			name = i.MCP.ConfigKey
		}
		return RemoveTarget{Kind: "mcp", MCPName: name, MCPMembers: i.MCPMembers}
	}
	return RemoveTarget{}
}

func (i Item) BindTarget() BindTarget {
	switch i.Kind {
	case ItemSkill:
		return BindTarget{Kind: "skill", Skill: i.Skill}
	case ItemMCP:
		key := i.MCPKey
		if key == "" && i.MCP != nil {
			key = i.MCP.ConfigKey
		}
		return BindTarget{Kind: "mcp", MCPKey: key, MCPMembers: i.MCPMembers}
	}
	return BindTarget{}
}

func (i Item) UpdateTarget() UpdateTarget {
	if i.Kind == ItemSkill {
		return UpdateTarget{Kind: "skill", Skill: i.Skill}
	}
	return UpdateTarget{}
}

// CommandItems builds list rows for the help command catalog.
func CommandItems(specs []commands.Spec) []Item {
	items := make([]Item, 0, len(specs))
	for _, spec := range specs {
		meta := spec.Usage
		if spec.Dangerous {
			meta += " | dangerous"
		}
		items = append(items, Item{
			Kind:    ItemCommand,
			Title:   "/" + spec.Name,
			Desc:    spec.Summary,
			Meta:    meta,
			Command: spec,
		})
	}
	return items
}

func skillListItems(skills []*skilldomain.Skill, agentFilter []string) []Item {
	if len(skills) == 0 {
		return []Item{{
			Kind:  ItemMessage,
			Title: "No skills found",
			Desc:  "Press Ctrl+R to rescan after adding local skills.",
			Meta:  "empty",
		}}
	}

	items := make([]Item, 0, len(skills))
	for _, sk := range skills {
		if !matchesAgentFilter(sk.GetAgents(), agentFilter) {
			continue
		}

		tools := "no tools"
		if len(sk.Tools) > 0 {
			tools = strings.Join(sk.Tools, ", ")
		}

		agents := "no agents"
		if len(sk.GetAgents()) > 0 {
			agents = strings.Join(sk.GetAgents(), ", ")
		}

		management := "unmanaged"
		if sk.IsManaged() {
			management = sk.SourceKind
		}

		title := sk.GetName()
		if sk.GetScope() == skilldomain.ScopeGlobal {
			title = sk.GetName() + " [global]"
		}
		if sk.IsDisabled() {
			title = "[x] " + title
		}

		items = append(items, Item{
			Kind:  ItemSkill,
			Title: title,
			Desc:  sk.GetDescription(),
			Meta:  fmt.Sprintf("%s | agents: %s | %s | %s | %s", sk.GetScope(), agents, tools, management, sk.GetUpdatedAt().Format(time.DateOnly)),
			Skill: sk,
		})
	}
	return items
}

func mcpListItems(servers []*mcpdomain.Server, agentFilter []string, home string) []Item {
	filtered := make([]*mcpdomain.Server, 0, len(servers))
	for _, srv := range servers {
		if matchesAgentFilter(srv.GetAgents(), agentFilter) {
			filtered = append(filtered, srv)
		}
	}
	if len(filtered) == 0 {
		return []Item{{
			Kind:  ItemMessage,
			Title: "No MCP servers found",
			Desc:  "Add MCP config for Cursor, Claude Code, or Windsurf, then press Ctrl+R.",
			Meta:  "empty",
		}}
	}

	// Level 1 only in the list; agent · scope · path appear in the preview pane.
	return mcpKeyListItems(filtered, home)
}

func matchesAgentFilter(agents, filter []string) bool {
	if len(filter) == 0 || slices.Contains(filter, "all") {
		return true
	}
	for _, id := range filter {
		if slices.Contains(agents, id) {
			return true
		}
	}
	return false
}
