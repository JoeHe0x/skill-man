package skill

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"

	skilldomain "skill-man/internal/domain/skill"
)

// styleOverride tunes the dark theme. It carries forward dark.json's
// document-level defaults (block_prefix, block_suffix, color) because
// glamour's MergeStyles replaces entire blocks — a partial "document"
// would zero-out inherited properties for every element.
const styleOverride = `{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "color": "252",
    "margin": 1,
    "border": true,
    "border_color": "240"
  },
  "heading": {
    "block_suffix": "\n",
    "color": "39",
    "bold": true
  },
  "h1": {
    "prefix": " ",
    "suffix": " ",
    "color": "231",
    "background_color": "99",
    "bold": true
  },
  "h2": {
    "prefix": "",
    "color": "87",
    "bold": true,
    "block_suffix": "\n"
  },
  "h3": {
    "prefix": "",
    "color": "210",
    "bold": true,
    "block_suffix": "\n"
  },
  "h4": {
    "prefix": "",
    "color": "120",
    "bold": true,
    "block_suffix": "\n"
  },
  "h5": {
    "prefix": "",
    "color": "147",
    "bold": false,
    "block_suffix": "\n"
  },
  "h6": {
    "prefix": "",
    "color": "246",
    "bold": false,
    "block_suffix": "\n"
  },
  "code_block": {
    "border": true,
    "border_color": "240",
    "margin": 0,
    "block_suffix": "\n"
  },
  "frontmatter": {
    "border": true,
    "border_color": "240",
    "margin": 0,
    "block_suffix": "\n"
  },
  "task": {
    "ticked": "☑ ",
    "unticked": "☐ "
  }
}`

var (
	rendererMu     sync.Mutex
	cachedRenderer *glamour.TermRenderer
	cachedWidth    int
)

func getRenderer(width int) (*glamour.TermRenderer, error) {
	rendererMu.Lock()
	defer rendererMu.Unlock()

	if cachedRenderer != nil && cachedWidth == width {
		return cachedRenderer, nil
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithStylesFromJSONBytes([]byte(styleOverride)),
		glamour.WithEmoji(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	cachedRenderer = r
	cachedWidth = width
	return r, nil
}

func RenderSkillPreview(skill skilldomain.Skill, width int) (string, error) {
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

	renderer, err := getRenderer(width)
	if err != nil {
		return "", err
	}

	md := formatFrontmatter(string(body))

	rendered, err := renderer.Render(md)
	if err != nil {
		return "", err
	}

	if len(skill.Tools) == 0 {
		return fmt.Sprintf("%s\n\nSource: %s\nPath: %s\nManaged: %t\nOrigin: %s\n\n%s", skill.Name, sourceLabel, skill.Path, skill.Managed, displayOrigin(skill.SourceKind, skill.SourcePath), rendered), nil
	}

	return fmt.Sprintf(
		"%s\n\nSource: %s\nPath: %s\nManaged: %t\nOrigin: %s\nTools: %s\n\n%s",
		skill.Name,
		sourceLabel,
		skill.Path,
		skill.Managed,
		displayOrigin(skill.SourceKind, skill.SourcePath),
		strings.Join(skill.Tools, ", "),
		rendered,
	), nil
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

func GetTestRenderer() (*glamour.TermRenderer, error) {
	return getRenderer(80)
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
