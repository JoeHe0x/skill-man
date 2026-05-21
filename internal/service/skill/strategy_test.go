package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillFrontmatter_ignoresNameInBody(t *testing.T) {
	content := `---
name: langchain-fundamentals
description: LangChain agents guide
---

const getWeather = tool(
  async ({ location }) => "...",
  { name: "get_weather" },
);
`
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skill, err := ParseSkillFile(path)
	if err != nil {
		t.Fatalf("ParseSkillFile: %v", err)
	}
	if skill.Name != "langchain-fundamentals" {
		t.Fatalf("expected langchain-fundamentals, got %q", skill.Name)
	}
	if skill.Description != "LangChain agents guide" {
		t.Fatalf("unexpected description: %q", skill.Description)
	}
}
