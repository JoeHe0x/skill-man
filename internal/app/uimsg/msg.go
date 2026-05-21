// Package uimsg defines Bubble Tea messages shared across app layers.
package uimsg

import usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"

// InstallCompleted is emitted when a background install finishes.
type InstallCompleted struct {
	Name string
	Err  error
}

// MutationCompleted is emitted when a command mutation finishes.
type MutationCompleted struct {
	Message    string
	SelectName string
	Err        error
	Kind       usecase.Kind
}

// ReselectSkill asks the list to reselect a skill by name after scan.
type ReselectSkill struct {
	Name string
}

// ReselectMCP asks the list to reselect an MCP config key after scan.
type ReselectMCP struct {
	Name string
}
