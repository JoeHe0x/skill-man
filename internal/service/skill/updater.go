package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	skilldomain "skill-man/internal/domain/skill"
)

type UpdateResult struct {
	Name       string
	SourcePath string
	TargetPath string
}

func UpdateSkill(skill skilldomain.Skill) (UpdateResult, error) {
	if !skill.Managed {
		return UpdateResult{}, fmt.Errorf("skill is not managed by skill-man: %s", skill.Name)
	}
	if skill.SourceKind != "local" {
		return UpdateResult{}, fmt.Errorf("unsupported source kind for update: %s", skill.SourceKind)
	}

	sourcePath, err := resolveSkillSource(skill.SourcePath)
	if err != nil {
		return UpdateResult{}, err
	}
	targetPath := filepath.Clean(skill.Path)

	if samePath(sourcePath, targetPath) {
		return UpdateResult{}, errors.New("source and target paths are identical")
	}

	if err := clearDirContents(targetPath); err != nil {
		return UpdateResult{}, err
	}
	if err := copyDir(sourcePath, targetPath); err != nil {
		return UpdateResult{}, err
	}
	if err := writeInstallMetadata(targetPath, installMetadata{
		Name:        skill.Name,
		SourceKind:  skill.SourceKind,
		SourcePath:  sourcePath,
		InstalledAt: skill.UpdatedAt.UTC(),
	}); err != nil {
		return UpdateResult{}, err
	}

	return UpdateResult{
		Name:       skill.Name,
		SourcePath: sourcePath,
		TargetPath: targetPath,
	}, nil
}

func clearDirContents(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == metadataFileName {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
