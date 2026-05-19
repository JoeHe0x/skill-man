package app

import (
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

type installSearchCompletedMsg struct {
	results []domaininstall.Candidate
	err     error
}

type installCompletedMsg struct {
	name string
	err  error
}

type installProgressTickMsg struct{}

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
