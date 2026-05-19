package panel

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// MCPKeyGroup is one MCP config key and every on-disk instance (backend stays per file).
type MCPKeyGroup struct {
	ConfigKey string
	Members   []*mcpdomain.Server
}

func configKeyOf(srv *mcpdomain.Server) string {
	if srv.ConfigKey != "" {
		return srv.ConfigKey
	}
	return srv.GetName()
}

func groupServersByConfigKey(servers []*mcpdomain.Server) map[string][]*mcpdomain.Server {
	out := make(map[string][]*mcpdomain.Server)
	for _, srv := range servers {
		key := configKeyOf(srv)
		out[key] = append(out[key], srv)
	}
	for key := range out {
		sort.Slice(out[key], func(i, j int) bool {
			return out[key][i].ConfigPath < out[key][j].ConfigPath
		})
	}
	return out
}

func sortedConfigKeys(groups map[string][]*mcpdomain.Server) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func mcpKeyListItems(servers []*mcpdomain.Server, home string) []Item {
	groups := groupServersByConfigKey(servers)
	keys := sortedConfigKeys(groups)
	items := make([]Item, 0, len(keys))

	for _, key := range keys {
		members := groups[key]
		rep := members[0]
		title := key
		if mcpKeyDisabled(members) {
			title = "[x] " + key
		}
		items = append(items, Item{
			Kind:       ItemMCP,
			Title:      title,
			Desc:       mcpKeyListDesc(members, home),
			Meta:       mcpKeyListMeta(members),
			MCP:        rep,
			MCPKey:     key,
			MCPMembers: members,
		})
	}
	return items
}

func mcpKeyListDesc(members []*mcpdomain.Server, home string) string {
	agents := uniqueAgentIDs(members)
	n := len(members)
	switch {
	case len(agents) == 0:
		return fmt.Sprintf("%d config file(s)", n)
	case len(agents) == 1:
		return fmt.Sprintf("%s · %d location(s)", agentDisplayName(agents[0]), n)
	default:
		return fmt.Sprintf("%d agents · %d location(s)", len(agents), n)
	}
}

func mcpKeyListMeta(members []*mcpdomain.Server) string {
	rep := members[0]
	transport := "stdio"
	if rep.URL != "" {
		transport = "url"
	}
	status := "enabled"
	if mcpKeyDisabled(members) {
		status = "disabled"
	}
	return fmt.Sprintf("%s | %s", transport, status)
}

func mcpKeyDisabled(members []*mcpdomain.Server) bool {
	if len(members) == 0 {
		return false
	}
	for _, srv := range members {
		if !srv.AggregatedDisabled() {
			return false
		}
	}
	return true
}

func uniqueAgentIDs(members []*mcpdomain.Server) []string {
	seen := map[string]bool{}
	var ids []string
	for _, srv := range members {
		for _, id := range srv.GetAgents() {
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return agentDisplayName(ids[i]) < agentDisplayName(ids[j])
	})
	return ids
}

// Placement is one agent/scope/path row for preview (levels 2+3 merged).
type Placement struct {
	Agent      string
	AgentName  string
	Scope      string
	ConfigPath string
	Disabled   bool
}

func PlacementsForKey(members []*mcpdomain.Server, home string) []Placement {
	out := make([]Placement, 0, len(members))
	for _, srv := range members {
		agentID := primaryAgentID(srv)
		out = append(out, Placement{
			Agent:      agentID,
			AgentName:  agentDisplayName(agentID),
			Scope:      string(srv.Scope),
			ConfigPath: servicemcp.ShortPath(home, srv.ConfigPath),
			Disabled:   srv.AggregatedDisabled(),
		})
	}
	return out
}

func primaryAgentID(srv *mcpdomain.Server) string {
	if len(srv.Agents) > 0 {
		return srv.Agents[0]
	}
	return "unknown"
}

func agentDisplayName(id string) string {
	if a, ok := agent.AgentByID(id); ok {
		return a.Name
	}
	return id
}

func mcpKeyFilterValue(key string, members []*mcpdomain.Server) string {
	var parts []string
	parts = append(parts, key)
	for _, srv := range members {
		parts = append(parts, srv.GetName(), srv.ConfigPath, srv.Command, srv.URL)
		parts = append(parts, srv.GetAgents()...)
	}
	return strings.ToLower(strings.Join(parts, " "))
}
