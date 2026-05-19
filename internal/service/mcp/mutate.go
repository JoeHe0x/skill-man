package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/pelletier/go-toml/v2"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

// Manager performs MCP config mutations (toggle, remove, bind).
type Manager struct{}

func NewManager() *Manager { return &Manager{} }

// ToggleDisable enables or disables MCP server entries across all aggregated bindings.
func (m *Manager) ToggleDisable(srv *mcpdomain.Server) error {
	if srv == nil {
		return errors.New("mcp server is nil")
	}
	targetDisabled := !srv.AggregatedDisabled()
	bindings := srv.AllBindings()
	for i, b := range bindings {
		sub := srv.WithBinding(b)
		if sub.Disabled == targetDisabled {
			continue
		}
		if err := toggleDisableOne(sub); err != nil {
			return err
		}
		if len(srv.Bindings) > 0 {
			srv.Bindings[i].Disabled = targetDisabled
		}
	}
	srv.SyncAggregatedFields()
	return nil
}

// Remove deletes server entries from every aggregated config binding.
func (m *Manager) Remove(srv *mcpdomain.Server) error {
	if srv == nil {
		return errors.New("mcp server is nil")
	}
	for _, b := range srv.AllBindings() {
		if err := removeOne(srv.WithBinding(b)); err != nil {
			return err
		}
	}
	return nil
}

func toggleDisableOne(srv *mcpdomain.Server) error {
	switch configFormatForPath(srv.ConfigPath) {
	case formatTOML:
		return toggleCodexServer(srv)
	default:
		return toggleJSONServer(srv)
	}
}

func removeOne(srv *mcpdomain.Server) error {
	switch configFormatForPath(srv.ConfigPath) {
	case formatTOML:
		return removeCodexServer(srv)
	default:
		return removeJSONServer(srv)
	}
}

// Bind copies a server definition into the target agent's MCP config.
func (m *Manager) Bind(srv *mcpdomain.Server, target agent.Agent, projectRoot, home string) error {
	clone := bindTargetView(srv, target, projectRoot, home)
	return m.BindAt(srv, target, clone.Scope, projectRoot, home)
}

// BindAt writes the server into a specific agent config file at the given scope.
func (m *Manager) BindAt(srv *mcpdomain.Server, target agent.Agent, scope extension.Scope, projectRoot, home string) error {
	if srv == nil {
		return errors.New("mcp server is nil")
	}
	clone := bindTargetView(srv, target, projectRoot, home)
	clone.Scope = scope
	targetPath := targetConfigPath(target, scope, projectRoot, home)
	if targetPath == "" {
		return fmt.Errorf("agent %s has no MCP config path for scope %s", target.ID, scope)
	}
	entry, err := exportServerEntry(targetPath, &clone)
	if err != nil {
		return err
	}
	return mergeServerEntry(targetPath, clone.ConfigKey, entry, scope, projectRoot, home)
}

// Unbind removes a server entry from the target agent's MCP config when present.
func (m *Manager) Unbind(srv *mcpdomain.Server, target agent.Agent, projectRoot, home string) error {
	clone := bindTargetView(srv, target, projectRoot, home)
	return m.UnbindAt(srv, target, clone.Scope, projectRoot, home)
}

// UnbindAt removes the server from a specific agent config file at the given scope.
func (m *Manager) UnbindAt(srv *mcpdomain.Server, target agent.Agent, scope extension.Scope, projectRoot, home string) error {
	if srv == nil {
		return errors.New("mcp server is nil")
	}
	clone := bindTargetView(srv, target, projectRoot, home)
	targetPath := targetConfigPath(target, scope, projectRoot, home)
	if targetPath == "" {
		return nil
	}
	if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	clone.ConfigPath = targetPath
	clone.Scope = scope
	return removeOne(&clone)
}

// bindTargetView picks config key, scope, and transport for a specific agent.
func bindTargetView(srv *mcpdomain.Server, target agent.Agent, projectRoot, home string) mcpdomain.Server {
	clone := *srv
	for _, b := range srv.AllBindings() {
		if !slices.Contains(b.Agents, target.ID) {
			continue
		}
		clone.ConfigKey = b.ConfigKey
		clone.Scope = b.Scope
		if b.Command != "" {
			clone.Command = b.Command
		}
		if len(b.Args) > 0 {
			clone.Args = append([]string(nil), b.Args...)
		}
		if b.URL != "" {
			clone.URL = b.URL
		}
		return clone
	}
	clone.Scope = defaultBindScope(target, projectRoot, home)
	return clone
}

func defaultBindScope(target agent.Agent, projectRoot, home string) extension.Scope {
	if target.ID == "claude-code" && projectRoot != "" {
		return extension.ScopeProject
	}
	if home != "" {
		return extension.ScopeGlobal
	}
	return extension.ScopeProject
}

