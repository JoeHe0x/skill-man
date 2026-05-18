package app

import (
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	"github.com/JoeHe0x/skill-man/internal/service/skill"
)

func newPanelRegistry() *panel.Registry {
	return panel.NewRegistry(
		panel.SkillDeps{Manager: manager.NewManager[*skilldomain.Skill](skill.SkillScanStrategy{})},
		panel.MCPDeps{},
	)
}

func (m *Model) activePanel() panel.Panel {
	return m.panels.Get(m.activeTab)
}
