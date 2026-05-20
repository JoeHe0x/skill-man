package command

import (
	"context"
	"fmt"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// RemoveMCPKey removes every config-file instance of an MCP key.
type RemoveMCPKey struct {
	Members []*mcpdomain.Server
	Manager *servicemcp.Manager
}

func (c *RemoveMCPKey) Label() string {
	key := mcpKey(c.Members)
	return "removed MCP " + key
}

func (c *RemoveMCPKey) Execute(ctx context.Context) Result {
	var errs []error
	for _, s := range c.Members {
		if err := c.Manager.Remove(s); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return Result{Err: joinErrs(errs), TargetTab: panel.TabMCP}
	}
	key := mcpKey(c.Members)
	return Result{
		AffectedName: key,
		Message:      fmt.Sprintf("removed MCP `%s` (%d locations)", key, len(c.Members)),
		TargetTab:    panel.TabMCP,
	}
}

// ToggleDisableMCPKey toggles the disabled state for every member of an MCP key.
type ToggleDisableMCPKey struct {
	Members []*mcpdomain.Server
	Manager *servicemcp.Manager
}

func (c *ToggleDisableMCPKey) Label() string {
	if mcpKeyDisabled(c.Members) {
		return "disabled MCP " + mcpKey(c.Members)
	}
	return "enabled MCP " + mcpKey(c.Members)
}

func (c *ToggleDisableMCPKey) Execute(ctx context.Context) Result {
	var errs []error
	for _, s := range c.Members {
		if err := c.Manager.ToggleDisable(s); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return Result{Err: joinErrs(errs), TargetTab: panel.TabMCP}
	}
	label := "disabled"
	if !mcpKeyDisabled(c.Members) {
		label = "enabled"
	}
	key := mcpKey(c.Members)
	return Result{
		AffectedName: key,
		Message:      fmt.Sprintf("%s MCP `%s` (%d locations)", label, key, len(c.Members)),
		TargetTab:    panel.TabMCP,
	}
}

func mcpKey(members []*mcpdomain.Server) string {
	if len(members) == 0 {
		return ""
	}
	if k := members[0].ConfigKey; k != "" {
		return k
	}
	return members[0].GetName()
}

func mcpKeyDisabled(members []*mcpdomain.Server) bool {
	for _, srv := range members {
		if !srv.AggregatedDisabled() {
			return false
		}
	}
	return len(members) > 0
}

func joinErrs(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%d error(s): %v", len(errs), errs)
}
