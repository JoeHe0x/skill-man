package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"skill-man/internal/domain"
)

type InstallResult struct {
	Name       string
	SourcePath string
	TargetPath string
}

func InstallLocalSkill(workspaceRoot, source string, agents []domain.Agent) (InstallResult, error) {
	sourcePath, err := resolveSkillSource(source)
	if err != nil {
		return InstallResult{}, fmt.Errorf("resolve skill source: %w", err)
	}

	skill, err := ParseSkillFile(filepath.Join(sourcePath, "SKILL.md"))
	if err != nil {
		return InstallResult{}, fmt.Errorf("parse skill file: %w", err)
	}

	// Install to the first available agent's project skills dir, or fall back to .skills
	targetRoot := filepath.Join(workspaceRoot, ".skills")
	if len(agents) > 0 {
		targetRoot = filepath.Join(workspaceRoot, agents[0].SkillsDir)
	}

	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		return InstallResult{}, err
	}

	targetPath := filepath.Join(targetRoot, sanitizeInstallName(skill.Name))
	if samePath(sourcePath, targetPath) {
		return InstallResult{}, fmt.Errorf("source is already installed at %s", targetPath)
	}
	if _, err := os.Stat(targetPath); err == nil {
		return InstallResult{}, fmt.Errorf("target already exists: %s", targetPath)
	}

	if err := copyDir(sourcePath, targetPath); err != nil {
		return InstallResult{}, err
	}
	if err := writeInstallMetadata(targetPath, installMetadata{
		Name:        skill.Name,
		SourceKind:  "local",
		SourcePath:  sourcePath,
		InstalledAt: time.Now().UTC(),
	}); err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		Name:       skill.Name,
		SourcePath: sourcePath,
		TargetPath: targetPath,
	}, nil
}

func resolveSkillSource(source string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", errors.New("source is required")
	}

	abs, err := filepath.Abs(source)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		if filepath.Base(abs) != "SKILL.md" {
			return "", fmt.Errorf("source must be a directory or SKILL.md file: %s", source)
		}
		abs = filepath.Dir(abs)
	}

	if _, err := os.Stat(filepath.Join(abs, "SKILL.md")); err != nil {
		return "", fmt.Errorf("source does not contain SKILL.md: %s", abs)
	}

	return abs, nil
}

func copyDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(target, 0o755)
		}

		dst := filepath.Join(target, rel)
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(dst, 0o755)
		}
		if d.Name() == metadataFileName {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(path, dst, info.Mode())
	})
}

func copyFile(source, target string, mode fs.FileMode) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	return a == b
}

func sanitizeInstallName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-", "\"", "", "'", "")
	name = replacer.Replace(name)

	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-':
			if !lastDash {
				b.WriteRune(r)
			}
			lastDash = true
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "skill"
	}
	return out
}
