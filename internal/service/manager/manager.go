package manager

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// ExtensionManager defines the operations common to all extensions
// (Skill, MCP Server, Sub-Agent, Hook).
type ExtensionManager[T extension.Extension] interface {
	// Scan discovers extensions across given agents.
	Scan(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]T, error)

	// Bind associates an extension with a specific agent.
	Bind(ctx context.Context, ext T, a agent.Agent, projectRoot, home string) error

	// Unbind removes the association between an extension and a specific agent.
	Unbind(ctx context.Context, ext T, a agent.Agent, projectRoot, home string) error

	// ToggleDisable toggles the disabled state of an extension.
	ToggleDisable(ctx context.Context, ext T) error

	// Remove deletes the extension.
	Remove(ctx context.Context, ext T, projectRoot, home string) error
}

// Strategy provides the type-specific implementation details for scanning and parsing.
type ScanStrategy[T extension.Extension] interface {
	// DefaultDir returns the default directory name for this extension type (e.g. ".skills", ".mcp")
	DefaultDir() string

	// AgentDir returns the agent-specific directory for this extension type
	AgentDir(a agent.Agent) string

	// SkipDir indicates if a directory should be skipped during scanning
	SkipDir(dirName string) bool

	// TargetFiles returns the list of configuration file names to look for (e.g. "SKILL.md", "mcp.json")
	TargetFiles() []string

	// ParseFile reads and parses the extension configuration
	ParseFile(filePath, projectRoot, home string, scope extension.Scope) (T, error)

	// Dedupe removes duplicate extensions from the scan results
	Dedupe(entities []T) []T
}
