package extension

import "time"

type Scope string

const (
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
)

// Extension defines the common interface for all manageable entities
// (skills, MCP servers, sub-agents, hooks)
type Extension interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetPath() string
	GetConfigPath() string
	GetScope() Scope
	IsDisabled() bool
	GetAgents() []string
	IsManaged() bool
	GetUpdatedAt() time.Time
}

// BaseExtension provides common fields for all extensions
type BaseExtension struct {
	ID           string
	Name         string
	Description  string
	Path         string
	ReadmePath   string
	ConfigPath   string // Path to the configuration file (e.g. SKILL.md, mcp.json)
	UpdatedAt    time.Time
	Managed      bool
	SourceKind   string
	SourcePath   string
	MetadataPath string
	Scope        Scope
	Agents       []string
	Disabled     bool
}

func (b BaseExtension) GetID() string           { return b.ID }
func (b BaseExtension) GetName() string         { return b.Name }
func (b BaseExtension) GetDescription() string  { return b.Description }
func (b BaseExtension) GetPath() string         { return b.Path }
func (b BaseExtension) GetConfigPath() string   { return b.ConfigPath }
func (b BaseExtension) GetScope() Scope         { return b.Scope }
func (b BaseExtension) IsDisabled() bool        { return b.Disabled }
func (b BaseExtension) GetAgents() []string     { return b.Agents }
func (b BaseExtension) IsManaged() bool         { return b.Managed }
func (b BaseExtension) GetUpdatedAt() time.Time { return b.UpdatedAt }
