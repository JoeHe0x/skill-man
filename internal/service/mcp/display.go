package mcp

import (
	"fmt"
	"path/filepath"
	"strings"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

// ListTitle returns the left-panel title, with a merge badge when aggregated.
func ListTitle(srv *mcpdomain.Server) string {
	name := srv.GetName()
	n := srv.BindingCount()
	if n <= 1 {
		if srv.AllScopesGlobal() {
			return name + " [global]"
		}
		return name
	}
	return fmt.Sprintf("%s [×%d merged]", name, n)
}

// ListBindingDetailLines returns one indented summary line per config binding.
func ListBindingDetailLines(srv *mcpdomain.Server, home string) []string {
	bindings := srv.AllBindings()
	if len(bindings) <= 1 {
		return nil
	}
	lines := make([]string, 0, len(bindings))
	for _, b := range bindings {
		lines = append(lines, formatBindingLine(b, home))
	}
	return lines
}

// ListDesc returns the description line for a list row.
func ListDesc(srv *mcpdomain.Server, home string) string {
	path := ShortPath(home, srv.ConfigPath)
	n := srv.BindingCount()
	if n > 1 {
		keys := uniqueConfigKeys(srv)
		if len(keys) == 1 {
			return fmt.Sprintf("%s · key %q in %d files", path, keys[0], n)
		}
		return fmt.Sprintf("%s · %d files · keys: %s", path, n, strings.Join(keys, ", "))
	}
	desc := srv.GetDescription()
	if srv.ConfigKey != "" && srv.ConfigKey != srv.GetName() {
		return fmt.Sprintf("%s · key: %s · %s", path, srv.ConfigKey, desc)
	}
	if path != "" {
		return fmt.Sprintf("%s · %s", path, desc)
	}
	return desc
}

// ListMeta returns the meta line for a list row.
func ListMeta(srv *mcpdomain.Server) string {
	transport := "stdio"
	if srv.URL != "" {
		transport = "url"
	}
	meta := fmt.Sprintf("%s | agents: %s | %s | %s",
		srv.FormatScopes(),
		joinOrNone(srv.GetAgents()),
		transport,
		srv.GetUpdatedAt().Format("2006-01-02"),
	)
	if roots := WorkspaceRoots(srv); len(roots) == 1 {
		meta += " | " + roots[0]
	} else if len(roots) > 1 {
		meta += " | " + strings.Join(roots, ", ")
	}
	return meta
}

func formatBindingLine(b mcpdomain.Binding, home string) string {
	source := ConfigSourceLabel(b.ConfigPath)
	scope := string(b.Scope)
	path := ShortPath(home, b.ConfigPath)
	agents := joinOrNone(b.Agents)
	prefix := "  ▸ "
	if b.Disabled {
		prefix = "  ▸ [x] "
	}
	return fmt.Sprintf("%s%s  %-7s  %-7s  %s  (%s)", prefix, agents, source, scope, path, b.ConfigKey)
}

// ConfigSourceLabel maps a config file path to a short agent/tool label.
func ConfigSourceLabel(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, ".cursor"):
		return "cursor"
	case strings.Contains(lower, ".codex"):
		return "codex"
	case strings.Contains(lower, "windsurf"), strings.Contains(lower, "codeium"):
		return "windsurf"
	case strings.HasSuffix(lower, ".mcp.json"), strings.Contains(lower, ".claude"):
		return "claude"
	default:
		return filepath.Base(filepath.Dir(path))
	}
}

// ShortPath replaces the home directory prefix with ~.
func ShortPath(home, path string) string {
	if home != "" {
		if strings.HasPrefix(path, home) {
			return "~" + strings.TrimPrefix(path, home)
		}
	}
	return path
}

func uniqueConfigKeys(srv *mcpdomain.Server) []string {
	seen := map[string]bool{}
	var keys []string
	for _, b := range srv.AllBindings() {
		if b.ConfigKey == "" || seen[b.ConfigKey] {
			continue
		}
		seen[b.ConfigKey] = true
		keys = append(keys, b.ConfigKey)
	}
	return keys
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "—"
	}
	return strings.Join(values, ", ")
}
