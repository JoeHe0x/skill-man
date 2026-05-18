package manager

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"

	"skill-man/internal/domain/agent"
	"skill-man/internal/domain/extension"
)

// ScanExtensions provides a generic implementation for scanning any extension type
// (skills, hooks, sub-agents, mcp servers).
func ScanExtensions[T extension.Extension](
	ctx context.Context,
	projectRoot, home string,
	agents []agent.Agent,
	strategy ScanStrategy[T],
) ([]T, error) {
	seen := map[string]bool{}
	var entities []T
	var mu sync.Mutex

	if len(agents) == 0 {
		agents = []agent.Agent{{EntityDirs: map[agent.EntityType]string{agent.EntitySkill: strategy.DefaultDir()}}}
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, a := range agents {
		agentDir := strategy.AgentDir(a)
		if agentDir == "" {
			continue
		}

		// Scan project directory
		projectDir := filepath.Join(projectRoot, agentDir)
		projectDir = filepath.Clean(projectDir)
		if !seen[projectDir] {
			seen[projectDir] = true
			pDir := projectDir
			g.Go(func() error {
				s, err := scanDir(gCtx, pDir, projectRoot, home, extension.ScopeProject, strategy)
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

		// Scan global directory
		if home == "" {
			continue
		}
		globalDir := filepath.Join(home, agentDir)
		globalDir = filepath.Clean(globalDir)
		if !seen[globalDir] {
			seen[globalDir] = true
			gDir := globalDir
			g.Go(func() error {
				s, err := scanDir(gCtx, gDir, projectRoot, home, extension.ScopeGlobal, strategy)
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

func scanDir[T extension.Extension](
	ctx context.Context,
	dir, projectRoot, home string,
	scope extension.Scope,
	strategy ScanStrategy[T],
) ([]T, error) {
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

		// Handle symlinks (usually pointers to installed extensions)
		if d.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(path)
			if err == nil && info.IsDir() {
				// Try all target files for symlinked directories
				for _, targetFile := range strategy.TargetFiles() {
					if entity, err := strategy.ParseFile(filepath.Join(path, targetFile), projectRoot, home, scope); err == nil {
						entities = append(entities, entity)
						break // Parsed successfully, no need to check other targets
					}
				}
				return nil
			}
		}

		// Ensure it's one of the target files
		isTarget := false
		for _, targetFile := range strategy.TargetFiles() {
			if d.Name() == targetFile {
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
