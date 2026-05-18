package service

import (
	"fmt"
	"os"
	"path/filepath"

	"skill-man/internal/domain"
)

func RemoveSkill(skill domain.Skill, projectRoot, home string) error {
	cleanPath := filepath.Clean(skill.Path)

	// First, find and remove any symlinks pointing to this skill from all agents
	for _, agent := range domain.DefaultAgents() {
		for _, baseDir := range []string{projectRoot, home} {
			if baseDir == "" {
				continue
			}
			targetDir := filepath.Join(baseDir, agent.SkillsDir)
			targetPath := filepath.Join(targetDir, filepath.Base(cleanPath))
			if samePath(targetPath, cleanPath) {
				continue // Skip the actual source dir for now
			}
			info, err := os.Lstat(targetPath)
			if err == nil && info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(targetPath)
				if err == nil {
					if !filepath.IsAbs(link) {
						link = filepath.Join(targetDir, link)
					}
					if samePath(link, cleanPath) {
						os.Remove(targetPath)
					}
				}
			}
		}
	}

	// Then, remove the actual skill directory
	if _, err := os.Stat(filepath.Join(cleanPath, "SKILL.md")); err != nil {
		if _, err := os.Stat(filepath.Join(cleanPath, "SKILL.md.disabled")); err != nil {
			return fmt.Errorf("missing SKILL.md under %s", cleanPath)
		}
	}

	return os.RemoveAll(cleanPath)
}
