package skill

import (
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

type Scope = extension.Scope

const (
	ScopeProject = extension.ScopeProject
	ScopeGlobal  = extension.ScopeGlobal
)

type Skill struct {
	extension.BaseExtension
	Tools []string
}
