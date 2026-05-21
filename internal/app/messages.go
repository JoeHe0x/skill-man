package app

import (
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

type installCompletedMsg struct {
	name string
	err  error
}

type mutationCompletedMsg struct {
	message    string
	selectName string
	err        error
	kind       usecase.Kind
}

type reselectSkillMsg struct {
	name string
}

type reselectMCPMsg struct {
	name string
}