func exportServerEntry(configPath string, srv *mcpdomain.Server) (map[string]any, error) {
	if configFormatForPath(configPath) == formatTOML {
		return exportCodexServerEntry(srv), nil
	}
	entry := map[string]any{}
	if srv.Command != "" {
		entry["command"] = srv.Command
	}
	if len(srv.Args) > 0 {
		entry["args"] = srv.Args
	}
	if srv.URL != "" {
		entry["url"] = srv.URL
	}
	if srv.IsDisabled() {
		entry["disabled"] = true
	}

	if filepath.Base(filepath.Dir(configPath)) == ".cursor" {
		if srv.URL != "" {
			entry["type"] = "sse"
		} else if srv.Command != "" {
			entry["type"] = "stdio"
		}
	}

	if len(entry) == 0 {
		return nil, fmt.Errorf("server %q has no transport fields to bind", srv.GetName())
	}
	return entry, nil
}

func toggleJSONServer(srv *mcpdomain.Server) error {
	if filepath.Base(srv.ConfigPath) == ".claude.json" {
		return toggleClaudeJSONServer(srv)
	}
	return editJSONObject(srv.ConfigPath, func(servers map[string]json.RawMessage) error {
		raw, ok := servers[srv.ConfigKey]
		if !ok {
			return fmt.Errorf("server %q not found in %s", srv.ConfigKey, srv.ConfigPath)
		}
		var entry map[string]any
		if err := json.Unmarshal(raw, &entry); err != nil {
			return err
		}
		if srv.IsDisabled() {
			delete(entry, "disabled")
		} else {
			entry["disabled"] = true
		}
		updated, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		servers[srv.ConfigKey] = updated
		return nil
	})
}

func removeJSONServer(srv *mcpdomain.Server) error {
	if filepath.Base(srv.ConfigPath) == ".claude.json" {
		return removeClaudeJSONServer(srv)
	}
	if _, err := os.Stat(srv.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return editJSONObject(srv.ConfigPath, func(servers map[string]json.RawMessage) error {
		if _, ok := servers[srv.ConfigKey]; !ok {
			return nil
		}
		delete(servers, srv.ConfigKey)
		return nil
	})
}

func editJSONObject(path string, edit func(map[string]json.RawMessage) error) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read mcp config: %w", err)
	}
	var root map[string]json.RawMessage
	if len(bytes.TrimSpace(content)) == 0 {
		root = make(map[string]json.RawMessage)
	} else if err := json.Unmarshal(content, &root); err != nil {
		return fmt.Errorf("parse mcp config: %w", err)
	}
	if root == nil {
		root = make(map[string]json.RawMessage)
	}
	serversRaw, ok := root["mcpServers"]
	var servers map[string]json.RawMessage
	if !ok {
		servers = make(map[string]json.RawMessage)
	} else if err := json.Unmarshal(serversRaw, &servers); err != nil {
		return err
	}
	if err := edit(servers); err != nil {
		return err
	}
	updated, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	root["mcpServers"] = updated
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func mergeServerEntry(path, key string, entry map[string]any, scope extension.Scope, projectRoot, home string) error {
	if filepath.Base(path) == ".claude.json" {
		return mergeClaudeJSONServer(path, key, entry, scope, projectRoot, home)
	}
	if configFormatForPath(path) == formatTOML {
		return mergeCodexServer(path, key, entry)
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return writeNewJSONObject(path, key, raw)
	}
	return editJSONObject(path, func(servers map[string]json.RawMessage) error {
		servers[key] = raw
		return nil
	})
}

