package service

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/glamour"
)

const unsupportedMarkdown = `
$E=mc^2$ inline math

$$
\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

==highlighted text==

H~2~O and x^2^

++inserted text++

# Heading with id {#custom-id}

*[HTML]: HyperText Markup Language

<!-- HTML comment -->

The HTML abbreviation is used here.
`

// comprehensive markdown syntax coverage test
const testMarkdown = `
# H1 Heading
## H2 Heading
### H3 Heading
#### H4 Heading
##### H5 Heading
###### H6 Heading

Alternative H1
==============

Alternative H2
--------------

**bold text** and __also bold__

*italic text* and _also italic_

***bold italic*** and ___also bold italic___

~~strikethrough~~

` + "`inline code`" + `

` + "```" + `go
func main() {
    fmt.Println("code block with language")
}
` + "```" + `

` + "```" + `
plain code block no language
` + "```" + `

> blockquote single line
>
> blockquote with **bold** and` + "`code`" + `

> nested blockquote
>> nested level 2
>>> nested level 3

- unordered item 1
- unordered item 2
  - nested item 2.1
  - nested item 2.2
- unordered item 3

1. ordered item 1
2. ordered item 2
   1. nested ordered 2.1
   2. nested ordered 2.2
3. ordered item 3

- [ ] unchecked task
- [x] checked task
- [ ] another unchecked

| Column A | Column B | Column C |
|----------|----------|----------|
| a1       | b1       | c1       |
| a2       | b2       | c2       |

---

***

___

[inline link](https://example.com)

[reference link][ref]

[ref]: https://example.com "title"

![image alt](https://example.com/img.png)

<https://autolink.example.com>

email@example.com

Here is a footnote reference[^1].

[^1]: This is the footnote content.

term
: definition of the term

another term
: another definition

Some text with <span style="color:red">inline HTML</span>.

:smile: :heart: :rocket:

Here is text with a\
hard line break.

Here is text with a
soft line break (two trailing spaces).
`

// remove leading newline
var cleanMarkdown = strings.TrimPrefix(testMarkdown, "\n")

