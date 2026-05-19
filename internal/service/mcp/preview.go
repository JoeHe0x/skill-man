package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	"github.com/JoeHe0x/skill-man/internal/render"
)

// RenderPreview returns a glamour-rendered markdown summary of an MCP server for the TUI viewport.
func RenderPreview(server mcpdomain.Server, width int) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# MCP: %s\n\n", server.GetName())
	bindings := server.AllBindings()
	if len(bindings) > 1 {
		fmt.Fprintf(&b, "> **Merged** — %d config files share this server (`%s`)\n\n", len(bindings), server.GetName())
	}

	if len(bindings) <= 1 {
		if server.ConfigKey != "" && server.ConfigKey != server.GetName() {
			fmt.Fprintf(&b, "**Config key:** `%s`\n\n", server.ConfigKey)
		}
		fmt.Fprintf(&b, "**Scope:** %s\n\n", server.FormatScopes())
		if len(server.GetAgents()) > 0 {
			fmt.Fprintf(&b, "**Agents:** %s\n\n", strings.Join(server.GetAgents(), ", "))
		}
		fmt.Fprintf(&b, "**Config:** `%s`\n\n", server.ConfigPath)
		writeTransport(&b, server.Command, server.Args, server.URL)
		appendConfigSnippet(&b, server.ConfigPath)
		return render.Markdown(b.String(), width)
	}

	fmt.Fprintf(&b, "**Scope:** %s\n\n", server.FormatScopes())
	if len(server.GetAgents()) > 0 {
		fmt.Fprintf(&b, "**Agents:** %s\n\n", strings.Join(server.GetAgents(), ", "))
	}
	fmt.Fprintf(&b, "**Status:** %s\n\n", disabledLabel(server.AggregatedDisabled()))

	for i, binding := range bindings {
		fmt.Fprintf(&b, "## Binding %d\n\n", i+1)
		if binding.ConfigKey != "" && binding.ConfigKey != server.GetName() {
			fmt.Fprintf(&b, "- **Config key:** `%s`\n", binding.ConfigKey)
		}
		fmt.Fprintf(&b, "- **Scope:** %s\n", binding.Scope)
		if len(binding.Agents) > 0 {
			fmt.Fprintf(&b, "- **Agents:** %s\n", strings.Join(binding.Agents, ", "))
		}
		fmt.Fprintf(&b, "- **Config:** `%s`\n", binding.ConfigPath)
		fmt.Fprintf(&b, "- **Status:** %s\n", disabledLabel(binding.Disabled))
		writeTransport(&b, binding.Command, binding.Args, binding.URL)
		appendConfigSnippet(&b, binding.ConfigPath)
	}

	return render.Markdown(b.String(), width)
}

func disabledLabel(disabled bool) string {
	if disabled {
		return "disabled"
	}
	return "enabled"
}

func writeTransport(b *strings.Builder, command string, args []string, url string) {
	switch {
	case url != "":
		fmt.Fprintf(b, "\n**Transport:** URL\n\n`%s`\n\n", url)
	case command != "":
		fmt.Fprintf(b, "\n**Transport:** stdio\n\n```\n%s", command)
		if len(args) > 0 {
			fmt.Fprintf(b, " %s", strings.Join(args, " "))
		}
		b.WriteString("\n```\n\n")
	default:
		b.WriteString("\n**Transport:** unknown\n\n")
	}
}

func appendConfigSnippet(b *strings.Builder, configPath string) {
	if filepath.Base(configPath) != "mcp.json" && filepath.Base(configPath) != "mcp_config.json" {
		return
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return // best-effort preview: skip if config unreadable
	}
	var pretty map[string]any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return // best-effort preview: skip if config unparseable
	}
	formatted, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return // best-effort preview: skip if config unformattable
	}
	fmt.Fprintf(b, "### Config file\n\n```json\n%s\n```\n\n", string(formatted))
}
