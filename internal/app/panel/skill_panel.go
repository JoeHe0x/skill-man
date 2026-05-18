package panel

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	serviceskill "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// SkillDeps configures the skill panel.
type SkillDeps struct {
	Manager manager.ExtensionManager[*skilldomain.Skill]
}

type skillPanel struct {
	deps   SkillDeps
	skills []*skilldomain.Skill
}

// NewSkillPanel creates the skills extension panel.
func NewSkillPanel(deps SkillDeps) Panel {
	return &skillPanel{deps: deps}
}

func (p *skillPanel) Tab() Tab { return TabSkills }

func (p *skillPanel) Count() int { return len(p.skills) }

func (p *skillPanel) CountLabel() string { return "skills" }

func (p *skillPanel) Capabilities() Capabilities {
	return Capabilities{
		Inspect: true,
		Disable: true,
		Bind:    true,
		Remove:  true,
		Update:  true,
		Find:    true,
		Add:     true,
		Init:    true,
	}
}

func (p *skillPanel) ScanCmd(cwd, home string, agents []agent.Agent) tea.Cmd {
	mgr := p.deps.Manager
	return func() tea.Msg {
		skills, err := mgr.Scan(context.Background(), cwd, home, agents)
		return SkillsScannedMsg{Skills: skills, Err: err}
	}
}

func (p *skillPanel) ApplyScan(msg tea.Msg) bool {
	m, ok := msg.(SkillsScannedMsg)
	if !ok {
		return false
	}
	if m.Err != nil {
		return false
	}
	p.skills = m.Skills
	return true
}

// Skills returns the last scanned skill list.
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

func (p *skillPanel) SyncPreview(selected Item, width int, previewGen *int) tea.Cmd {
	if selected.Kind != ItemSkill || selected.Skill == nil {
		return nil
	}
	if previewGen != nil {
		*previewGen++
		gen := *previewGen
		skillCopy := *selected.Skill
		return func() tea.Msg {
			content, err := serviceskill.RenderSkillPreview(skillCopy, width)
			return PreviewLoadedMsg{Tab: TabSkills, Content: content, Err: err, Gen: gen}
		}
	}
	return nil
}

func (p *skillPanel) SelectedSkill(item Item) bool {
	return item.Kind == ItemSkill && item.Skill != nil
}

func (p *skillPanel) SelectedMCP(item Item) bool { return false }