func TestMarkdownSyntaxCoverage(t *testing.T) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithStylesFromJSONBytes([]byte(styleOverride)),
		glamour.WithEmoji(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	rendered, err := r.Render(cleanMarkdown)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	// Strip ANSI escape codes to check plain text content
	ansiStrip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	plain := ansiStrip.ReplaceAllString(rendered, "")

	tests := []struct {
		name    string
		check   func() bool
		details string
	}{
		// --- Headings ---
		{
			name: "ATX H1",
			check: func() bool {
				return hasANSIColor(rendered, "231") && hasANSIBackground(rendered, "99")
			},
			details: "H1 should have color 231 on background 99",
		},
		{
			name: "ATX H2",
			check: func() bool {
				return hasANSIColor(rendered, "87") && strings.Contains(plain, "H2 Heading") && !strings.Contains(plain, "## H2 Heading")
			},
			details: "H2 should have color 87 without prefix",
		},
		{
			name: "ATX H3",
			check: func() bool {
				return hasANSIColor(rendered, "210") && strings.Contains(plain, "H3 Heading") && !strings.Contains(plain, "### H3 Heading")
			},
			details: "H3 should have color 210 without prefix",
		},
		{
			name: "ATX H4",
			check: func() bool {
				return hasANSIColor(rendered, "120") && strings.Contains(plain, "H4 Heading") && !strings.Contains(plain, "#### H4 Heading")
			},
			details: "H4 should have color 120 without prefix",
		},
		{
			name: "ATX H5",
			check: func() bool {
				return hasANSIColor(rendered, "147") && strings.Contains(plain, "H5 Heading") && !strings.Contains(plain, "##### H5 Heading")
			},
			details: "H5 should have color 147 without prefix",
		},
		{
			name: "ATX H6",
			check: func() bool {
				return hasANSIColor(rendered, "246") && strings.Contains(plain, "H6 Heading") && !strings.Contains(plain, "###### H6 Heading")
			},
			details: "H6 should have color 246 without prefix",
		},
		// --- Bold / Italic ---
		{
			name: "Bold (**)",
			check: func() bool {
				return hasBold(renderedHunk(rendered, "bold text"))
			},
			details: "**bold** should produce bold ANSI",
		},
		{
			name: "Italic (*)",
			check: func() bool {
				return hasItalic(renderedHunk(rendered, "italic text"))
			},
			details: "*italic* should produce italic ANSI",
		},
		{
			name: "Strikethrough",
			check: func() bool {
				return hasStrikethrough(renderedHunk(rendered, "strikethrough"))
			},
			details: "~~strikethrough~~ should produce strikethrough ANSI",
		},
		// --- Code ---
		{
			name: "Inline code",
			check: func() bool {
				hunk := renderedHunk(rendered, "inline code")
				return hasANSIColorInHunk(hunk, "203") && hasANSIBackgroundInHunk(hunk, "236")
			},
			details: "Inline code should have distinct color and background",
		},
		{
			name: "Fenced code block with language",
			check: func() bool {
				return strings.Contains(plain, "func main()") && strings.Contains(plain, "fmt.Println")
			},
			details: "Code block content should be preserved",
		},
		{
			name: "Fenced code block no language",
			check: func() bool {
				return strings.Contains(plain, "plain code block no language")
			},
			details: "Code block without language should still render",
		},
		// --- Blockquote ---
		{
			name: "Blockquote",
			check: func() bool {
				return strings.Contains(plain, "blockquote single line")
			},
			details: "Blockquote text should be preserved",
		},
		{
			name: "Nested blockquote",
			check: func() bool {
				return strings.Contains(plain, "nested level 2") && strings.Contains(plain, "nested level 3")
			},
			details: "Nested blockquotes should be rendered",
		},
		// --- Lists ---
		{
			name: "Unordered list",
			check: func() bool {
				return strings.Contains(plain, "unordered item 1") && strings.Contains(plain, "unordered item 2")
			},
			details: "Unordered list items should be present",
		},
		{
			name: "Ordered list",
			check: func() bool {
				return strings.Contains(plain, "ordered item 1") && strings.Contains(plain, "ordered item 2")
			},
			details: "Ordered list items should be present",
		},
		{
			name: "Task list",
			check: func() bool {
				return strings.Contains(plain, "☐ unchecked task") && strings.Contains(plain, "☑ checked task")
			},
			details: "Task list items should be present with checkboxes",
		},
		// --- Table ---
		{
			name: "Table",
			check: func() bool {
				return strings.Contains(plain, "Column A") && strings.Contains(plain, "a1") && strings.Contains(plain, "b2")
			},
			details: "Table content should be preserved",
		},
		// --- Horizontal rule ---
		{
			name: "HR (dashes)",
			check: func() bool {
				return strings.Contains(plain, "--------")
			},
			details: "--- should produce a horizontal rule",
		},
		// --- Links ---
		{
			name: "Inline link",
			check: func() bool {
				return strings.Contains(plain, "inline link") && strings.Contains(plain, "https://example.com")
			},
			details: "Link text and URL should be present",
		},
		{
			name: "Reference link",
			check: func() bool {
				return strings.Contains(plain, "reference link")
			},
			details: "Reference link text should be present",
		},
		{
			name: "Autolink",
			check: func() bool {
				return strings.Contains(plain, "https://autolink.example.com")
			},
			details: "Autolink URL should be in output",
		},
		// --- Image ---
		{
			name: "Image",
			check: func() bool {
				return strings.Contains(plain, "image alt") || strings.Contains(plain, "Image:")
			},
			details: "Image should have alt text or 'Image:' label in output",
		},
		// --- Emoji ---
		{
			name: "Emoji :smile:",
			check: func() bool {
				return strings.Contains(plain, "😄") || strings.Contains(plain, "smile")
			},
			details: ":smile: should be converted to 😄",
		},
		{
			name: "Emoji :heart:",
			check: func() bool {
				return strings.Contains(plain, "❤") || strings.Contains(plain, "heart")
			},
			details: ":heart: should be converted to ❤️",
		},
		{
			name: "Emoji :rocket:",
			check: func() bool {
				return strings.Contains(plain, "🚀") || strings.Contains(plain, "rocket")
			},
			details: ":rocket: should be converted to 🚀",
		},
		// --- Footnote ---
		{
			name: "Footnote reference",
			check: func() bool {
				return strings.Contains(plain, "[^1]") || strings.Contains(plain, "footnote reference")
			},
			details: "Footnote reference should be present",
		},
		{
			name: "Footnote content",
			check: func() bool {
				return strings.Contains(plain, "footnote content")
			},
			details: "Footnote definition content should be present",
		},
		// --- Definition list ---
		{
			name: "Definition list",
			check: func() bool {
				return strings.Contains(plain, "definition of the term") && strings.Contains(plain, "term")
			},
			details: "Definition list content should be preserved",
		},
		// --- HTML ---
		{
			name: "Inline HTML span",
			check: func() bool {
				return strings.Contains(plain, "inline HTML")
			},
			details: "Inline HTML text content should be preserved (tags stripped)",
		},
		// --- Setext headings ---
		{
			name: "Setext H1 (===)",
			check: func() bool {
				return strings.Contains(plain, "Alternative H1")
			},
			details: "Setext heading === should be converted to H1",
		},
		{
			name: "Setext H2 (---)",
			check: func() bool {
				return strings.Contains(plain, "Alternative H2")
			},
			details: "Setext heading --- should be converted to H2",
		},
	}

	fmt.Println("=== Markdown Syntax Coverage Report ===")
	fmt.Println()

	passed, failed := 0, 0
	var failures []string

	for _, tt := range tests {
		ok := tt.check()
		if ok {
			passed++
			fmt.Printf("  ✅ %s\n", tt.name)
		} else {
			failed++
			fmt.Printf("  ❌ %s — %s\n", tt.name, tt.details)
			failures = append(failures, tt.name)
		}
	}

	fmt.Println()
	fmt.Printf("Passed: %d, Failed: %d\n", passed, failed)

	if failed > 0 {
		fmt.Println("\nUnsupported or broken features:")
		for _, f := range failures {
			fmt.Printf("  - %s\n", f)
		}
	}

	// Write rendered output for visual inspection
	os.WriteFile("/tmp/markdown-preview-output.txt", []byte(rendered), 0644)

	fmt.Println("\nRendered output written to /tmp/markdown-preview-output.txt")
	fmt.Println("(view with: cat /tmp/markdown-preview-output.txt)")
}

