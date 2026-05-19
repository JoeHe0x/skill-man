package app

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func panelToListItems(items []panel.Item) []list.Item {
	out := make([]list.Item, 0, len(items))
	for _, item := range items {
		out = append(out, panelToListItem(item))
	}
	return out
}

func panelToListItem(item panel.Item) listItem {
	li := listItem{
		title:       item.Title,
		desc:        item.Desc,
		meta:        item.Meta,
		detailLines: item.DetailLines,
		mcpKey:      item.MCPKey,
	}
	switch item.Kind {
	case panel.ItemCommand:
		li.kind = itemKindCommand
		li.command = item.Command
	case panel.ItemSkill:
		li.kind = itemKindSkill
		li.skill = item.Skill
	case panel.ItemMCP:
		li.kind = itemKindMCP
		li.mcp = item.MCP
		if len(item.MCPMembers) > 0 {
			li.mcpMembers = append([]*mcpdomain.Server(nil), item.MCPMembers...)
		}
	default:
		li.kind = itemKindMessage
	}
	return li
}

func listItemToPanel(item listItem) panel.Item {
	pi := panel.Item{
		Title:       item.title,
		Desc:        item.desc,
		Meta:        item.meta,
		DetailLines: item.detailLines,
	}
	switch item.kind {
	case itemKindCommand:
		pi.Kind = panel.ItemCommand
		pi.Command = item.command
	case itemKindSkill:
		pi.Kind = panel.ItemSkill
		pi.Skill = item.skill
	case itemKindMCP:
		pi.Kind = panel.ItemMCP
		pi.MCP = item.mcp
		pi.MCPKey = item.mcpKey
		if len(item.mcpMembers) > 0 {
			pi.MCPMembers = append([]*mcpdomain.Server(nil), item.mcpMembers...)
		}
	default:
		pi.Kind = panel.ItemMessage
	}
	return pi
}

func appViewState(state SessionState) panel.ViewState {
	switch state {
	case stateSearching:
		return panel.ViewSearching
	case stateInstalling:
		return panel.ViewInstalling
	case stateHelpOverlay:
		return panel.ViewListing
	case stateBindingAgent:
		return panel.ViewBinding
	case stateInspecting:
		return panel.ViewInspecting
	default:
		return panel.ViewListing
	}
}
