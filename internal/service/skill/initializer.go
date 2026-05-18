package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func InitializeSkill(root, name string) (string, string, error) {
	dirName := sanitizeSkillName(name)
	if dirName == "" {
		dirName = "new-skill"
	}

	skillDir := filepath.Join(root, dirName)
	if _, err := os.Stat(skillDir); err == nil {
		return "", "", fmt.Errorf("target already exists: %s", skillDir)
	}

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return "", "", err
	}

	readme := fmt.Sprintf("# %s\n\nDescribe what this skill does, when to use it, and any constraints.\n", dirName)
	skillMD := fmt.Sprintf(`---
name: %s
description: Describe this skill in one sentence.
license: MIT
---

# %s

## When to Use This Skill

- Add the situations where this skill should be invoked.

## Usage

- Document the expected workflow here.
`, dirName, dirName)

	if err := os.WriteFile(filepath.Join(skillDir, "README.md"), []byte(readme), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		return "", "", err
	}

	return skillDir, dirName, nil
}

func sanitizeSkillName(name string) string {
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

	return strings.Trim(b.String(), "-")
}
