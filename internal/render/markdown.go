package render

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/glamour"
)

// styleOverrideDark tunes the dark theme. It carries forward dark.json's
// document-level defaults (block_prefix, block_suffix, color) because
// glamour's MergeStyles replaces entire blocks — a partial "document"
// would zero-out inherited properties for every element.
const styleOverrideDark = `{
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

const styleOverrideLight = `{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "color": "235",
    "margin": 1,
    "border": true,
    "border_color": "250"
  },
  "heading": {
    "block_suffix": "\n",
    "color": "33",
    "bold": true
  },
  "h1": {
    "prefix": " ",
    "suffix": " ",
    "color": "255",
    "background_color": "62",
    "bold": true
  },
  "h2": {
    "prefix": "",
    "color": "62",
    "bold": true,
    "block_suffix": "\n"
  },
  "h3": {
    "prefix": "",
    "color": "61",
    "bold": true,
    "block_suffix": "\n"
  },
  "h4": {
    "prefix": "",
    "color": "35",
    "bold": true,
    "block_suffix": "\n"
  },
  "h5": {
    "prefix": "",
    "color": "37",
    "bold": false,
    "block_suffix": "\n"
  },
  "h6": {
    "prefix": "",
    "color": "241",
    "bold": false,
    "block_suffix": "\n"
  },
  "code_block": {
    "border": true,
    "border_color": "250",
    "margin": 0,
    "block_suffix": "\n"
  },
  "frontmatter": {
    "border": true,
    "border_color": "250",
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
	cachedDark     bool
	darkTheme      = true
)

// SetDarkTheme switches glamour markdown rendering between dark and light palettes.
func SetDarkTheme(dark bool) {
	rendererMu.Lock()
	defer rendererMu.Unlock()
	if darkTheme == dark {
		return
	}
	darkTheme = dark
	cachedRenderer = nil
	cachedWidth = 0
}

func renderer(width int) (*glamour.TermRenderer, error) {
	rendererMu.Lock()
	defer rendererMu.Unlock()
	return rendererLocked(width)
}

func rendererLocked(width int) (*glamour.TermRenderer, error) {
	if cachedRenderer != nil && cachedWidth == width && cachedDark == darkTheme {
		return cachedRenderer, nil
	}

	styleName := "dark"
	override := styleOverrideDark
	if !darkTheme {
		styleName = "light"
		override = styleOverrideLight
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styleName),
		glamour.WithStylesFromJSONBytes([]byte(override)),
		glamour.WithEmoji(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	cachedRenderer = r
	cachedWidth = width
	cachedDark = darkTheme
	return r, nil
}

// Markdown renders markdown for the TUI preview viewport.
// Glamour's renderer is not safe for concurrent use; the mutex covers the full render.
func Markdown(md string, width int) (string, error) {
	rendererMu.Lock()
	defer rendererMu.Unlock()

	var rendered string
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("markdown render panic: %v", r)
			}
		}()
		r, rerr := rendererLocked(width)
		if rerr != nil {
			err = rerr
			return
		}
		rendered, err = r.Render(md)
	}()
	if err != nil {
		return "", err
	}
	return ApplyHyperlinks(rendered, md), nil
}

// TestRenderer returns a glamour renderer for tests (width 80).
func TestRenderer() (*glamour.TermRenderer, error) {
	return renderer(80)
}