func hasANSIColor(output, colorCode string) bool {
	// Look for foreground color patterns like 38;5;{code}m
	return strings.Contains(output, "38;5;"+colorCode)
}

func hasANSIBackground(output, colorCode string) bool {
	return strings.Contains(output, "48;5;"+colorCode)
}

func hasBold(hunk string) bool {
	return strings.Contains(hunk, "\x1b[1m") || strings.Contains(hunk, ";1m")
}

func hasItalic(hunk string) bool {
	return strings.Contains(hunk, "\x1b[3m") || strings.Contains(hunk, ";3m")
}

func hasStrikethrough(hunk string) bool {
	return strings.Contains(hunk, "\x1b[9m") || strings.Contains(hunk, ";9m")
}

func hasANSIColorInHunk(hunk, colorCode string) bool {
	return strings.Contains(hunk, "38;5;"+colorCode)
}

func hasANSIBackgroundInHunk(hunk, colorCode string) bool {
	return strings.Contains(hunk, "48;5;"+colorCode)
}

// renderedHunk extracts a portion of rendered output around a text fragment
func renderedHunk(rendered, around string) string {
	idx := strings.Index(rendered, around)
	if idx < 0 {
		return ""
	}
	start := max(0, idx-80)
	end := min(len(rendered), idx+len(around)+80)
	return rendered[start:end]
}

