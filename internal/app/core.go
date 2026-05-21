package app

import (
	"slices"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

// Core holds session shell state: terminal size, paths, status/footer, agent filter, and scan batch tracking.
type Core struct {
	width  int
	height int

	cwd  string
	home string

	status string
	scan   scanCoordinator
	errMsg string

	footerFlash    string
	footerFlashTag int
	footerContext  string

	focusedPane focusPane
	agentIDs    []string
	allAgents   []agent.Agent
}

func newCore(cwd, home string) Core {
	return Core{
		cwd:         cwd,
		home:        home,
		status:      "loading",
		focusedPane: focusPaneList,
		agentIDs:    []string{"all"},
		allAgents:   agent.DefaultAgents(),
	}
}

func (c *Core) contentWidth() int {
	w := c.width - 2 // doc horizontal padding
	if w < 20 {
		return 20
	}
	return w
}

func (c *Core) shouldStack() bool {
	return c.width < 80
}

func (c *Core) reportError(err error) {
	if err == nil {
		return
	}
	c.status = "error"
	c.errMsg = err.Error()
}

func (c *Core) clearError() {
	c.errMsg = ""
}

func (c *Core) activeAgents() []agent.Agent {
	if len(c.agentIDs) == 0 || slices.Contains(c.agentIDs, "all") {
		return c.allAgents
	}
	var out []agent.Agent
	for _, id := range c.agentIDs {
		if a, ok := agent.AgentByID(id); ok {
			out = append(out, a)
		}
	}
	return out
}

func (c *Core) agentDisplay() string {
	if slices.Contains(c.agentIDs, "all") || len(c.agentIDs) == 0 {
		return "all"
	}
	return strings.Join(c.agentIDs, ",")
}
