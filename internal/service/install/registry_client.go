package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

const defaultRegistryBase = "https://skills.sh"

// RegistryClient talks to the skills.sh HTTP API (no Node/npx).
type RegistryClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewRegistryClient() *RegistryClient {
	return &RegistryClient{
		BaseURL: defaultRegistryBase,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type registrySearchResponse struct {
	Query  string          `json:"query"`
	Skills []registrySkill `json:"skills"`
}

type registrySkill struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"`
}

// RegistrySnapshot is a downloaded skill bundle from skills.sh.
type RegistrySnapshot struct {
	Files []RegistrySnapshotFile `json:"files"`
	Hash  string                 `json:"hash"`
}

// RegistrySnapshotFile is one file in a skill snapshot.
type RegistrySnapshotFile struct {
	Path     string `json:"path"`
	Contents string `json:"contents"`
}

func (c *RegistryClient) base() string {
	if strings.TrimSpace(c.BaseURL) == "" {
		return defaultRegistryBase
	}
	return strings.TrimRight(c.BaseURL, "/")
}

func (c *RegistryClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *RegistryClient) Search(ctx context.Context, query string) ([]domaininstall.Candidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	u := c.base() + "/api/search?" + url.Values{"q": {query}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "skill-man")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("skills.sh search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("skills.sh search: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skills.sh search: HTTP %d: %s", resp.StatusCode, trimBody(body))
	}

	var parsed registrySearchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("skills.sh search: parse JSON: %w", err)
	}
	if len(parsed.Skills) == 0 {
		return nil, fmt.Errorf("no skills found for %q", query)
	}

	out := make([]domaininstall.Candidate, 0, len(parsed.Skills))
	for _, s := range parsed.Skills {
		skillName := strings.TrimSpace(s.SkillID)
		if skillName == "" {
			skillName = strings.TrimSpace(s.Name)
		}
		repoSource := strings.TrimSpace(s.Source)
		if repoSource == "" || skillName == "" {
			continue
		}
		candidate := domaininstall.Candidate{
			Source:   repoSource + "@" + skillName,
			Name:     skillName,
			Installs: formatInstallCount(s.Installs),
		}
		if id := strings.TrimSpace(s.ID); id != "" {
			candidate.URL = c.base() + "/" + id
		}
		out = append(out, candidate)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no skills found for %q", query)
	}
	return out, nil
}

func (c *RegistryClient) Download(ctx context.Context, owner, repo, slug string) (*RegistrySnapshot, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	slug = strings.TrimSpace(slug)
	if owner == "" || repo == "" || slug == "" {
		return nil, fmt.Errorf("registry download: owner, repo, and skill slug are required")
	}

	path := fmt.Sprintf("/api/download/%s/%s/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(slug),
	)
	u := c.base() + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "skill-man")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("skills.sh download: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("skills.sh download: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skills.sh download: HTTP %d: %s", resp.StatusCode, trimBody(body))
	}

	var snap RegistrySnapshot
	if err := json.Unmarshal(body, &snap); err != nil {
		return nil, fmt.Errorf("skills.sh download: parse JSON: %w", err)
	}
	if len(snap.Files) == 0 {
		return nil, fmt.Errorf("skills.sh download: empty snapshot for %s/%s/%s", owner, repo, slug)
	}
	hasSkillMD := false
	for _, f := range snap.Files {
		if strings.EqualFold(filepathBase(f.Path), "SKILL.md") {
			hasSkillMD = true
			break
		}
	}
	if !hasSkillMD {
		return nil, fmt.Errorf("skills.sh download: snapshot missing SKILL.md")
	}
	return &snap, nil
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

func trimBody(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}
