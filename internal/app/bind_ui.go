package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

func bindChoicesToListItems(choices []usecasebind.Choice, projectRoot, home string) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		title := bindAgentTitle(c.Agent.Name, c.Desired)
		desc := bindAgentDesc(c.Agent)
		if c.Scope != "" {
			title = bindAgentTitle(fmt.Sprintf("%s (%s)", c.Agent.Name, c.Scope), c.Desired)
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

func bindAgentTitle(name string, checked bool) string {
	if checked {
		return "✓ " + name
	}
	return "  " + name
}

func bindAgentDesc(a agent.Agent) string {
	if dir := agent.MCPEntityDir(a); dir != "" {
		return dir
	}
	return a.EntityDirs[agent.EntitySkill]
}
