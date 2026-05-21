package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceskill "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// SkillsCLIProvider searches and installs skills via skills.sh HTTP APIs (no Node/npx).
type SkillsCLIProvider struct {
	Registry *RegistryClient
}

func NewSkillsCLIProvider() *SkillsCLIProvider {
	return &SkillsCLIProvider{Registry: NewRegistryClient()}
}

func (p *SkillsCLIProvider) Kind() domaininstall.Kind { return domaininstall.KindSkill }

func (p *SkillsCLIProvider) SupportedAgents() []agent.Agent {
	return agent.DefaultAgents()
}

func (p *SkillsCLIProvider) registry() *RegistryClient {
	if p != nil && p.Registry != nil {
		return p.Registry
	}
	return NewRegistryClient()
}

func (p *SkillsCLIProvider) Search(query string) ([]domaininstall.Candidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query is required")
	}

	if local, ok := localSkillCandidate(query); ok {
		return []domaininstall.Candidate{local}, nil
	}

	return p.registry().Search(context.Background(), query)
}

func (p *SkillsCLIProvider) Install(ctx context.Context, cwd, home string, candidate domaininstall.Candidate, agentIDs []string, scope extension.Scope) (string, error) {
	if candidate.Local {
		agents := agentsByIDs(agentIDs, p.SupportedAgents())
		result, err := serviceskill.InstallLocalSkill(cwd, home, scope, candidate.Source, agents)
		if err != nil {
			return "", err
		}
		return result.Name, nil
	}

	if len(agentIDs) == 0 {
		return "", errors.New("select at least one agent")
	}

	owner, repo, skillRef, err := parseRegistrySource(candidate.Source)
	if err != nil {
		return "", err
	}

	slug := skillSlug(skillRef)
	snap, err := p.registry().Download(ctx, owner, repo, slug)
	if err != nil {
		return "", err
	}

	files := make([]serviceskill.RegistrySnapshotFile, len(snap.Files))
	for i, f := range snap.Files {
		files[i] = serviceskill.RegistrySnapshotFile{Path: f.Path, Contents: f.Contents}
	}

	agents := agentsByIDs(agentIDs, p.SupportedAgents())
	result, err := serviceskill.InstallRegistrySkill(cwd, home, scope, candidate.Source, files, agents)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func localSkillCandidate(source string) (domaininstall.Candidate, bool) {
	abs, err := filepath.Abs(source)
	if err != nil {
		return domaininstall.Candidate{}, false
	}
	info, err := os.Stat(abs)
	if err != nil {
		return domaininstall.Candidate{}, false
	}
	dir := abs
	if !info.IsDir() {
		if filepath.Base(abs) != "SKILL.md" {
			return domaininstall.Candidate{}, false
		}
		dir = filepath.Dir(abs)
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		return domaininstall.Candidate{}, false
	}
	name := filepath.Base(dir)
	return domaininstall.Candidate{
		Source: dir,
		Name:   name,
		Local:  true,
	}, true
}

func agentsByIDs(ids []string, all []agent.Agent) []agent.Agent {
	if len(ids) == 0 {
		return nil
	}
	var out []agent.Agent
	for _, id := range ids {
		if a, ok := agent.AgentByID(id); ok {
			out = append(out, a)
		}
	}
	if len(out) > 0 {
		return out
	}
	return all
}

func skillNameFromSource(source string) string {
	if at := strings.LastIndex(source, "@"); at >= 0 && at < len(source)-1 {
		return source[at+1:]
	}
	return source
}
