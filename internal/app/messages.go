package app

import (
	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

type installCompletedMsg struct {
	name string
	err  error
}

type mutationCompletedMsg struct {
	message    string
	selectName string
	err        error
	targetTab  panel.Tab
}

type reselectSkillMsg struct {
	name string
}

type reselectMCPMsg struct {
	name string
}
