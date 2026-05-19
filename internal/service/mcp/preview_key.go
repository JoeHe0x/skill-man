package mcp

import (
	"fmt"
	"strings"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	"github.com/JoeHe0x/skill-man/internal/render"
)

// RenderKeyPreview renders the right-pane detail for a selected MCP config key.
// Agent, scope, and path are shown here only (not in the left list).
func RenderKeyPreview(configKey string, members []*mcpdomain.Server, home string, width int) (string, error) {
	if len(members) == 0 {
		return "", fmt.Errorf("no MCP instances for key %q", configKey)
	}
	rep := members[0]

	var b strings.Builder
	fmt.Fprintf(&b, "# MCP key: `%s`\n\n", configKey)
	if rep.GetName() != "" && rep.GetName() != configKey {
		fmt.Fprintf(&b, "**Implementation:** %s\n\n", rep.GetName())
	}
	writeTransport(&b, rep.Command, rep.Args, rep.URL)

	fmt.Fprintf(&b, "## Locations\n\n")
	fmt.Fprintf(&b, "| Agent | Scope | Config |\n")
	fmt.Fprintf(&b, "|-------|-------|--------|\n")
	for _, srv := range members {
		agentID := "—"
		if len(srv.GetAgents()) > 0 {
			agentID = srv.GetAgents()[0]
		}
		status := ""
		if srv.AggregatedDisabled() {
			status = " [x]"
		}
		path := ShortPath(home, srv.ConfigPath)
		fmt.Fprintf(&b, "| %s%s | %s | `%s` |\n", agentID, status, srv.Scope, path)
	}

	fmt.Fprintf(&b, "\n_List shows MCP keys only · select a key to inspect placements · **b** bind · **x** toggle · **del** remove_\n")

	return render.Markdown(b.String(), width)
}
