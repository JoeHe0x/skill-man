package panel

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// ViewState mirrors app session states relevant to panel titles.
type ViewState int

const (
	ViewHome ViewState = iota
	ViewListing
	ViewSearching
	ViewHelp
	ViewBinding
	ViewInspecting
)

// Capabilities describes which actions the active panel supports.
type Capabilities struct {
	Inspect bool
	Disable bool
	Bind    bool
	Remove  bool
	Update  bool
	Find    bool
	Add     bool
	Init    bool
}

// Panel drives list content, scanning, and preview for one extension tab.
type Panel interface {
	Tab() Tab
	Count() int
	CountLabel() string
	Capabilities() Capabilities

	ScanCmd(cwd, home string, agents []agent.Agent) tea.Cmd
	ApplyScan(msg tea.Msg) (refresh bool)

	ListItems(agentFilter []string) []Item
	SearchItems(query string, agentFilter []string) []Item

	PanelTitle(state ViewState) string
	ReloadHint() string
	StaticPreview() string
	SyncPreview(selected Item, width int, previewGen *int) tea.Cmd

	SelectedSkill(item Item) bool
	SelectedMCP(item Item) bool
}

// Registry holds all extension panels keyed by tab.
type Registry struct {
	panels map[Tab]Panel
	order  []Tab
}

// NewRegistry builds the default skill and MCP panels.
func NewRegistry(skill SkillDeps, mcp MCPDeps) *Registry {
	return &Registry{
		panels: map[Tab]Panel{
			TabSkills: NewSkillPanel(skill),
			TabMCP:    NewMCPPanel(mcp),
		},
		order: []Tab{TabSkills, TabMCP},
	}
}

// Get returns the panel for a tab.
func (r *Registry) Get(tab Tab) Panel {
	return r.panels[tab]
}

// Tabs returns tab order for rendering.
func (r *Registry) Tabs() []Tab {
	return r.order
}

// Skills returns scanned skills from the skills panel.
func (r *Registry) Skills() []*skilldomain.Skill {
	if p, ok := r.panels[TabSkills].(*skillPanel); ok {
		return p.Skills()
	}
	return nil
}

// MCPServers returns scanned MCP servers from the MCP panel.
func (r *Registry) MCPServers() []*mcpdomain.Server {
	if p, ok := r.panels[TabMCP].(*mcpPanel); ok {
		return p.Servers()
	}
	return nil
}

// ScanAllCmd triggers scan for every panel.
func (r *Registry) ScanAllCmd(cwd, home string, agents []agent.Agent) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(r.order))
	for _, tab := range r.order {
		cmds = append(cmds, r.panels[tab].ScanCmd(cwd, home, agents))
	}
	return tea.Batch(cmds...)
}
