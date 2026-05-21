package skill

import (
	"fmt"
	"os"
	"strings"

	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// PreviewMarkdown builds the markdown document for a skill preview (no terminal rendering).
func PreviewMarkdown(skill skilldomain.Skill) (string, error) {
	contentPath := skill.ReadmePath
	sourceLabel := "README.md"
	if contentPath == "" {
		contentPath = skill.ConfigPath
		sourceLabel = "SKILL.md"
	}

	body, err := os.ReadFile(contentPath)
	if err != nil {
		return "", err
	}

	md := formatFrontmatter(string(body))
	header := fmt.Sprintf(
		"%s\n\nSource: %s\nPath: %s\nManaged: %t\nOrigin: %s",
		skill.Name,
		sourceLabel,
		skill.Path,
		skill.Managed,
		displayOrigin(skill.SourceKind, skill.SourcePath),
	)
	if len(skill.Tools) > 0 {
		header = fmt.Sprintf("%s\nTools: %s", header, strings.Join(skill.Tools, ", "))
	}
	return header + "\n\n" + md, nil
}

func RenderCommandPreview(name, usage, summary string, implemented bool) string {
	status := "Implemented"
	if !implemented {
		status = "Planned"
	}

	return fmt.Sprintf(
		"# /%s\n\nUsage: `%s`\n\nStatus: %s\n\n%s",
		name,
		usage,
		status,
		summary,
	)
}

// formatFrontmatter detects YAML frontmatter (--- delimited at file start)
// and wraps it in a ```yaml fenced code block so glamour/chroma syntax-highlights
// keys and values in distinct colors.
func formatFrontmatter(md string) string {
	if !strings.HasPrefix(md, "---\n") {
		return md
	}

	idx := strings.Index(md[4:], "\n---\n")
	if idx < 0 {
		return md
	}
	fmEnd := 4 + idx + 5 // past "\n---\n"
	fm := md[4 : 4+idx]

	var out strings.Builder
	out.WriteString("```yaml\n")
	out.WriteString(fm)
	if !strings.HasSuffix(fm, "\n") {
		out.WriteString("\n")
	}
	out.WriteString("```\n")
	out.WriteString(md[fmEnd:])
	return out.String()
}

func FormatFrontmatterForTest(md string) string {
	return formatFrontmatter(md)
}

func displayOrigin(kind, path string) string {
	if kind == "" && path == "" {
		return "n/a"
	}
	if kind == "" {
		return path
	}
	if path == "" {
		return kind
	}
	return fmt.Sprintf("%s:%s", kind, path)
}
