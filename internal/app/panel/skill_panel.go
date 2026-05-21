package panel

import (
	"context"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
)

type skillPanel struct {
	mgr    manager.ExtensionManager[*skilldomain.Skill]
	skills []*skilldomain.Skill
}

// NewSkillPanel creates the skills extension panel.
func NewSkillPanel(mgr manager.ExtensionManager[*skilldomain.Skill]) Panel {
	return &skillPanel{mgr: mgr}
}

func (p *skillPanel) Tab() Tab { return TabSkills }

func (p *skillPanel) Count() int { return len(p.skills) }

func (p *skillPanel) CountLabel() string { return "skills" }

func (p *skillPanel) Capabilities() Capabilities {
	return Capabilities{
		Inspect:       true,
		Disable:       true,
		Bind:          true,
		Remove:        true,
		Update:        true,
		Find:          true,
		Add:           true,
		Init:          true,
		SearchInstall: true,
	}
}

func (p *skillPanel) Scan(ctx context.Context, cwd, home string, agents []agent.Agent) ScannedMsg {
	skills, err := p.mgr.Scan(ctx, cwd, home, agents)
	return SkillsScan(skills, err)
}

func (p *skillPanel) ApplyScan(msg ScannedMsg) bool {
	if msg.Tab != TabSkills || msg.Err != nil {
		return false
	}
	p.skills = msg.Skills
	return true
}

// Skills returns the last scanned skill list (implements SkillProvider).
func (p *skillPanel) Skills() []*skilldomain.Skill { return p.skills }

func (p *skillPanel) ListItems(agentFilter []string) []Item {
	return skillListItems(p.skills, agentFilter)
}

func (p *skillPanel) SearchItems(query string, agentFilter []string) []Item {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return p.ListItems(agentFilter)
	}
	var results []*skilldomain.Skill
	for _, sk := range p.skills {
		haystack := strings.ToLower(strings.Join([]string{
			sk.GetName(),
			sk.GetDescription(),
			strings.Join(sk.Tools, " "),
			sk.GetPath(),
		}, " "))
		if strings.Contains(haystack, query) {
			results = append(results, sk)
		}
	}
	return skillListItems(results, agentFilter)
}

func (p *skillPanel) PanelTitle(state ViewState) string {
	switch state {
	case ViewSearching:
		return "Search Results"
	case ViewInstalling:
		return "Install Skill"
	case ViewHelp:
		return "Commands"
	case ViewBinding:
		return "Bind Agents"
	case ViewInspecting:
		return "Files"
	default:
		return "Skills"
	}
}

func (p *skillPanel) ReloadHint() string { return "Rescanning local skills..." }

func (p *skillPanel) StaticPreview() string { return "" }

func (p *skillPanel) PreviewMarkdown(selected Item, width int) (string, error) {
	if selected.Kind != ItemSkill || selected.Skill == nil {
		return "", nil
	}
	return renderSkillPreview(*selected.Skill, width)
}

func (p *skillPanel) SelectedSkill(item Item) bool {
	return item.Kind == ItemSkill && item.Skill != nil
}

func (p *skillPanel) SelectedMCP(item Item) bool { return false }
