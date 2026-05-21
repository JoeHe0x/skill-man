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
	ViewInstalling
	ViewHelp
	ViewBinding
	ViewInspecting
)

// Capabilities describes which actions the active panel supports.
type Capabilities struct {
	Inspect       bool
	Disable       bool
	Bind          bool
	Remove        bool
	Update        bool
	Find          bool
	Add           bool
	Init          bool
	SearchInstall bool
}

// Panel drives list content, scanning, and preview for one extension tab.
type Panel interface {
	Tab() Tab
	Count() int
	CountLabel() string
	Capabilities() Capabilities

	ScanCmd(cwd, home string, agents []agent.Agent) tea.Cmd
	ApplyScan(msg ScannedMsg) (refresh bool)

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
// Panels are registered by position; Tab is auto-assigned from registration order.
type Registry struct {
	panels map[Tab]Panel
	order  []Tab
}

// NewRegistry builds a registry from the given panels. Tabs are assigned
// in registration order: the first panel gets TabSkills, second gets TabMCP, etc.
func NewRegistry(panels ...Panel) *Registry {
	r := &Registry{
		panels: make(map[Tab]Panel, len(panels)),
		order:  make([]Tab, len(panels)),
	}
	for i, p := range panels {
		tab := Tab(i)
		r.panels[tab] = p
		r.order[i] = tab
	}
	return r
}

// Get returns the panel for a tab.
func (r *Registry) Get(tab Tab) Panel {
	return r.panels[tab]
}

// Tabs returns tab order for rendering.
func (r *Registry) Tabs() []Tab {
	return r.order
}

// Skills returns scanned skills from the first panel that implements SkillProvider.
func (r *Registry) Skills() []*skilldomain.Skill {
	for _, tab := range r.order {
		if sp, ok := r.panels[tab].(SkillProvider); ok {
			return sp.Skills()
		}
	}
	return nil
}

// MCPServers returns scanned MCP servers from the first panel that implements MCPProvider.
func (r *Registry) MCPServers() []*mcpdomain.Server {
	for _, tab := range r.order {
		if mp, ok := r.panels[tab].(MCPProvider); ok {
			return mp.Servers()
		}
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
