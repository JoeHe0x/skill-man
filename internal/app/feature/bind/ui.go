package bind

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

// ChoicesToListItems renders bind choices for the agent overlay list.
func ChoicesToListItems(choices []usecasebind.Choice, projectRoot, home string) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		title := agentTitle(c.Agent.Name, c.Desired)
		desc := agentDesc(c.Agent)
		if c.Scope != "" {
			title = agentTitle(fmt.Sprintf("%s (%s)", c.Agent.Name, c.Scope), c.Desired)
			desc = servicemcp.ShortPath(home, c.ConfigPath)
		}
		meta := c.Agent.ID
		if c.SkillDir != "" {
			meta = c.SkillDir
		}
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: title,
			Desc:  desc,
			Meta:  meta,
		})
	}
	return items
}

func agentTitle(name string, checked bool) string {
	if checked {
		return "✓ " + name
	}
	return "  " + name
}

// AgentDesc returns a one-line description for an agent in bind/filter lists.
func AgentDesc(a agent.Agent) string {
	return agentDesc(a)
}

func agentDesc(a agent.Agent) string {
	if dir := agent.MCPEntityDir(a); dir != "" {
		return dir
	}
	return a.EntityDirs[agent.EntitySkill]
}
