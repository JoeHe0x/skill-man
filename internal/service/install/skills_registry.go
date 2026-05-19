package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceskill "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// SkillsCLIProvider searches and installs via the vercel-labs skills CLI (npx skills).
type SkillsCLIProvider struct {
	FindCmd string
	AddCmd  string
}

func NewSkillsCLIProvider() *SkillsCLIProvider {
	return &SkillsCLIProvider{
		FindCmd: "npx",
		AddCmd:  "npx",
	}
}

func (p *SkillsCLIProvider) Kind() domaininstall.Kind { return domaininstall.KindSkill }

func (p *SkillsCLIProvider) SupportedAgents() []agent.Agent {
	return agent.DefaultAgents()
}

func (p *SkillsCLIProvider) Search(query string) ([]domaininstall.Candidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query is required")
	}

	if local, ok := localSkillCandidate(query); ok {
		return []domaininstall.Candidate{local}, nil
	}

	out, err := p.runFind(query)
	if err != nil {
		return nil, err
	}
	if msg := extractNoSkillsMessage(out); msg != "" {
		return nil, errors.New(msg)
	}
	results := parseFindOutput(out)
	if len(results) == 0 {
		return nil, fmt.Errorf("no skills found for %q (is npx skills CLI available?)", query)
	}
	return results, nil
}

func extractNoSkillsMessage(out string) string {
	clean := ansiPattern.ReplaceAllString(out, "")
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "no skills found") {
			return line
		}
	}
	return ""
}

func (p *SkillsCLIProvider) Install(ctx context.Context, cwd, home string, candidate domaininstall.Candidate, agentIDs []string) (string, error) {
	if candidate.Local {
		agents := agentsByIDs(agentIDs, p.SupportedAgents())
		result, err := serviceskill.InstallLocalSkill(cwd, candidate.Source, agents)
		if err != nil {
			return "", err
		}
		return result.Name, nil
	}

	if len(agentIDs) == 0 {
		return "", errors.New("select at least one agent")
	}

	args := []string{"skills", "add", candidate.Source, "-y", "--agent"}
	args = append(args, agentIDs...)
	cmd := exec.CommandContext(ctx, p.AddCmd, append([]string{"--yes"}, args...)...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "CI=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("skills add: %s", msg)
	}

	name := skillNameFromSource(candidate.Source)
	return name, nil
}

func (p *SkillsCLIProvider) runFind(query string) (string, error) {
	if _, err := exec.LookPath(p.FindCmd); err != nil {
		return "", fmt.Errorf("skills find: %q not found in PATH (install Node.js/npx)", p.FindCmd)
	}

	cmd := exec.Command(p.FindCmd, "--yes", "skills", "find", query)
	cmd.Env = append(os.Environ(), "NO_COLOR=1", "CI=1", "FORCE_COLOR=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(combineOutput(stderr.String(), stdout.String()))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("skills find: %s", msg)
	}
	return combineOutput(stdout.String(), stderr.String()), nil
}

func combineOutput(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(p)
	}
	return b.String()
}

var (
	ansiPattern     = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	findLinePattern = regexp.MustCompile(`^([^\s]+@[^\s]+)\s+([\d.,]+[KMB]?\s+installs?)$`)
)

func parseFindOutput(raw string) []domaininstall.Candidate {
	clean := ansiPattern.ReplaceAllString(raw, "")
	lines := strings.Split(clean, "\n")

	var results []domaininstall.Candidate
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "Install with") {
			continue
		}
		m := findLinePattern.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		c := domaininstall.Candidate{
			Source:   m[1],
			Name:     skillNameFromSource(m[1]),
			Installs: strings.TrimSpace(m[2]),
		}
		if i+1 < len(lines) {
			next := strings.TrimSpace(lines[i+1])
			next = strings.TrimPrefix(next, "└")
			next = strings.TrimPrefix(next, "|-")
			next = strings.TrimSpace(next)
			if strings.HasPrefix(next, "http") {
				c.URL = next
				i++
			}
		}
		results = append(results, c)
	}
	return results
}

func skillNameFromSource(source string) string {
	if at := strings.LastIndex(source, "@"); at >= 0 && at < len(source)-1 {
		return source[at+1:]
	}
	return source
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
