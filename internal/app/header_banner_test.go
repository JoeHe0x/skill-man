package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestJoinHeaderColumnsVerticalSeparator(t *testing.T) {
	t.Parallel()

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	out := joinHeaderColumns("AAA\nBBB", "one\ntwo\nthree", sep)
	if !strings.Contains(out, "│") {
		t.Fatalf("expected vertical separator, got:\n%s", out)
	}
	if strings.Count(out, "\n") < 2 {
		t.Fatal("expected multi-line join")
	}
}

func TestHeaderOverviewText(t *testing.T) {
	t.Parallel()

	m := mustModel(t, New("/tmp/workspace", "/home/test"))
	m.status = "ready"
	out := m.headerOverviewText()
	if !strings.Contains(out, "Skills") || !strings.Contains(out, "MCP") {
		t.Fatalf("unexpected overview: %q", out)
	}
}
