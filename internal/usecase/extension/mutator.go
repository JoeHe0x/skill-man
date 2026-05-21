package extension

import (
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// Mutator runs extension mutation use cases against configured services.
type Mutator struct {
	Skills manager.ExtensionManager[*skill.Skill]
	MCP    *servicemcp.Manager
	CWD    string
	Home   string
}

// NewMutator wires dependencies for extension mutations.
func NewMutator(
	skills manager.ExtensionManager[*skill.Skill],
	mcp *servicemcp.Manager,
	cwd, home string,
) Mutator {
	return Mutator{Skills: skills, MCP: mcp, CWD: cwd, Home: home}
}
