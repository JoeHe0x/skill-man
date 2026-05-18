package mcp

import (
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// Binding is one MCP server entry in a specific config file.
type Binding struct {
	ConfigPath  string
	ConfigKey   string
	Scope       extension.Scope
	Agents      []string
	Disabled    bool
	Command     string
	Args        []string
	URL         string
	Description string
}

// Server represents an MCP server, possibly aggregated across multiple config files.
type Server struct {
	extension.BaseExtension
	ConfigKey string
	Command   string
	Args      []string
	URL       string
	Bindings  []Binding
}

// AllBindings returns explicit bindings or a single synthetic binding from top-level fields.
func (s *Server) AllBindings() []Binding {
	if len(s.Bindings) > 0 {
		return s.Bindings
	}
	if s.ConfigPath == "" {
		return nil
	}
	return []Binding{{
		ConfigPath:  s.ConfigPath,
		ConfigKey:   s.ConfigKey,
		Scope:       s.Scope,
		Agents:      append([]string(nil), s.Agents...),
		Disabled:    s.Disabled,
		Command:     s.Command,
		Args:        append([]string(nil), s.Args...),
		URL:         s.URL,
		Description: s.Description,
	}}
}

// WithBinding returns a single-binding view for config mutations.
func (s *Server) WithBinding(b Binding) *Server {
	clone := *s
	clone.ConfigPath = b.ConfigPath
	clone.ConfigKey = b.ConfigKey
	clone.Scope = b.Scope
	clone.Agents = append([]string(nil), b.Agents...)
	clone.Disabled = b.Disabled
	clone.Command = b.Command
	clone.Args = append([]string(nil), b.Args...)
	clone.URL = b.URL
	clone.Description = b.Description
	clone.Bindings = nil
	return &clone
}

// BindingCount is the number of config file entries represented by this server.
func (s *Server) BindingCount() int {
	return len(s.AllBindings())
}

// AggregatedDisabled is true when every binding is disabled.
func (s *Server) AggregatedDisabled() bool {
	bindings := s.AllBindings()
	if len(bindings) == 0 {
		return s.Disabled
	}
	for _, b := range bindings {
		if !b.Disabled {
			return false
		}
	}
	return true
}

// FormatScopes returns a sorted, comma-separated scope label for list meta.
func (s *Server) FormatScopes() string {
	bindings := s.AllBindings()
	if len(bindings) == 0 {
		return string(s.Scope)
	}
	seen := map[extension.Scope]bool{}
	var scopes []string
	for _, b := range bindings {
		if seen[b.Scope] {
			continue
		}
		seen[b.Scope] = true
		scopes = append(scopes, string(b.Scope))
	}
	return strings.Join(scopes, ", ")
}

// AllScopesGlobal reports whether every binding uses global scope.
func (s *Server) AllScopesGlobal() bool {
	bindings := s.AllBindings()
	if len(bindings) == 0 {
		return s.Scope == extension.ScopeGlobal
	}
	for _, b := range bindings {
		if b.Scope != extension.ScopeGlobal {
			return false
		}
	}
	return true
}

// SyncAggregatedFields updates top-level Agents and Disabled from bindings.
func (s *Server) SyncAggregatedFields() {
	bindings := s.AllBindings()
	if len(bindings) == 0 {
		return
	}
	if len(bindings) == 1 {
		b := bindings[0]
		s.ConfigPath = b.ConfigPath
		s.ConfigKey = b.ConfigKey
		s.Scope = b.Scope
		s.Agents = append([]string(nil), b.Agents...)
		s.Disabled = b.Disabled
		s.Command = b.Command
		s.Args = append([]string(nil), b.Args...)
		s.URL = b.URL
		s.Description = b.Description
		return
	}

	agentSet := map[string]bool{}
	for _, b := range bindings {
		for _, id := range b.Agents {
			agentSet[id] = true
		}
	}
	s.Agents = make([]string, 0, len(agentSet))
	for id := range agentSet {
		s.Agents = append(s.Agents, id)
	}
	s.Disabled = true
	for _, b := range bindings {
		if !b.Disabled {
			s.Disabled = false
			break
		}
	}

	b0 := bindings[0]
	s.ConfigPath = b0.ConfigPath
	s.ConfigKey = b0.ConfigKey
	s.Command = b0.Command
	s.Args = append([]string(nil), b0.Args...)
	s.URL = b0.URL
	s.Description = b0.Description
}
