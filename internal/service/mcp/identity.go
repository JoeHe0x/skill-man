package mcp

import (
	"net/url"
	"path/filepath"
	"strings"
)

// InferImplementationName derives a stable display name from transport fields.
// It prefers npm package names (e.g. server-filesystem) over config aliases.
func InferImplementationName(command string, args []string, rawURL string) string {
	if rawURL != "" {
		if name := nameFromURL(rawURL); name != "" {
			return name
		}
	}
	if pkg := npmPackageFromArgs(args); pkg != "" {
		return shortPackageName(pkg)
	}
	for _, arg := range args {
		base := filepath.Base(arg)
		switch {
		case strings.HasSuffix(base, ".py"), strings.HasSuffix(base, ".js"), strings.HasSuffix(base, ".ts"):
			return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(base, ".py"), ".js"), ".ts")
		case strings.Contains(base, "mcp") && !strings.HasPrefix(arg, "-"):
			return base
		}
	}
	if command != "" && !isLauncher(command) {
		base := filepath.Base(command)
		if strings.Contains(base, "mcp") || strings.Contains(base, "server-") {
			return base
		}
	}
	return ""
}

// WorkspaceRootFromArgs returns the last non-flag argument after an npm package arg, when present.
func WorkspaceRootFromArgs(args []string) string {
	foundPackage := false
	var root string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if npmPackageFromArgs([]string{arg}) != "" {
			foundPackage = true
			continue
		}
		if foundPackage {
			root = arg
		}
	}
	return root
}

func isLauncher(command string) bool {
	switch filepath.Base(command) {
	case "npx", "uvx", "uv", "node", "deno", "bun", "pnpm", "yarn":
		return true
	default:
		return false
	}
}

func npmPackageFromArgs(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if strings.HasPrefix(arg, "@") && strings.Contains(arg, "/") {
			return arg
		}
		if strings.Contains(arg, "/") && (strings.Contains(arg, "server-") || strings.Contains(arg, "mcp")) {
			return arg
		}
	}
	return ""
}

func shortPackageName(pkg string) string {
	pkg = strings.TrimPrefix(pkg, "@")
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		return pkg[i+1:]
	}
	return pkg
}

func nameFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	host := u.Hostname()
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return host
}
