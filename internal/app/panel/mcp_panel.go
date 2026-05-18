package panel

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// MCPDeps configures the MCP panel.
type MCPDeps struct {
	Scan func(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]*mcpdomain.Server, error)
}

type mcpPanel struct {
	deps    MCPDeps
	servers []*mcpdomain.Server
	home    string
}

// NewMCPPanel creates the MCP extension panel.
func NewMCPPanel(deps MCPDeps) Panel {
	if deps.Scan == nil {
		deps.Scan = servicemcp.Scan
	}
	return &mcpPanel{deps: deps}
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

func (p *mcpPanel) ScanCmd(cwd, home string, agents []agent.Agent) tea.Cmd {
	scan := p.deps.Scan
	return func() tea.Msg {
		p.home = home
		servers, err := scan(context.Background(), cwd, home, agents)
		return MCPScannedMsg{Servers: servers, Err: err}
	}
}

func (p *mcpPanel) ApplyScan(msg tea.Msg) bool {
	m, ok := msg.(MCPScannedMsg)
	if !ok {
		return false
	}
	if m.Err != nil {
		return false
	}
	p.servers = m.Servers
	return true
}

// Servers returns the last scanned MCP server list.
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
		haystack := strings.ToLower(strings.Join([]string{
			srv.GetName(),
			srv.GetDescription(),
			srv.ConfigPath,
			srv.Command,
			srv.URL,
		}, " "))
		if strings.Contains(haystack, query) {
			results = append(results, srv)
		}
	}
	return mcpListItems(results, agentFilter, p.home)
}

func (p *mcpPanel) PanelTitle(state ViewState) string {
	if state == ViewSearching {
		return "MCP Search Results"
	}
	return "MCP Servers"
}

func (p *mcpPanel) ReloadHint() string { return "Rescanning MCP configs..." }

const mcpWelcomePreview = `# MCP Servers

Local MCP configs: Cursor (.cursor/mcp.json), Claude Code (.mcp.json, ~/.claude.json projects), Codex (.codex/config.toml), Windsurf (~/.codeium/windsurf/mcp_config.json).

- Tab: switch to Skills
- Ctrl+R: rescan configs
- Enter: preview selected server
- X: toggle disable | B: bind agents | Del: remove entry
- Ctrl+F: search servers
`

func (p *mcpPanel) StaticPreview() string { return mcpWelcomePreview }

func (p *mcpPanel) SyncPreview(selected Item, width int, previewGen *int) tea.Cmd {
	if selected.Kind != ItemMCP || selected.MCP == nil {
		return nil
	}
	if previewGen != nil {
		*previewGen++
		gen := *previewGen
		serverCopy := *selected.MCP
		return func() tea.Msg {
			content, err := servicemcp.RenderPreview(serverCopy, width)
			return PreviewLoadedMsg{Tab: TabMCP, Content: content, Err: err, Gen: gen}
		}
	}
	return nil
}

func (p *mcpPanel) SelectedSkill(item Item) bool { return false }

func (p *mcpPanel) SelectedMCP(item Item) bool {
	return item.Kind == ItemMCP && item.MCP != nil
}
