package panel

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/render"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	serviceskill "github.com/JoeHe0x/skill-man/internal/service/skill"
)

func renderSkillPreview(skill skilldomain.Skill, width int) (string, error) {
	md, err := serviceskill.PreviewMarkdown(skill)
	if err != nil {
		return "", err
	}
	return render.Markdown(md, width)
}

func renderMCPKeyPreview(configKey string, members []*mcpdomain.Server, home string, width int) (string, error) {
	md, err := servicemcp.KeyPreviewMarkdown(configKey, members, home)
	if err != nil {
		return "", err
	}
	return render.Markdown(md, width)
}
