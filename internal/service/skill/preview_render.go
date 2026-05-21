package skill

import (
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/render"
)

func renderMarkdown(md string, width int) (string, error) {
	return render.Markdown(md, width)
}

// RenderSkillPreview renders a skill preview for the TUI (markdown + glamour).
func RenderSkillPreview(skill skilldomain.Skill, width int) (string, error) {
	md, err := PreviewMarkdown(skill)
	if err != nil {
		return "", err
	}
	return renderMarkdown(md, width)
}
