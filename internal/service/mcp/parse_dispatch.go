package mcp

import (
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

// ParseConfigAtPath parses an MCP config file based on its name and format.
func ParseConfigAtPath(filePath, projectRoot, home string, scope extension.Scope) ([]*mcpdomain.Server, error) {
	return ParseConfigAtPathForAgents(filePath, projectRoot, home, scope, nil)
}

// ParseConfigAtPathForAgents parses a config file and optionally pins agent IDs.
func ParseConfigAtPathForAgents(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	base := strings.ToLower(filepath.Base(filePath))
	switch base {
	case "config.toml":
		return ParseCodexConfigFile(filePath, projectRoot, home, scope, agentIDs)
	case ".claude.json":
		return ParseClaudeJSON(filePath, projectRoot, home)
	default:
		return ParseConfigFileForAgents(filePath, projectRoot, home, scope, agentIDs)
	}
}
