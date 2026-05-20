package install

import (
	"testing"
)

func TestParseFindOutput(t *testing.T) {
	raw := `
Install with npx skills add <owner/repo@skill>

vercel-labs/agent-skills@vercel-react-best-practices 406.6K installs
└ https://skills.sh/vercel-labs/agent-skills/vercel-react-best-practices

vercel-labs/json-render@react 2K installs
└ https://skills.sh/vercel-labs/json-render/react
`
	results := parseFindOutput(raw)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Source != "vercel-labs/agent-skills@vercel-react-best-practices" {
		t.Fatalf("unexpected source: %q", results[0].Source)
	}
	if results[0].Name != "vercel-react-best-practices" {
		t.Fatalf("unexpected name: %q", results[0].Name)
	}
	if results[0].URL == "" {
		t.Fatal("expected URL on first result")
	}
	if results[0].Installs != "406.6K installs" {
		t.Fatalf("expected installs count, got %q", results[0].Installs)
	}
}

func TestSkillNameFromSource(t *testing.T) {
	if got := skillNameFromSource("owner/repo@my-skill"); got != "my-skill" {
		t.Fatalf("got %q", got)
	}
}
