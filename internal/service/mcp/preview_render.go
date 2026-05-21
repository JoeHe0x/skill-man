package mcp

import (
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	"github.com/JoeHe0x/skill-man/internal/render"
)

// RenderPreview returns a glamour-rendered markdown summary of an MCP server.
func RenderPreview(server mcpdomain.Server, width int) (string, error) {
	md, err := PreviewMarkdown(server)
	if err != nil {
		return "", err
	}
	return render.Markdown(md, width)
}

// RenderKeyPreview renders the right-pane detail for a selected MCP config key.
func RenderKeyPreview(configKey string, members []*mcpdomain.Server, home string, width int) (string, error) {
	md, err := KeyPreviewMarkdown(configKey, members, home)
	if err != nil {
		return "", err
	}
	return render.Markdown(md, width)
}
