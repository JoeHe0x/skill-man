package bind

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
)

type bindSession struct {
	skill      *skilldomain.Skill
	mcp        *mcpdomain.Server
	mcpMembers []*mcpdomain.Server
	agents     []usecasebind.Choice
}

func (s *bindSession) clear() {
	s.skill = nil
	s.mcp = nil
	s.mcpMembers = nil
	s.agents = nil
}
