package mcp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

func serversFromEntries(
	mcpServers map[string]serverConfig,
	filePath, entityPath string,
	scope extension.Scope,
	agentIDs []string,
	projectRoot, home string,
) ([]*mcpdomain.Server, error) {
	if len(mcpServers) == 0 {
		return nil, nil
	}

	dir := filepath.Dir(filePath)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolvedDir
	}
	if entityPath != "" {
		dir = entityPath
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	agents := agentIDs
	if len(agents) == 0 {
		agents = agent.ResolveMCPAgentIDs(dir, projectRoot, home)
	}

	servers := make([]*mcpdomain.Server, 0, len(mcpServers))
	for configKey, sc := range mcpServers {
		displayName := InferImplementationName(sc.Command, sc.Args, serverURL(sc))
		if displayName == "" {
			displayName = configKey
		}
		disabled := sc.Disabled || strings.HasSuffix(filePath, ".disabled")
		desc := describeServer(sc)
		servers = append(servers, &mcpdomain.Server{
			BaseExtension: extension.BaseExtension{
				ID:          serverID(filePath, configKey),
				Name:        displayName,
				Description: desc,
				Path:        dir,
				ConfigPath:  filePath,
				UpdatedAt:   info.ModTime(),
				Scope:       scope,
				Agents:      agents,
				Disabled:    disabled,
			},
			ConfigKey: configKey,
			Command:   sc.Command,
			Args:      sc.Args,
			URL:       serverURL(sc),
			Bindings: []mcpdomain.Binding{{
				ConfigPath:  filePath,
				ConfigKey:   configKey,
				Scope:       scope,
				Agents:      append([]string(nil), agents...),
				Disabled:    disabled,
				Command:     sc.Command,
				Args:        append([]string(nil), sc.Args...),
				URL:         serverURL(sc),
				Description: desc,
			}},
		})
	}
	return servers, nil
}
