package install

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

// Provider searches a registry and installs candidates for one extension kind.
// MCP will implement this interface later; skills uses the skills CLI today.
type Provider interface {
	Kind() domaininstall.Kind
	Search(query string) ([]domaininstall.Candidate, error)
	Install(ctx context.Context, cwd, home string, candidate domaininstall.Candidate, agentIDs []string, scope extension.Scope) (installedName string, err error)
	SupportedAgents() []agent.Agent
}
