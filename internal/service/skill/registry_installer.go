package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

// RegistrySnapshotFile is one file from a skills.sh download bundle.
type RegistrySnapshotFile struct {
	Path     string
	Contents string
}

// InstallRegistrySkill writes a registry snapshot into each selected agent's skills directory.
func InstallRegistrySkill(workspaceRoot, home string, scope extension.Scope, source string, files []RegistrySnapshotFile, agents []agent.Agent) (InstallResult, error) {
	if len(files) == 0 {
		return InstallResult{}, errors.New("registry snapshot is empty")
	}

	skillName, err := skillNameFromSnapshot(files)
	if err != nil {
		return InstallResult{}, err
	}

	baseDir := workspaceRoot
	if scope == extension.ScopeGlobal {
		if home == "" {
			return InstallResult{}, errors.New("home directory not available for global install")
		}
		baseDir = home
	}

	targetRoots := registryTargetRoots(baseDir, agents)
	if len(targetRoots) == 0 {
		return InstallResult{}, errors.New("no install target directories for selected agents")
	}

	sanitized := sanitizeInstallName(skillName)
	var first InstallResult

	for i, targetRoot := range targetRoots {
		targetPath := filepath.Join(targetRoot, sanitized)
		if _, err := os.Stat(targetPath); err == nil {
			return InstallResult{}, fmt.Errorf("target already exists: %s", targetPath)
		} else if !errors.Is(err, os.ErrNotExist) {
			return InstallResult{}, fmt.Errorf("check target %s: %w", targetPath, err)
		}

		if err := os.MkdirAll(targetRoot, 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("create install dir: %w", err)
		}
		if err := writeSnapshotFiles(targetPath, files); err != nil {
			return InstallResult{}, fmt.Errorf("write skill to %s: %w", targetPath, err)
		}
		if err := writeInstallMetadata(targetPath, installMetadata{
			Name:        skillName,
			SourceKind:  "registry",
			SourcePath:  source,
			InstalledAt: time.Now().UTC(),
		}); err != nil {
			return InstallResult{}, fmt.Errorf("write install metadata: %w", err)
		}

		if i == 0 {
			first = InstallResult{
				Name:       skillName,
				SourcePath: source,
				TargetPath: targetPath,
			}
			if resolved, err := filepath.EvalSymlinks(targetPath); err == nil {
				first.TargetPath = resolved
			}
		}
	}

	return first, nil
}

func registryTargetRoots(baseDir string, agents []agent.Agent) []string {
	if len(agents) == 0 {
		return []string{filepath.Join(baseDir, ".skills")}
	}
	seen := map[string]bool{}
	var roots []string
	for _, dir := range agent.UniqueSkillDirs(agents) {
		root := filepath.Join(baseDir, dir)
		if seen[root] {
			continue
		}
		seen[root] = true
		roots = append(roots, root)
	}
	return roots
}

func skillNameFromSnapshot(files []RegistrySnapshotFile) (string, error) {
	for _, f := range files {
		if !strings.EqualFold(filepathBaseName(f.Path), "SKILL.md") {
			continue
		}
		tmp, err := os.MkdirTemp("", "skill-man-skill-*")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(tmp)

		skillPath := filepath.Join(tmp, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(f.Contents), 0o644); err != nil {
			return "", err
		}
		skill, err := ParseSkillFile(skillPath)
		if err != nil {
			return "", fmt.Errorf("parse SKILL.md: %w", err)
		}
		if skill.Name == "" {
			return "", errors.New("SKILL.md has no skill name")
		}
		return skill.Name, nil
	}
	return "", errors.New("snapshot missing SKILL.md")
}

func writeSnapshotFiles(targetDir string, files []RegistrySnapshotFile) error {
	targetDir = filepath.Clean(targetDir)
	for _, f := range files {
		rel := filepath.FromSlash(strings.TrimPrefix(strings.ReplaceAll(f.Path, "\\", "/"), "/"))
		rel = filepath.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
			return fmt.Errorf("invalid file path in snapshot: %q", f.Path)
		}
		dest := filepath.Join(targetDir, rel)
		if !pathWithin(targetDir, dest) {
			return fmt.Errorf("invalid file path in snapshot: %q", f.Path)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, []byte(f.Contents), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func pathWithin(base, target string) bool {
	base = filepath.Clean(base)
	target = filepath.Clean(target)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func filepathBaseName(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
