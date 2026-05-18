package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

type claudeJSONFile struct {
	MCPServers map[string]serverConfig      `json:"mcpServers"`
	Projects   map[string]claudeJSONProject `json:"projects"`
}

type claudeJSONProject struct {
	MCPServers map[string]serverConfig `json:"mcpServers"`
}

// ParseClaudeJSON reads ~/.claude.json including per-project mcpServers entries.
func ParseClaudeJSON(filePath, projectRoot, home string) ([]*mcpdomain.Server, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read claude config: %w", err)
	}

	var cfg claudeJSONFile
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse claude config: %w", err)
	}

	agents := []string{"claude-code"}
	var servers []*mcpdomain.Server

	if len(cfg.MCPServers) > 0 {
		batch, err := serversFromEntries(cfg.MCPServers, filePath, "", extension.ScopeGlobal, agents, projectRoot, home)
		if err != nil {
			return nil, err
		}
		servers = append(servers, batch...)
	}

	cleanRoot := ""
	if projectRoot != "" {
		if abs, err := filepath.Abs(projectRoot); err == nil {
			cleanRoot = filepath.Clean(abs)
		} else {
			cleanRoot = filepath.Clean(projectRoot)
		}
	}

	for projPath, proj := range cfg.Projects {
		if len(proj.MCPServers) == 0 {
			continue
		}
		if cleanRoot != "" && !pathsReferToSameProject(projPath, cleanRoot) {
			continue
		}
		scope := extension.ScopeGlobal
		if cleanRoot != "" {
			scope = extension.ScopeProject
		}
		batch, err := serversFromEntries(proj.MCPServers, filePath, projPath, scope, agents, projectRoot, home)
		if err != nil {
			return nil, err
		}
		servers = append(servers, batch...)
	}

	return servers, nil
}

func pathsReferToSameProject(configuredPath, projectRoot string) bool {
	a := filepath.Clean(configuredPath)
	b := filepath.Clean(projectRoot)
	if a == b {
		return true
	}
	if resolvedA, err := filepath.EvalSymlinks(a); err == nil {
		a = resolvedA
	}
	if resolvedB, err := filepath.EvalSymlinks(b); err == nil {
		b = resolvedB
	}
	return a == b
}
