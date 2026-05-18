package service

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

	"skill-man/internal/domain"
)

type ScanStrategy[T any] interface {
	DefaultDir() string
	AgentDir(agent domain.Agent) string
	SkipDir(dirName string) bool
	TargetFiles() []string
	ParseFile(filePath, projectRoot, home string, scope domain.Scope) (T, error)
	Dedupe(entities []T) []T
}

func ScanEntities[T any](ctx context.Context, projectRoot, home string, agents []domain.Agent, strategy ScanStrategy[T]) ([]T, error) {
	seen := map[string]bool{}
	var entities []T
	var mu sync.Mutex

	if len(agents) == 0 {
		// fallback
		agents = []domain.Agent{{EntityDirs: map[domain.EntityType]string{domain.EntitySkill: strategy.DefaultDir()}}}
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, agent := range agents {
		agentDir := strategy.AgentDir(agent)
		if agentDir == "" {
			continue
		}

		projectDir := filepath.Join(projectRoot, agentDir)
		projectDir = filepath.Clean(projectDir)
		if !seen[projectDir] {
			seen[projectDir] = true
			pDir := projectDir
			g.Go(func() error {
				s, err := scanDirEntities(gCtx, pDir, projectRoot, home, domain.ScopeProject, strategy)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return nil
					}
					return err
				}
				mu.Lock()
				entities = append(entities, s...)
				mu.Unlock()
				return nil
			})
		}

		if home == "" {
			continue
		}
		globalDir := filepath.Join(home, agentDir)
		globalDir = filepath.Clean(globalDir)
		if !seen[globalDir] {
			seen[globalDir] = true
			gDir := globalDir
			g.Go(func() error {
				s, err := scanDirEntities(gCtx, gDir, projectRoot, home, domain.ScopeGlobal, strategy)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						return nil
					}
					return err
				}
				mu.Lock()
				entities = append(entities, s...)
				mu.Unlock()
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	entities = strategy.Dedupe(entities)
	return entities, nil
}

func scanDirEntities[T any](ctx context.Context, dir, projectRoot, home string, scope domain.Scope, strategy ScanStrategy[T]) ([]T, error) {
	var entities []T

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			if strategy.SkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				for _, filename := range strategy.TargetFiles() {
					if entity, err := strategy.ParseFile(filepath.Join(path, filename), projectRoot, home, scope); err == nil {
						entities = append(entities, entity)
						break
					}
				}
				return nil
			}
		}

		isTarget := false
		for _, filename := range strategy.TargetFiles() {
			if d.Name() == filename {
				isTarget = true
				break
			}
		}

		if !isTarget {
			return nil
		}

		entity, parseErr := strategy.ParseFile(path, projectRoot, home, scope)
		if parseErr != nil {
			return parseErr
		}
		entities = append(entities, entity)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entities, nil
}

type SkillScanStrategy struct{}

func (s SkillScanStrategy) DefaultDir() string {
	return ".skills"
}

func (s SkillScanStrategy) AgentDir(agent domain.Agent) string {
	if agent.EntityDirs != nil {
		return agent.EntityDirs[domain.EntitySkill]
	}
	return agent.SkillsDir
}

func (s SkillScanStrategy) SkipDir(dirName string) bool {
	return dirName == ".git" || dirName == "node_modules"
}

func (s SkillScanStrategy) TargetFiles() []string {
	return []string{"SKILL.md", "SKILL.md.disabled"}
}

func (s SkillScanStrategy) ParseFile(filePath, projectRoot, home string, scope domain.Scope) (domain.Skill, error) {
	return parseSkillFile(filePath, projectRoot, home, scope)
}

func (s SkillScanStrategy) Dedupe(skills []domain.Skill) []domain.Skill {
	return dedupeSkills(skills)
}

func ScanSkills(ctx context.Context, projectRoot, home string, agents []domain.Agent) ([]domain.Skill, error) {
	return ScanEntities(ctx, projectRoot, home, agents, SkillScanStrategy{})
}

func dedupeSkills(skills []domain.Skill) []domain.Skill {
	type key struct {
		path  string
		scope domain.Scope
	}
	seen := map[key]int{}
	var out []domain.Skill

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

func parseSkillFile(skillPath, projectRoot, home string, scope domain.Scope) (domain.Skill, error) {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return domain.Skill{}, err
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
		return domain.Skill{}, err
	}

	meta, ok, err := readInstallMetadata(dir)
	if err != nil {
		return domain.Skill{}, err
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

	agents := domain.ResolveAgentIDs(dir, projectRoot, home)

	return domain.Skill{
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

func ParseSkillFile(skillPath string) (domain.Skill, error) {
	return parseSkillFile(skillPath, "", "", domain.ScopeProject)
}
