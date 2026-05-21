package install

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegistryClient_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(registrySearchResponse{
			Skills: []registrySkill{{
				ID:       "vercel-labs/agent-skills/vercel-react-best-practices",
				SkillID:  "vercel-react-best-practices",
				Name:     "vercel-react-best-practices",
				Installs: 406600,
				Source:   "vercel-labs/agent-skills",
			}},
		})
	}))
	defer srv.Close()

	client := &RegistryClient{BaseURL: srv.URL, HTTPClient: srv.Client()}
	results, err := client.Search(context.Background(), "react")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Source != "vercel-labs/agent-skills@vercel-react-best-practices" {
		t.Fatalf("unexpected source: %q", results[0].Source)
	}
	if results[0].URL != srv.URL+"/vercel-labs/agent-skills/vercel-react-best-practices" {
		t.Fatalf("unexpected url: %q", results[0].URL)
	}
	if results[0].Installs != "406.6K installs" {
		t.Fatalf("unexpected installs: %q", results[0].Installs)
	}
}

func TestRegistryClient_Download(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/download/vercel-labs/agent-skills/vercel-react-best-practices" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(RegistrySnapshot{
			Files: []RegistrySnapshotFile{{Path: "SKILL.md", Contents: "---\nname: demo\ndescription: x\n---\n"}},
			Hash:  "abc",
		})
	}))
	defer srv.Close()

	client := &RegistryClient{BaseURL: srv.URL, HTTPClient: srv.Client()}
	snap, err := client.Download(context.Background(), "vercel-labs", "agent-skills", "vercel-react-best-practices")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(snap.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(snap.Files))
	}
}

func TestParseRegistrySource(t *testing.T) {
	owner, repo, skill, err := parseRegistrySource("vercel-labs/agent-skills@vercel-react-best-practices")
	if err != nil {
		t.Fatal(err)
	}
	if owner != "vercel-labs" || repo != "agent-skills" || skill != "vercel-react-best-practices" {
		t.Fatalf("got %s/%s@%s", owner, repo, skill)
	}
}

func TestSkillSlug(t *testing.T) {
	if got := skillSlug("react:components"); got != "reactcomponents" {
		t.Fatalf("got %q", got)
	}
	if got := skillSlug("My_Skill Name"); got != "my-skill-name" {
		t.Fatalf("got %q", got)
	}
}

func TestSkillNameFromSource(t *testing.T) {
	if got := skillNameFromSource("owner/repo@my-skill"); got != "my-skill" {
		t.Fatalf("got %q", got)
	}
}
