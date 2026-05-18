package mcp

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type codexConfigFile struct {
	MCPServers map[string]codexServerConfig `toml:"mcp_servers"`
}

type codexServerConfig struct {
	Command string            `toml:"command" json:"command"`
	Args    []string          `toml:"args" json:"args"`
	URL     string            `toml:"url" json:"url"`
	Env     map[string]string `toml:"env" json:"env"`
	Enabled *bool             `toml:"enabled" json:"enabled"`
}

// ParseCodexConfigFile reads a Codex config.toml and returns MCP servers from [mcp_servers.*] tables.
func ParseCodexConfigFile(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read codex config: %w", err)
	}

	var cfg codexConfigFile
	if err := toml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse codex config: %w", err)
	}
	if len(cfg.MCPServers) == 0 {
		return nil, nil
	}

	agents := agentIDs
	if len(agents) == 0 {
		agents = []string{"codex"}
	}

	entries := make(map[string]serverConfig, len(cfg.MCPServers))
	for name, sc := range cfg.MCPServers {
		disabled := sc.Enabled != nil && !*sc.Enabled
		entries[name] = serverConfig{
			Command:  sc.Command,
			Args:     sc.Args,
			URL:      sc.URL,
			Env:      sc.Env,
			Disabled: disabled,
		}
	}

	return serversFromEntries(entries, filePath, "", scope, agents, projectRoot, home)
}