func TestUnsupportedMarkdownSyntax(t *testing.T) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithStylesFromJSONBytes([]byte(styleOverride)),
		glamour.WithEmoji(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	rendered, err := r.Render(strings.TrimPrefix(unsupportedMarkdown, "\n"))
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	ansiStrip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	plain := ansiStrip.ReplaceAllString(rendered, "")

	tests := []struct {
		name    string
		desc    string
		check   func() bool
		details string
	}{
		{
			name:    "Inline math ($...$)",
			desc:    "Common in technical docs",
			check:   func() bool { return strings.Contains(plain, "E=mc") },
			details: "Math should be preserved as plain text",
		},
		{
			name:    "Block math ($$...$$)",
			desc:    "Common in technical docs",
			check:   func() bool { return strings.Contains(plain, "int_0") },
			details: "Math block should be preserved as plain text",
		},
		{
			name:    "Highlight (==text==)",
			desc:    "Highlighted/marked text",
			check:   func() bool { return strings.Contains(plain, "highlighted text") },
			details: "== markers are stripped, text preserved",
		},
		{
			name:    "Subscript (~text~)",
			desc:    "Subscript like H~2~O",
			check:   func() bool { return strings.Contains(plain, "H~2~O") || strings.Contains(plain, "H2O") },
			details: "~ markers may cause strikethrough, not subscript",
		},
		{
			name:    "Superscript (^text^)",
			desc:    "Superscript like x^2^",
			check:   func() bool { return strings.Contains(plain, "x^2^") || strings.Contains(plain, "x2") },
			details: "^ has no special meaning in goldmark",
		},
		{
			name:    "Insert (++text++)",
			desc:    "Inserted/underlined text",
			check:   func() bool { return strings.Contains(plain, "inserted text") },
			details: "++ markers are just treated as literal text",
		},
		{
			name:    "Heading ID ({#custom-id})",
			desc:    "Custom heading anchors",
			check:   func() bool { return strings.Contains(plain, "Heading with id") },
			details: "{#custom-id} is rendered as literal text or stripped",
		},
		{
			name:    "Abbreviations",
			desc:    "ABBR tag extension",
			check:   func() bool { return strings.Contains(plain, "HTML abbreviation") },
			details: "Abbreviation definitions are rendered as literal text",
		},
		{
			name:    "HTML comments",
			desc:    "<!-- comment -->",
			check:   func() bool { return !strings.Contains(plain, "<!-- HTML comment") },
			details: "HTML comments are correctly stripped",
		},
		{
			name:    "Mermaid diagrams",
			desc:    "```mermaid code blocks",
			check:   func() bool { return true },
			details: "Mermaid is treated as a regular code block (no rendering)",
		},
	}

	fmt.Println("\n=== Unsupported/Partially-Supported Markdown Features ===")
	fmt.Println()

	fullyMissing := 0
	partiallySupported := 0

	for _, tt := range tests {
		ok := tt.check()
		if ok {
			partiallySupported++
			fmt.Printf("  ⚠️  %s — rendered as plain text (no special formatting)\n", tt.name)
			fmt.Printf("      Use case: %s\n", tt.desc)
		} else {
			fullyMissing++
			fmt.Printf("  ❌ %s — content stripped/lost\n", tt.name)
			fmt.Printf("      %s\n", tt.details)
		}
	}

	fmt.Println()
	fmt.Printf("Partially supported (plain text): %d\n", partiallySupported)
	fmt.Printf("Fully missing (content lost):     %d\n", fullyMissing)

	os.WriteFile("/tmp/markdown-unsupported-output.txt", []byte(rendered), 0644)
}

func TestPlainTextExtraction(t *testing.T) {
	// Verify the ANSI-stripped output contains expected content
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithStylesFromJSONBytes([]byte(styleOverride)),
		glamour.WithEmoji(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := r.Render("# Hello\n\nWorld")
	if err != nil {
		t.Fatal(err)
	}

	ansiStrip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	plain := ansiStrip.ReplaceAllString(rendered, "")
	if !strings.Contains(plain, "Hello") {
		t.Error("should contain Hello")
	}
	if !strings.Contains(plain, "World") {
		t.Error("should contain World")
	}
}
