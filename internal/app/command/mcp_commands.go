package command

import (
	"context"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

// RemoveMCPKey removes every config-file instance of an MCP key.
type RemoveMCPKey struct {
	Members []*mcpdomain.Server
	Mutator usecase.Mutator
}

func (c *RemoveMCPKey) Label() string {
	key := ""
	if len(c.Members) > 0 {
		if k := c.Members[0].ConfigKey; k != "" {
			key = k
		} else {
			key = c.Members[0].GetName()
		}
	}
	return "removed MCP " + key
}

func (c *RemoveMCPKey) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.RemoveMCPKey(ctx, c.Members))
}

// ToggleDisableMCPKey toggles the disabled state for every member of an MCP key.
type ToggleDisableMCPKey struct {
	Members []*mcpdomain.Server
	Mutator usecase.Mutator
}

func (c *ToggleDisableMCPKey) Label() string {
	disabled := true
	for _, srv := range c.Members {
		if !srv.AggregatedDisabled() {
			disabled = false
			break
		}
	}
	key := ""
	if len(c.Members) > 0 {
		if k := c.Members[0].ConfigKey; k != "" {
			key = k
		} else {
			key = c.Members[0].GetName()
		}
	}
	if disabled && len(c.Members) > 0 {
		return "disabled MCP " + key
	}
	return "enabled MCP " + key
}

func (c *ToggleDisableMCPKey) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.ToggleDisableMCPKey(ctx, c.Members))
}
