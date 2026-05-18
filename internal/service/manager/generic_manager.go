package manager

import (
	"context"

	"skill-man/internal/domain/agent"
	"skill-man/internal/domain/extension"
)

type genericManager[T extension.Extension] struct {
	strategy ScanStrategy[T]
	binding  *CommonBinding[T]
}

func NewManager[T extension.Extension](strategy ScanStrategy[T]) ExtensionManager[T] {
	return &genericManager[T]{
		strategy: strategy,
		binding:  &CommonBinding[T]{Strategy: strategy},
	}
}

func (m *genericManager[T]) Scan(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]T, error) {
	return ScanExtensions(ctx, projectRoot, home, agents, m.strategy)
}

func (m *genericManager[T]) Bind(ctx context.Context, ext T, a agent.Agent, projectRoot, home string) error {
	return m.binding.BindAgent(ext, a, projectRoot, home)
}

func (m *genericManager[T]) Unbind(ctx context.Context, ext T, a agent.Agent, projectRoot, home string) error {
	return m.binding.UnbindAgent(ext, a, projectRoot, home)
}

func (m *genericManager[T]) ToggleDisable(ctx context.Context, ext T) error {
	return ToggleDisable(ext)
}

func (m *genericManager[T]) Remove(ctx context.Context, ext T, projectRoot, home string) error {
	return m.binding.RemoveExtension(ext, projectRoot, home)
}
