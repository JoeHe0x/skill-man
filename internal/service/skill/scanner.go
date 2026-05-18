package skill

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"skill-man/internal/domain/agent"
	skilldomain "skill-man/internal/domain/skill"
)

func ScanSkills(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]skilldomain.Skill, error) {
	seen := map[string]bool{}
	var skills []skilldomain.Skill
	var mu sync.Mutex

	if len(agents) == 0 {
		agents = []agent.Agent{{EntityDirs: map[agent.EntityType]string{agent.EntitySkill: ".skills"}}}
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, a := range agents {
		projectDir := filepath.Join(projectRoot, a.EntityDirs[agent.EntitySkill])
		projectDir = filepath.Clean(projectDir)
		if !seen[projectDir] {
			seen[projectDir] = true
			pDir := projectDir
			g.Go(func() error {
				s, err := scanDir(gCtx, pDir, projectRoot, home, skilldomain.ScopeProject)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return nil
					}
					return err
				}
				mu.Lock()
				skills = append(skills, s...)
				mu.Unlock()
				return nil
			})
		}

		if home == "" {
			continue
		}
		globalDir := filepath.Join(home, a.EntityDirs[agent.EntitySkill])
		globalDir = filepath.Clean(globalDir)
		if !seen[globalDir] {
			seen[globalDir] = true
			gDir := globalDir
			g.Go(func() error {
				s, err := scanDir(gCtx, gDir, projectRoot, home, skilldomain.ScopeGlobal)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return nil
					}
					return err
				}
				mu.Lock()
				skills = append(skills, s...)
				mu.Unlock()
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	skills = dedupeSkills(skills)
	return skills, nil
}

func scanDir(ctx context.Context, dir, projectRoot, home string, scope skilldomain.Scope) ([]skilldomain.Skill, error) {
	var skills []skilldomain.Skill

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				if skill, err := parseSkillFile(filepath.Join(path, "SKILL.md"), projectRoot, home, scope); err == nil {
					skills = append(skills, skill)
				} else if skill, err := parseSkillFile(filepath.Join(path, "SKILL.md.disabled"), projectRoot, home, scope); err == nil {
					skills = append(skills, skill)
				}
				return nil
			}
		}

		if d.Name() != "SKILL.md" && d.Name() != "SKILL.md.disabled" {
			return nil
		}

		skill, parseErr := parseSkillFile(path, projectRoot, home, scope)
		if parseErr != nil {
			return parseErr
		}
		skills = append(skills, skill)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return skills, nil
}

func dedupeSkills(skills []skilldomain.Skill) []skilldomain.Skill {
	type key struct {
		path  string
		scope skilldomain.Scope
	}
	seen := map[key]int{}
	var out []skilldomain.Skill

	for _, s := range skills {
		k := key{path: s.Path, scope: s.Scope}
		if idx, ok := seen[k]; ok {
			out[idx].Agents = mergeAgentIDs(out[idx].Agents, s.Agents)
			continue
		}
		seen[k] = len(out)
		out = append(out, s)
	}
	return out
}

func mergeAgentIDs(a, b []string) []string {
	set := map[string]bool{}
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		set[id] = true
	}
	var out []string
	for id := range set {
		out = append(out, id)
	}
	return out
}

func parseSkillFile(skillPath, projectRoot, home string, scope skilldomain.Scope) (skilldomain.Skill, error) {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return skilldomain.Skill{}, err
	}

	dir := filepath.Dir(skillPath)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolvedDir
	}
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		readmePath = ""
	}

	name := filepath.Base(dir)
	description := ""
	tools := []string{}

	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		text := strings.TrimSpace(string(line))
		switch {
		case strings.HasPrefix(text, "name:") && name == filepath.Base(dir):
			name = strings.TrimSpace(strings.TrimPrefix(text, "name:"))
		case strings.HasPrefix(text, "description:") && description == "":
			description = strings.TrimSpace(strings.TrimPrefix(text, "description:"))
		case strings.HasPrefix(text, "- "):
			tools = append(tools, strings.TrimSpace(strings.TrimPrefix(text, "- ")))
		case strings.HasPrefix(text, "# ") && description == "":
			description = strings.TrimSpace(strings.TrimPrefix(text, "# "))
		}
		if description != "" && len(tools) >= 4 {
			break
		}
	}

	if description == "" {
		description = "No description found in SKILL.md."
	}

	hash := sha1.Sum([]byte(skillPath))

	info, err := os.Stat(skillPath)
	if err != nil {
		return skilldomain.Skill{}, err
	}

	meta, ok, err := readInstallMetadata(dir)
	if err != nil {
		return skilldomain.Skill{}, err
	}

	managed := false
	sourceKind := ""
	sourcePath := ""
	metadataPath := ""
	if ok {
		managed = true
		sourceKind = meta.SourceKind
		sourcePath = meta.SourcePath
		metadataPath = metadataPathForDir(dir)
	}

	agents := agent.ResolveAgentIDs(dir, projectRoot, home)

	return skilldomain.Skill{
		ID:            fmt.Sprintf("%x", hash[:8]),
		Name:          strings.Trim(name, `"'`),
		Description:   strings.Trim(description, `"'`),
		Tools:         tools,
		Path:          dir,
		ReadmePath:    readmePath,
		SkillFilePath: filepath.Join(dir, filepath.Base(skillPath)),
		UpdatedAt:     info.ModTime(),
		Managed:       managed,
		SourceKind:    sourceKind,
		SourcePath:    sourcePath,
		MetadataPath:  metadataPath,
		Scope:         scope,
		Agents:        agents,
		Disabled:      strings.HasSuffix(skillPath, ".disabled"),
	}, nil
}

func ParseSkillFile(skillPath string) (skilldomain.Skill, error) {
	return parseSkillFile(skillPath, "", "", skilldomain.ScopeProject)
}
