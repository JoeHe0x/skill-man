package bind

import (
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// Binder runs agent bind/unbind use cases.
type Binder struct {
	Skills manager.ExtensionManager[*skill.Skill]
	MCP    *servicemcp.Manager
	CWD    string
	Home   string
}

// NewBinder wires dependencies for bind flows.
func NewBinder(
	skills manager.ExtensionManager[*skill.Skill],
	mcp *servicemcp.Manager,
	cwd, home string,
) Binder {
	return Binder{Skills: skills, MCP: mcp, CWD: cwd, Home: home}
}
