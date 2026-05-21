package extension

import (
	"context"
	"fmt"

	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
)

// RemoveMCPKey removes every config-file instance of an MCP key.
func (m Mutator) RemoveMCPKey(ctx context.Context, members []*mcpdomain.Server) Outcome {
	var errs []error
	for _, s := range members {
		if err := m.MCP.Remove(s); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return Outcome{Kind: KindMCP, Err: joinErrs(errs)}
	}
	key := mcpKey(members)
	return Outcome{
		Kind:         KindMCP,
		AffectedName: key,
		Message:      fmt.Sprintf("removed MCP `%s` (%d locations)", key, len(members)),
	}
}

// ToggleDisableMCPKey toggles the disabled state for every member of an MCP key.
func (m Mutator) ToggleDisableMCPKey(ctx context.Context, members []*mcpdomain.Server) Outcome {
	var errs []error
	for _, s := range members {
		if err := m.MCP.ToggleDisable(s); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return Outcome{Kind: KindMCP, Err: joinErrs(errs)}
	}
	label := "disabled"
	if !mcpKeyDisabled(members) {
		label = "enabled"
	}
	key := mcpKey(members)
	return Outcome{
		Kind:         KindMCP,
		AffectedName: key,
		Message:      fmt.Sprintf("%s MCP `%s` (%d locations)", label, key, len(members)),
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
