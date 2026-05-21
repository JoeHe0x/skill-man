package panel

import (
	"context"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

type mcpPanel struct {
	servers []*mcpdomain.Server
	home    string
}

// NewMCPPanel creates the MCP extension panel.
func NewMCPPanel() Panel {
	return &mcpPanel{}
}

func (p *mcpPanel) Tab() Tab { return TabMCP }

func (p *mcpPanel) Count() int { return len(p.servers) }

func (p *mcpPanel) CountLabel() string { return "mcp" }

func (p *mcpPanel) Capabilities() Capabilities {
	return Capabilities{
		Inspect: true,
		Disable: true,
		Bind:    true,
		Remove:  true,
		Find:    true,
	}
}

func (p *mcpPanel) Scan(ctx context.Context, cwd, home string, agents []agent.Agent) ScannedMsg {
	p.home = home
	servers, err := servicemcp.Scan(ctx, cwd, home, agents)
	return MCPScan(servers, err)
}

func (p *mcpPanel) ApplyScan(msg ScannedMsg) bool {
	if msg.Tab != TabMCP || msg.Err != nil {
		return false
	}
	p.servers = msg.Servers
	return true
}

// Servers returns the last scanned MCP server list (implements MCPProvider).
func (p *mcpPanel) Servers() []*mcpdomain.Server { return p.servers }

func (p *mcpPanel) ListItems(agentFilter []string) []Item {
	return mcpListItems(p.servers, agentFilter, p.home)
}

func (p *mcpPanel) SearchItems(query string, agentFilter []string) []Item {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return p.ListItems(agentFilter)
	}
	var results []*mcpdomain.Server
	for _, srv := range p.servers {
		if strings.Contains(mcpKeyFilterValue(configKeyOf(srv), []*mcpdomain.Server{srv}), query) {
			results = append(results, srv)
		}
	}
	return mcpListItems(results, agentFilter, p.home)
}

func (p *mcpPanel) PanelTitle(state ViewState) string {
	switch state {
	case ViewSearching:
		return "MCP Search Results"
	case ViewBinding:
		return "Bind MCP placements"
	default:
		return "MCP Servers"
	}
}

func (p *mcpPanel) ReloadHint() string { return "Rescanning MCP configs..." }

const mcpWelcomePreview = `# MCP Servers

Select an MCP **key** in the list. The preview pane shows every **agent · scope · path** for that key.

- Tab: switch to Skills
- Ctrl+R: rescan configs
- X: toggle disable (all locations) | B: bind to other agents | Del: remove key everywhere
- Ctrl+F: search keys
`

func (p *mcpPanel) StaticPreview() string { return mcpWelcomePreview }

func (p *mcpPanel) PreviewMarkdown(selected Item, width int) (string, error) {
	if selected.Kind != ItemMCP || len(selected.MCPMembers) == 0 {
		return "", nil
	}
	key := selected.MCPKey
	members := append([]*mcpdomain.Server(nil), selected.MCPMembers...)
	return renderMCPKeyPreview(key, members, p.home, width)
}

func (p *mcpPanel) SelectedSkill(item Item) bool { return false }

func (p *mcpPanel) SelectedMCP(item Item) bool {
	return item.Kind == ItemMCP && len(item.MCPMembers) > 0
}
