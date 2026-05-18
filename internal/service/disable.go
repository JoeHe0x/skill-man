package service

import (
	"fmt"
	"os"
	"strings"

	"skill-man/internal/domain"
)

func ToggleDisableSkill(skill domain.Skill) error {
	oldPath := skill.SkillFilePath
	var newPath string
	if skill.Disabled {
		if !strings.HasSuffix(oldPath, ".disabled") {
			return fmt.Errorf("skill is marked disabled but path does not end with .disabled: %s", oldPath)
		}
		newPath = strings.TrimSuffix(oldPath, ".disabled")
	} else {
		newPath = oldPath + ".disabled"
	}
	return os.Rename(oldPath, newPath)
}
