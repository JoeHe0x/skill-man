package command

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// RemoveSkill removes a skill from disk.
type RemoveSkill struct {
	Skill       *skilldomain.Skill
	Manager     manager.ExtensionManager[*skilldomain.Skill]
	ProjectRoot string
	Home        string
}

func (c *RemoveSkill) Label() string { return "removed " + c.Skill.GetName() }

func (c *RemoveSkill) Execute(ctx context.Context) Result {
	if err := c.Manager.Remove(ctx, c.Skill, c.ProjectRoot, c.Home); err != nil {
		return Result{Err: err}
	}
	return Result{
		AffectedName: c.Skill.GetName(),
		Message:      fmt.Sprintf("removed %s", c.Skill.GetName()),
	}
}

// ToggleDisableSkill toggles a skill's disabled state.
type ToggleDisableSkill struct {
	Skill   *skilldomain.Skill
	Manager manager.ExtensionManager[*skilldomain.Skill]
}

func (c *ToggleDisableSkill) Label() string {
	if c.Skill.IsDisabled() {
		return "disabled " + c.Skill.GetName()
	}
	return "enabled " + c.Skill.GetName()
}

func (c *ToggleDisableSkill) Execute(ctx context.Context) Result {
	if err := c.Manager.ToggleDisable(ctx, c.Skill); err != nil {
		return Result{Err: err}
	}
	action := "disabled"
	if c.Skill.IsDisabled() {
		action = "enabled"
	}
	return Result{
		AffectedName: c.Skill.GetName(),
		Message:      fmt.Sprintf("%s %s", action, c.Skill.GetName()),
	}
}

// AddSkill installs a skill from a local source path.
type AddSkill struct {
	Source string
	CWD    string
	Agents []agent.Agent
}

func (c *AddSkill) Label() string { return "installed " + c.Source }

func (c *AddSkill) Execute(ctx context.Context) Result {
	result, err := service.InstallLocalSkill(c.CWD, "", extension.ScopeProject, c.Source, c.Agents)
	if err != nil {
		return Result{Err: err}
	}
	return Result{
		AffectedName: result.Name,
		Message:      fmt.Sprintf("installed %s -> %s", result.Name, result.TargetPath),
	}
}

// InitSkill creates a new skill template on disk.
type InitSkill struct {
	Root string
	Name string
}

func (c *InitSkill) Label() string { return "created " + c.Name }

func (c *InitSkill) Execute(ctx context.Context) Result {
	path, createdName, err := service.InitializeSkill(c.Root, c.Name)
	if err != nil {
		return Result{Err: err}
	}
	return Result{
		AffectedName: createdName,
		Message:      fmt.Sprintf("created skill template at %s", path),
	}
}

// UpdateSkill updates a single managed local skill from its source.
type UpdateSkill struct {
	Skill *skilldomain.Skill
}

func (c *UpdateSkill) Label() string { return "updated " + c.Skill.GetName() }

func (c *UpdateSkill) Execute(ctx context.Context) Result {
	result, err := service.UpdateSkill(*c.Skill)
	if err != nil {
		return Result{Err: err}
	}
	return Result{
		AffectedName: result.Name,
		Message:      fmt.Sprintf("updated %s from %s", result.Name, result.SourcePath),
	}
}

// UpdateAllSkills updates every managed local skill concurrently.
type UpdateAllSkills struct {
	Skills []*skilldomain.Skill
}

func (c *UpdateAllSkills) Label() string { return "updated all skills" }

func (c *UpdateAllSkills) Execute(ctx context.Context) Result {
	var g errgroup.Group
	var mu sync.Mutex
	updated := 0
	firstName := ""

	for _, skill := range c.Skills {
		if !skill.IsManaged() || skill.SourceKind != "local" {
			continue
		}
		g.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic updating skill %s: %v", skill.GetName(), r)
				}
			}()
			if _, err := service.UpdateSkill(*skill); err != nil {
				return err
			}
			mu.Lock()
			updated++
			if firstName == "" {
				firstName = skill.GetName()
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Result{Err: err}
	}
	if updated == 0 {
		return Result{Message: "no managed local skills available to update"}
	}
	return Result{
		AffectedName: firstName,
		Message:      fmt.Sprintf("updated %d skill(s)", updated),
	}
}
