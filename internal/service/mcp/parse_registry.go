package mcp

import (
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type configParseFn func(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error)

var configFileParsers = map[string]configParseFn{
	"config.toml":  parseCodexAdapter,
	".claude.json": parseClaudeConfigFile,
}

func parseCodexAdapter(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	return ParseCodexConfigFile(filePath, projectRoot, home, scope, agentIDs)
}

func parseClaudeConfigFile(filePath, projectRoot, home string, _ extension.Scope, _ []string) ([]*mcpdomain.Server, error) {
	return ParseClaudeJSON(filePath, projectRoot, home)
}

// ParseConfigAtPath parses an MCP config file using the registered parser for its basename.
func ParseConfigAtPath(filePath, projectRoot, home string, scope extension.Scope) ([]*mcpdomain.Server, error) {
	return ParseConfigAtPathForAgents(filePath, projectRoot, home, scope, nil)
}

// ParseConfigAtPathForAgents parses a config file and optionally pins agent IDs.
func ParseConfigAtPathForAgents(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	key := strings.ToLower(filepath.Base(filePath))
	if fn, ok := configFileParsers[key]; ok {
		return fn(filePath, projectRoot, home, scope, agentIDs)
	}
	return ParseConfigFileForAgents(filePath, projectRoot, home, scope, agentIDs)
}