func writeNewJSONObject(path, key string, entry json.RawMessage) error {
	root := map[string]any{
		"mcpServers": map[string]json.RawMessage{key: entry},
	}
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func toggleClaudeJSONServer(srv *mcpdomain.Server) error {
	return editClaudeFile(srv.ConfigPath, srv, func(servers map[string]serverConfig) error {
		sc, ok := servers[srv.ConfigKey]
		if !ok {
			return fmt.Errorf("server %q not found", srv.ConfigKey)
		}
		if srv.IsDisabled() {
			sc.Disabled = false
		} else {
			sc.Disabled = true
		}
		servers[srv.ConfigKey] = sc
		return nil
	})
}

func removeClaudeJSONServer(srv *mcpdomain.Server) error {
	if _, err := os.Stat(srv.ConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return editClaudeFile(srv.ConfigPath, srv, func(servers map[string]serverConfig) error {
		if _, ok := servers[srv.ConfigKey]; !ok {
			return nil
		}
		delete(servers, srv.ConfigKey)
		return nil
	})
}

func mergeClaudeJSONServer(path, key string, entry map[string]any, scope extension.Scope, projectRoot, home string) error {
	srv := &mcpdomain.Server{
		BaseExtension: extension.BaseExtension{Path: projectRoot, ConfigPath: path},
		ConfigKey:     key,
	}
	if scope == extension.ScopeGlobal {
		srv.Scope = extension.ScopeGlobal
		return editClaudeFile(path, srv, func(servers map[string]serverConfig) error {
			var sc serverConfig
			b, _ := json.Marshal(entry)
			_ = json.Unmarshal(b, &sc)
			servers[key] = sc
			return nil
		})
	}
	return editClaudeProject(path, projectRoot, func(servers map[string]serverConfig) error {
		var sc serverConfig
		b, _ := json.Marshal(entry)
		_ = json.Unmarshal(b, &sc)
		servers[key] = sc
		return nil
	})
}

func editClaudeFile(path string, srv *mcpdomain.Server, edit func(map[string]serverConfig) error) error {
	if srv.GetScope() == extension.ScopeProject && srv.GetPath() != "" {
		return editClaudeProject(path, filepath.ToSlash(srv.GetPath()), edit)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		content = []byte("{}")
	}
	var cfg claudeJSONFile
	if err := json.Unmarshal(content, &cfg); err != nil {
		return err
	}
	if cfg.MCPServers == nil {
		cfg.MCPServers = map[string]serverConfig{}
	}
	if err := edit(cfg.MCPServers); err != nil {
		return err
	}
	return writeClaudeJSON(path, cfg)
}

func editClaudeProject(path, projectKey string, edit func(map[string]serverConfig) error) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		content = []byte("{}")
	}
	var cfg claudeJSONFile
	if err := json.Unmarshal(content, &cfg); err != nil {
		return err
	}
	if cfg.Projects == nil {
		cfg.Projects = map[string]claudeJSONProject{}
	}
	proj, ok := cfg.Projects[projectKey]
	if !ok {
		proj = claudeJSONProject{MCPServers: map[string]serverConfig{}}
	}
	if proj.MCPServers == nil {
		proj.MCPServers = map[string]serverConfig{}
	}
	if err := edit(proj.MCPServers); err != nil {
		return err
	}
	cfg.Projects[projectKey] = proj
	return writeClaudeJSON(path, cfg)
}

func writeClaudeJSON(path string, cfg claudeJSONFile) error {
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func exportCodexServerEntry(srv *mcpdomain.Server) map[string]any {
	entry := map[string]any{
		"enabled": !srv.IsDisabled(),
	}
	if srv.URL != "" {
		entry["url"] = srv.URL
		return entry
	}
	if srv.Command != "" {
		entry["command"] = srv.Command
	}
	if len(srv.Args) > 0 {
		entry["args"] = srv.Args
	}
	return entry
}

func editTOMLObject(path string, edit func(map[string]any) error) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		content = []byte{}
	}
	var root map[string]any
	if len(content) > 0 {
		if err := toml.Unmarshal(content, &root); err != nil {
			return err
		}
	}
	if root == nil {
		root = make(map[string]any)
	}

	serversRaw, ok := root["mcp_servers"]
	var servers map[string]any
	if ok {
		if m, ok := serversRaw.(map[string]any); ok {
			servers = m
		} else {
			servers = make(map[string]any)
		}
	} else {
		servers = make(map[string]any)
	}

	if err := edit(servers); err != nil {
		return err
	}

	root["mcp_servers"] = servers
	out, err := toml.Marshal(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func toggleCodexServer(srv *mcpdomain.Server) error {
	return editTOMLObject(srv.ConfigPath, func(servers map[string]any) error {
		raw, ok := servers[srv.ConfigKey]
		if !ok {
			return fmt.Errorf("server %q not found in %s", srv.ConfigKey, srv.ConfigPath)
		}
		var sc codexServerConfig
		b, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &sc); err != nil {
			return err
		}
		if srv.IsDisabled() {
			t := true
			sc.Enabled = &t
		} else {
			f := false
			sc.Enabled = &f
		}
		sc = sanitizeCodexServer(sc)

		b2, _ := json.Marshal(sc)
		var scMap map[string]any
		json.Unmarshal(b2, &scMap)
		servers[srv.ConfigKey] = scMap
		return nil
	})
}

func removeCodexServer(srv *mcpdomain.Server) error {
	return editTOMLObject(srv.ConfigPath, func(servers map[string]any) error {
		delete(servers, srv.ConfigKey)
		return nil
	})
}

func mergeCodexServer(path, key string, entry map[string]any) error {
	return editTOMLObject(path, func(servers map[string]any) error {
		var sc codexServerConfig
		b, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &sc); err != nil {
			return err
		}
		if sc.Enabled == nil {
			t := true
			sc.Enabled = &t
		}
		sc = sanitizeCodexServer(sc)

		b2, _ := json.Marshal(sc)
		var scMap map[string]any
		json.Unmarshal(b2, &scMap)
		servers[key] = scMap
		return nil
	})
}
