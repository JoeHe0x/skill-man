package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func buildPaletteCatalog(h paletteActionHost) []paletteItem {
	caps := h.ActivePanel().Capabilities()
	var items []paletteItem

	add := func(title, desc, search string, enabled bool, run func(paletteActionHost) (tea.Model, tea.Cmd)) {
		items = append(items, paletteItem{
			title:   title,
			desc:    desc,
			search:  strings.ToLower(title + " " + desc + " " + search),
			enabled: enabled,
			run:     run,
		})
	}

	add("Reload", "Rescan skills and MCP on disk", "reload rescan refresh ctrl+r", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.TeaModel(), h.BeginScanAllCmd()
	})
	add("Filter list", "Fuzzy filter in the skills/MCP list", "find filter search ctrl+f /", caps.Find, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.StartListFilter()
	})
	add("Agent filter", "Choose which agent context to show", "agent filter ctrl+a", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.OpenAgentFilter()
	})
	add("Command reference", "Show help and slash commands", "help commands f1 ?", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.OpenHelpScreen()
	})
	add("Focus list", "Return to the extension list", "list home ctrl+l", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.GoToListingWithPreview()
	})
	add("Tab: Skills", "Switch to the Skills panel", "tab skills", h.ActiveTab() != panel.TabSkills, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.TeaModel(), h.SetActiveTab(panel.TabSkills)
	})
	add("Tab: MCP", "Switch to the MCP panel", "tab mcp", h.ActiveTab() != panel.TabMCP, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.TeaModel(), h.SetActiveTab(panel.TabMCP)
	})

	if caps.SearchInstall && h.ActiveTab() == panel.TabSkills {
		add("Search & install", "Search skills.sh and install", "install add registry ctrl+d", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
			return h.StartInstallFlow()
		})
	}
	if caps.Init && h.ActiveTab() == panel.TabSkills {
		add("New skill template", "Create SKILL.md scaffold", "init new ctrl+n", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
			return h.ShowInitPrompt()
		})
	}
	if caps.Update && h.ActiveTab() == panel.TabSkills {
		add("Update skills", "Update selected or all local skills", "update ctrl+u", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
			return h.HandleUpdate()
		})
	}

	if sel, ok := h.SelectedListItem(); ok {
		if caps.Inspect && sel.CanInspect() {
			add("Inspect", "Browse or preview "+sel.Title, "inspect enter tree", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
				return h.HandleInspectSelected()
			})
		}
		if caps.Bind && sel.CanBind() {
			add("Bind agents", "Manage agent bindings for "+sel.Title, "bind b", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
				return h.HandleBindSelected()
			})
		}
		if caps.Disable && sel.CanDisable() {
			add("Toggle disable", "Enable or disable "+sel.Title, "toggle disable x", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
				return h.HandleDisableSelected()
			})
		}
		if caps.Remove && sel.CanRemove() {
			add("Remove", "Delete "+sel.Title+" (confirmed)", "remove delete del", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
				return h.HandleRemoveSelected()
			})
		}
	}

	for _, spec := range h.CommandSpecs() {
		if !spec.Implemented {
			continue
		}
		spec := spec
		enabled := true
		switch spec.Name {
		case "find":
			enabled = caps.Find
		case "update":
			enabled = caps.Update && h.ActiveTab() == panel.TabSkills
		case "init":
			enabled = caps.Init && h.ActiveTab() == panel.TabSkills
		case "inspect":
			enabled = caps.Inspect && h.ActiveTab() == panel.TabSkills
		case "remove":
			enabled = caps.Remove
		case "add":
			enabled = caps.SearchInstall && h.ActiveTab() == panel.TabSkills
		}
		add(spec.Name, spec.Summary, spec.Usage+" "+strings.Join(spec.Aliases, " "), enabled, func(h paletteActionHost) (tea.Model, tea.Cmd) {
			return h.RunRegistryCommand(spec.Name)
		})
	}

	add("Quit", "Exit skill-man", "quit exit ctrl+c q", true, func(h paletteActionHost) (tea.Model, tea.Cmd) {
		return h.TeaModel(), tea.Quit
	})

	return items
}
