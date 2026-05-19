package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestApplyHyperlinks_bareURL(t *testing.T) {
	in := "See https://example.com/path for details"
	out := ApplyHyperlinks(in, "")
	if !strings.Contains(out, ansi.SetHyperlink("https://example.com/path")) {
		t.Fatalf("expected OSC 8 open sequence in %q", out)
	}
	if !strings.Contains(out, ansi.ResetHyperlink()) {
		t.Fatalf("expected OSC 8 reset sequence in %q", out)
	}
}

func TestApplyHyperlinks_markdownLink(t *testing.T) {
	md := "Read [the docs](https://example.com/docs) now."
	rendered, err := Markdown(md, 80)
	if err != nil {
		t.Fatal(err)
	}
	out := ApplyHyperlinks(rendered, md)
	plain := ansi.Strip(out)
	if !strings.Contains(plain, "the docs") {
		t.Fatalf("link label missing: %q", plain)
	}
	if !strings.Contains(out, ansi.SetHyperlink("https://example.com/docs")) {
		t.Fatalf("expected hyperlink target in output: %q", out)
	}
}

func TestMarkdown_includesHyperlinks(t *testing.T) {
	md := "Visit <https://skills.sh/demo> today."
	out, err := Markdown(md, 60)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, ansi.SetHyperlink("https://skills.sh/demo")) {
		t.Fatalf("Markdown() should apply hyperlinks: %q", out)
	}
}
