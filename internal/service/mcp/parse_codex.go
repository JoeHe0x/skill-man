package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/pelletier/go-toml/v2"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type codexConfigFile struct {
	MCPServers map[string]codexServerConfig `toml:"mcp_servers"`
}

type codexServerConfig struct {
	Command string            `toml:"command,omitempty" json:"command,omitempty"`
	Args    []string          `toml:"args,omitempty" json:"args,omitempty"`
	URL     string            `toml:"url,omitempty" json:"url,omitempty"`
	Env     map[string]string `toml:"env,omitempty" json:"env,omitempty"`
	Enabled *bool             `toml:"enabled,omitempty" json:"enabled,omitempty"`
}

// sanitizeCodexServer keeps Codex transport fields mutually exclusive (stdio vs url).
func sanitizeCodexServer(sc codexServerConfig) codexServerConfig {
	if sc.Command != "" || len(sc.Args) > 0 {
		sc.URL = ""
		return sc
	}
	if sc.URL != "" {
		sc.Command = ""
		sc.Args = nil
	}
	return sc
}

func sanitizeCodexConfig(cfg *codexConfigFile) {
	if cfg == nil || cfg.MCPServers == nil {
		return
	}
	for key, sc := range cfg.MCPServers {
		cfg.MCPServers[key] = sanitizeCodexServer(sc)
	}
}

// RepairCodexConfigFile removes invalid url fields from stdio MCP servers.
// Codex rejects stdio entries that include url (even url = "").
// See https://developers.openai.com/codex/config-reference — command/args = stdio, url = HTTP.
func RepairCodexConfigFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("read codex config: %w", err)
	}

	var root map[string]any
	if len(content) > 0 {
		if err := toml.Unmarshal(content, &root); err != nil {
			return false, fmt.Errorf("parse codex config: %w", err)
		}
	} else {
		return false, nil
	}

	serversRaw, ok := root["mcp_servers"]
	if !ok {
		return false, nil
	}
	servers, ok := serversRaw.(map[string]any)
	if !ok {
		return false, nil
	}

	changed := false
	for key, raw := range servers {
		var sc codexServerConfig
		b, err := json.Marshal(raw)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &sc); err == nil {
			sanitized := sanitizeCodexServer(sc)

			sb, err := json.Marshal(sanitized)
			if err != nil {
				continue
			}
			var sanitizedMap map[string]any
			if err := json.Unmarshal(sb, &sanitizedMap); err != nil {
				continue
			}

			if reflect.DeepEqual(raw, sanitizedMap) {
				continue
			}

			servers[key] = sanitizedMap
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	root["mcp_servers"] = servers
	out, err := toml.Marshal(root)
	if err != nil {
		return false, fmt.Errorf("marshal codex config: %w", err)
	}
	if bytes.Equal(bytes.TrimSpace(content), bytes.TrimSpace(out)) {
		return false, nil
	}
	if err := os.WriteFile(filePath, out, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// ParseCodexConfigFile reads a Codex config.toml and returns MCP servers from [mcp_servers.*] tables.
func ParseCodexConfigFile(filePath, projectRoot, home string, scope extension.Scope, agentIDs []string) ([]*mcpdomain.Server, error) {
	if _, err := RepairCodexConfigFile(filePath); err != nil {
		return nil, err
	}

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
