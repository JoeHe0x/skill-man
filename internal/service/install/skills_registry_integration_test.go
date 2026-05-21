//go:build integration

package install

import (
	"context"
	"os"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

func TestSearchLiveRegistry(t *testing.T) {
	if os.Getenv("SKILL_MAN_INTEGRATION") == "" {
		t.Skip("set SKILL_MAN_INTEGRATION=1 to run live skills.sh API test")
	}
	p := NewSkillsCLIProvider()
	results, err := p.Search("react")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for react")
	}
	if results[0].Source == "" {
		t.Fatal("expected registry source id on first result")
	}
}

func TestInstallLiveRegistryDownload(t *testing.T) {
	if os.Getenv("SKILL_MAN_INTEGRATION") == "" {
		t.Skip("set SKILL_MAN_INTEGRATION=1 to run live skills.sh download test")
	}
	workspace := t.TempDir()
	p := NewSkillsCLIProvider()
	results, err := p.Search("find-skills")
	if err != nil || len(results) == 0 {
		t.Fatalf("Search find-skills: %v (len=%d)", err, len(results))
	}
	name, err := p.Install(context.Background(), workspace, "", results[0], []string{"cursor"}, extension.ScopeProject)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if name == "" {
		t.Fatal("expected skill name")
	}
}
