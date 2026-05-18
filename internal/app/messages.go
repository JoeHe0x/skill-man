package app

import "skill-man/internal/domain/skill"

type skillsScannedMsg struct {
	skills []*skill.Skill
	err    error
}

type previewLoadedMsg struct {
	content string
	err     error
	gen     int // generation; drop if stale
}

type mutationCompletedMsg struct {
	message    string
	selectName string
	err        error
}

type reselectSkillMsg struct {
	name string
}
