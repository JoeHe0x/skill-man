//go:build integration

package install

import (
	"os"
	"os/exec"
	"testing"
)

func TestSearchLiveRegistry(t *testing.T) {
	if os.Getenv("SKILL_MAN_INTEGRATION") == "" {
		t.Skip("set SKILL_MAN_INTEGRATION=1 to run live npx skills find test")
	}
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not in PATH")
	}
	p := NewSkillsCLIProvider()
	results, err := p.Search("react")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for react")
	}
}
