package render

import (
	"bytes"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var bareURLPattern = regexp.MustCompile(`https?://[^\s\]\)>]+|mailto:[^\s\]\)>]+`)

type mdLink struct {
	text string
	href string
}

// ApplyHyperlinks wraps visible link text and URLs with OSC 8 sequences so
// supporting terminals open them on click.
func ApplyHyperlinks(rendered, mdSource string) string {
	s := rendered
	if mdSource != "" {
		plain := ansi.Strip(s)
		type positioned struct {
			idx  int
			link mdLink
		}
		var ordered []positioned
		for _, link := range extractMarkdownLinks(mdSource) {
			if idx := strings.Index(plain, link.text); idx >= 0 {
				ordered = append(ordered, positioned{idx, link})
			}
		}
		slices.SortFunc(ordered, func(a, b positioned) int { return b.idx - a.idx })
		for _, p := range ordered {
			s = wrapLinkText(s, p.link.text, p.link.href)
		}
	}
	return linkifyBareURLs(s)
}

func extractMarkdownLinks(md string) []mdLink {
	source := []byte(md)
	doc := goldmark.New().Parser().Parse(text.NewReader(source))
	var links []mdLink
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.Link:
			label := linkLabel(n, source)
			href := string(n.Destination)
			if label != "" && href != "" {
				links = append(links, mdLink{text: label, href: href})
			}
		case *ast.AutoLink:
			u := string(n.URL(source))
			if n.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(u), "mailto:") {
				u = "mailto:" + u
			}
			if u != "" {
				links = append(links, mdLink{text: u, href: u})
			}
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
		}
	}
	walk(doc)
	return links
}

func linkLabel(n *ast.Link, source []byte) string {
	var b strings.Builder
	var collect func(ast.Node)
	collect = func(node ast.Node) {
		if t, ok := node.(*ast.Text); ok {
			b.Write(t.Segment.Value(source))
		}
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			collect(c)
		}
	}
	collect(n)
	return strings.TrimSpace(b.String())
}

func wrapLinkText(s, label, href string) string {
	if label == "" || href == "" {
		return s
	}
	if _, err := url.Parse(href); err != nil {
		return s
	}
	plain := ansi.Strip(s)
	idx := strings.Index(plain, label)
	if idx < 0 {
		return s
	}
	start := utf8.RuneCountInString(plain[:idx])
	end := start + utf8.RuneCountInString(label)
	return injectVisibleRange(s, start, end, ansi.SetHyperlink(href), ansi.ResetHyperlink())
}

func linkifyBareURLs(s string) string {
	plain := ansi.Strip(s)
	matches := bareURLPattern.FindAllStringIndex(plain, -1)
	if len(matches) == 0 {
		return s
	}
	out := s
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		u := plain[m[0]:m[1]]
		start := utf8.RuneCountInString(plain[:m[0]])
		end := utf8.RuneCountInString(plain[:m[1]])
		out = injectVisibleRange(out, start, end, ansi.SetHyperlink(u), ansi.ResetHyperlink())
	}
	return out
}

// injectVisibleRange inserts prefix/suffix around visible runes [start, end).
func injectVisibleRange(s string, start, end int, prefix, suffix string) string {
	if start >= end {
		return s
	}
	var out bytes.Buffer
	vis := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' || s[i] == '\x9b' {
			j := i + 1
			if j < len(s) && s[i] == '\x1b' && s[j] == ']' {
				j++
				for j < len(s) && s[j] != '\x07' && s[j] != '\x9c' {
					if s[j] == '\x1b' {
						break
					}
					j++
				}
				if j < len(s) && (s[j] == '\x07' || s[j] == '\x9c') {
					j++
				}
			} else {
				for j < len(s) && !isFinalST(s[j]) {
					j++
				}
			}
			out.WriteString(s[i:j])
			i = j
			continue
		}
		_, width := utf8.DecodeRuneInString(s[i:])
		if vis == start {
			out.WriteString(prefix)
		}
		out.WriteString(s[i : i+width])
		vis++
		if vis == end {
			out.WriteString(suffix)
		}
		i += width
	}
	if vis > start && vis < end {
		out.WriteString(suffix)
	}
	return out.String()
}

func isFinalST(b byte) bool {
	return (b >= 0x40 && b <= 0x7e) || b == '\x9c'
}
