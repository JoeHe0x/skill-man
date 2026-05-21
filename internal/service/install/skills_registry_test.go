package install

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestSkillsCLIProvider_InstallRegistry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/download/acme/skills/demo-skill" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(RegistrySnapshot{
			Files: []RegistrySnapshotFile{{
				Path:     "SKILL.md",
				Contents: "---\nname: demo-skill\ndescription: test skill\n---\nbody\n",
			}},
		})
	}))
	defer srv.Close()

	workspace := t.TempDir()
	p := &SkillsCLIProvider{Registry: &RegistryClient{BaseURL: srv.URL, HTTPClient: srv.Client()}}
	name, err := p.Install(context.Background(), workspace, "", domaininstall.Candidate{
		Source: "acme/skills@demo-skill",
		Name:   "demo-skill",
	}, []string{"cursor"}, extension.ScopeProject)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if name != "demo-skill" {
		t.Fatalf("unexpected name: %q", name)
	}

	a, _ := agent.AgentByID("cursor")
	target := filepath.Join(workspace, a.EntityDirs[agent.EntitySkill], "demo-skill", "SKILL.md")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected installed skill at %s: %v", target, err)
	}
}
