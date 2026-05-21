package extension

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	serviceskill "github.com/JoeHe0x/skill-man/internal/service/skill"
)

// RemoveSkill removes a skill from disk.
func (m Mutator) RemoveSkill(ctx context.Context, sk *skilldomain.Skill) Outcome {
	if err := m.Skills.Remove(ctx, sk, m.CWD, m.Home); err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: sk.GetName(),
		Message:      fmt.Sprintf("removed %s", sk.GetName()),
	}
}

// ToggleDisableSkill toggles a skill's disabled state.
func (m Mutator) ToggleDisableSkill(ctx context.Context, sk *skilldomain.Skill) Outcome {
	if err := m.Skills.ToggleDisable(ctx, sk); err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	action := "disabled"
	if sk.IsDisabled() {
		action = "enabled"
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: sk.GetName(),
		Message:      fmt.Sprintf("%s %s", action, sk.GetName()),
	}
}

// AddSkill installs a skill from a local source path.
func (m Mutator) AddSkill(ctx context.Context, source string, agents []agent.Agent) Outcome {
	result, err := serviceskill.InstallLocalSkill(m.CWD, "", extension.ScopeProject, source, agents)
	if err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: result.Name,
		Message:      fmt.Sprintf("installed %s -> %s", result.Name, result.TargetPath),
	}
}

// InitSkill creates a new skill template on disk.
func (m Mutator) InitSkill(ctx context.Context, name string) Outcome {
	path, createdName, err := serviceskill.InitializeSkill(m.CWD, name)
	if err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: createdName,
		Message:      fmt.Sprintf("created skill template at %s", path),
	}
}

// UpdateSkill updates a single managed local skill from its source.
func (m Mutator) UpdateSkill(ctx context.Context, sk *skilldomain.Skill) Outcome {
	result, err := serviceskill.UpdateSkill(*sk)
	if err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: result.Name,
		Message:      fmt.Sprintf("updated %s from %s", result.Name, result.SourcePath),
	}
}

// UpdateAllSkills updates every managed local skill concurrently.
func (m Mutator) UpdateAllSkills(ctx context.Context, skills []*skilldomain.Skill) Outcome {
	var g errgroup.Group
	var mu sync.Mutex
	updated := 0
	firstName := ""

	for _, sk := range skills {
		if !sk.IsManaged() || sk.SourceKind != "local" {
			continue
		}
		g.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic updating skill %s: %v", sk.GetName(), r)
				}
			}()
			if _, err := serviceskill.UpdateSkill(*sk); err != nil {
				return err
			}
			mu.Lock()
			updated++
			if firstName == "" {
				firstName = sk.GetName()
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Outcome{Kind: KindSkill, Err: err}
	}
	if updated == 0 {
		return Outcome{Kind: KindSkill, Message: "no managed local skills available to update"}
	}
	return Outcome{
		Kind:         KindSkill,
		AffectedName: firstName,
		Message:      fmt.Sprintf("updated %d skill(s)", updated),
	}
}
