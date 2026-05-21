package command

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

// RemoveSkill removes a skill from disk.
type RemoveSkill struct {
	Skill   *skilldomain.Skill
	Mutator usecase.Mutator
}

func (c *RemoveSkill) Label() string { return "removed " + c.Skill.GetName() }

func (c *RemoveSkill) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.RemoveSkill(ctx, c.Skill))
}

// ToggleDisableSkill toggles a skill's disabled state.
type ToggleDisableSkill struct {
	Skill   *skilldomain.Skill
	Mutator usecase.Mutator
}

func (c *ToggleDisableSkill) Label() string {
	if c.Skill.IsDisabled() {
		return "disabled " + c.Skill.GetName()
	}
	return "enabled " + c.Skill.GetName()
}

func (c *ToggleDisableSkill) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.ToggleDisableSkill(ctx, c.Skill))
}

// AddSkill installs a skill from a local source path.
type AddSkill struct {
	Source  string
	Agents  []agent.Agent
	Mutator usecase.Mutator
}

func (c *AddSkill) Label() string { return "installed " + c.Source }

func (c *AddSkill) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.AddSkill(ctx, c.Source, c.Agents))
}

// InitSkill creates a new skill template on disk.
type InitSkill struct {
	Name    string
	Mutator usecase.Mutator
}

func (c *InitSkill) Label() string { return "created " + c.Name }

func (c *InitSkill) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.InitSkill(ctx, c.Name))
}

// UpdateSkill updates a single managed local skill from its source.
type UpdateSkill struct {
	Skill   *skilldomain.Skill
	Mutator usecase.Mutator
}

func (c *UpdateSkill) Label() string { return "updated " + c.Skill.GetName() }

func (c *UpdateSkill) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.UpdateSkill(ctx, c.Skill))
}

// UpdateAllSkills updates every managed local skill concurrently.
type UpdateAllSkills struct {
	Skills  []*skilldomain.Skill
	Mutator usecase.Mutator
}

func (c *UpdateAllSkills) Label() string { return "updated all skills" }

func (c *UpdateAllSkills) Execute(ctx context.Context) Result {
	return resultFrom(c.Mutator.UpdateAllSkills(ctx, c.Skills))
}
