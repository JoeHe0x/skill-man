package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	agent "skill-man/internal/domain/agent"
	skilldomain "skill-man/internal/domain/skill"
)

func BindAgent(skill skilldomain.Skill, a agent.Agent, projectRoot, home string) error {
	var baseDir string
	if skill.Scope == skilldomain.ScopeGlobal {
		if home == "" {
			return errors.New("home directory not available for global skill binding")
		}
		baseDir = home
	} else {
		if projectRoot == "" {
			return errors.New("project root not available for project skill binding")
		}
		baseDir = projectRoot
	}

	targetDir := filepath.Join(baseDir, a.EntityDirs[agent.EntitySkill])
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create agent skills dir %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, filepath.Base(skill.Path))

	// If it's the exact same directory, nothing to do (it is intrinsically bound)
	if samePath(skill.Path, targetPath) {
		return nil
	}

	// Check if target exists
	info, err := os.Lstat(targetPath)
	if err == nil {
		// Target exists. If it's a symlink to our skill, we're good.
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(targetPath)
			if err == nil {
				// Resolve link relative to targetDir if needed
				if !filepath.IsAbs(link) {
					link = filepath.Join(targetDir, link)
				}
				if samePath(link, skill.Path) {
					return nil // Already bound
				}
			}
		}
		return fmt.Errorf("target %s already exists and is not a symlink to this skill", targetPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Create relative symlink if possible, else absolute
	relPath, err := filepath.Rel(targetDir, skill.Path)
	if err != nil {
		relPath = skill.Path
	}

	return os.Symlink(relPath, targetPath)
}

func UnbindAgent(skill skilldomain.Skill, a agent.Agent, projectRoot, home string) error {
	var baseDir string
	if skill.Scope == skilldomain.ScopeGlobal {
		if home == "" {
			return errors.New("home directory not available for global skill binding")
		}
		baseDir = home
	} else {
		if projectRoot == "" {
			return errors.New("project root not available for project skill binding")
		}
		baseDir = projectRoot
	}

	targetDir := filepath.Join(baseDir, a.EntityDirs[agent.EntitySkill])
	targetPath := filepath.Join(targetDir, filepath.Base(skill.Path))

	if samePath(skill.Path, targetPath) {
		return errors.New("cannot unbind skill from its primary location")
	}

	info, err := os.Lstat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // Already unbound
		}
		return err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("target %s is not a symlink, refusing to delete", targetPath)
	}

	return os.Remove(targetPath)
}
