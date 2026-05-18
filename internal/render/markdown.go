package render

import (
	"sync"

	"github.com/charmbracelet/glamour"
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

func renderer(width int) (*glamour.TermRenderer, error) {
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

// Markdown renders markdown for the TUI preview viewport.
func Markdown(md string, width int) (string, error) {
	r, err := renderer(width)
	if err != nil {
		return "", err
	}
	return r.Render(md)
}

// TestRenderer returns a glamour renderer for tests (width 80).
func TestRenderer() (*glamour.TermRenderer, error) {
	return renderer(80)
}
