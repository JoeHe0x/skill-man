package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type configFile struct {
	MCPServers map[string]serverConfig `json:"mcpServers"`
}

type serverConfig struct {
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	URL       string            `json:"url"`
	ServerURL string            `json:"serverUrl"`
	Env       map[string]string `json:"env"`
	Disabled  bool              `json:"disabled"`
}

// ParseConfigFile reads an MCP config file and returns one Server per mcpServers entry.
func ParseConfigFile(filePath, projectRoot, home string, scope extension.Scope) ([]*mcpdomain.Server, error) {
	return ParseConfigFileForAgents(filePath, projectRoot, home, scope, nil)
}

// ParseConfigFileForAgents is like ParseConfigFile but pins agent IDs when auto-resolve is insufficient.
func ParseConfigFileForAgents(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read mcp config: %w", err)
	}

	var cfg configFile
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse mcp config: %w", err)
	}

	return serversFromEntries(cfg.MCPServers, filePath, "", scope, agentIDs, projectRoot, home)
}

func serverURL(sc serverConfig) string {
	if sc.URL != "" {
		return sc.URL
	}
	return sc.ServerURL
}

func describeServer(sc serverConfig) string {
	switch {
	case serverURL(sc) != "":
		return "url: " + serverURL(sc)
	case sc.Command != "":
		if len(sc.Args) > 0 {
			return sc.Command + " " + strings.Join(sc.Args, " ")
		}
		return sc.Command
	default:
		return "MCP server entry"
	}
}

func serverID(filePath, name string) string {
	base := filepath.Base(filePath) + ":" + name
	if len(base) > 16 {
		return base[:16]
	}
	return base
}
