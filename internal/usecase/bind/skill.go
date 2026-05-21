package bind

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// DisplayAgent builds a combined agent label for a shared skills directory row.
func DisplayAgent(agents []agent.Agent) agent.Agent {
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	rep := agents[0]
	rep.Name = strings.Join(names, ", ")
	return rep
}

// NewSkillChoices builds bind rows for unique skill directory groups.
func (b Binder) NewSkillChoices(skill *skilldomain.Skill) []Choice {
	dirs := agent.UniqueSkillDirs(agent.DefaultAgents())
	choices := make([]Choice, 0, len(dirs))
	for _, dir := range dirs {
		groupAgents := agent.AgentBySkillsDir(dir)
		if len(groupAgents) == 0 {
			continue
		}
		bound := SkillDirGroupBoundOnDisk(skill, groupAgents[0], b.CWD, b.Home)
		choices = append(choices, Choice{
			Agents:   groupAgents,
			SkillDir: dir,
			Agent:    DisplayAgent(groupAgents),
			Initial:  bound,
			Desired:  bound,
		})
	}
	return choices
}

// ApplySkill applies pending skill bind/unbind changes.
func (b Binder) ApplySkill(ctx context.Context, skill *skilldomain.Skill, choices []Choice) error {
	var errs []error
	for _, c := range choices {
		if c.SkillDir != "" {
			agents := groupAgents(c)
			var err error
			if c.Desired {
				err = applySkillDirBind(ctx, b.Skills, skill, agents, b.CWD, b.Home)
			} else {
				err = applySkillDirUnbind(ctx, b.Skills, skill, agents, b.CWD, b.Home)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", c.SkillDir, err))
			}
			continue
		}
		for _, a := range groupAgents(c) {
			var err error
			if c.Desired {
				err = b.Skills.Bind(ctx, skill, a, b.CWD, b.Home)
			} else {
				err = b.Skills.Unbind(ctx, skill, a, b.CWD, b.Home)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", a.Name, err))
			}
		}
	}
	return errors.Join(errs...)
}

// SkillDirGroupBoundOnDisk is true when the skill lives in the shared dir or a symlink there points at it.
func SkillDirGroupBoundOnDisk(skill *skilldomain.Skill, rep agent.Agent, projectRoot, home string) bool {
	skillPath := filepath.Clean(skill.GetPath())
	linkPath := filepath.Clean(skillBindTargetPath(skill, rep, projectRoot, home))
	if skillPath == linkPath {
		return true
	}
	info, err := os.Lstat(linkPath)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	return symlinkResolvesTo(linkPath, skillPath)
}

func skillBindBaseDir(skill *skilldomain.Skill, projectRoot, home string) string {
	if skill.GetScope() == extension.ScopeGlobal {
		return home
	}
	return projectRoot
}

func skillBindTargetPath(skill *skilldomain.Skill, a agent.Agent, projectRoot, home string) string {
	dir := a.EntityDirs[agent.EntitySkill]
	if dir == "" {
		dir = a.SkillsDir
	}
	return filepath.Join(skillBindBaseDir(skill, projectRoot, home), dir, filepath.Base(skill.GetPath()))
}

func applySkillDirBind(ctx context.Context, mgr interface {
	Bind(context.Context, *skilldomain.Skill, agent.Agent, string, string) error
}, skill *skilldomain.Skill, agents []agent.Agent, projectRoot, home string) error {
	if len(agents) == 0 {
		return nil
	}
	rep := agents[0]
	if SkillDirGroupBoundOnDisk(skill, rep, projectRoot, home) {
		return nil
	}
	if err := mgr.Bind(ctx, skill, rep, projectRoot, home); err != nil {
		return fmt.Errorf("%s: %w", rep.EntityDirs[agent.EntitySkill], err)
	}
	return nil
}

func applySkillDirUnbind(ctx context.Context, mgr interface {
	Unbind(context.Context, *skilldomain.Skill, agent.Agent, string, string) error
}, skill *skilldomain.Skill, agents []agent.Agent, projectRoot, home string) error {
	if len(agents) == 0 {
		return nil
	}
	rep := agents[0]
	skillPath := filepath.Clean(skill.GetPath())
	linkPath := filepath.Clean(skillBindTargetPath(skill, rep, projectRoot, home))
	if skillPath == linkPath {
		return relocateSkillOutOfSharedDir(skill, projectRoot, home)
	}
	if !SkillDirGroupBoundOnDisk(skill, rep, projectRoot, home) {
		return nil
	}
	if err := mgr.Unbind(ctx, skill, rep, projectRoot, home); err != nil {
		return fmt.Errorf("%s: %w", rep.EntityDirs[agent.EntitySkill], err)
	}
	return nil
}

func relocateSkillOutOfSharedDir(skill *skilldomain.Skill, projectRoot, home string) error {
	baseDir := skillBindBaseDir(skill, projectRoot, home)
	destParent := filepath.Join(baseDir, ".skills")
	if err := os.MkdirAll(destParent, 0o755); err != nil {
		return fmt.Errorf("create .skills: %w", err)
	}
	src := filepath.Clean(skill.GetPath())
	dest := filepath.Join(destParent, filepath.Base(src))
	if filepath.Clean(dest) == src {
		return nil
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("cannot unbind: %s already exists", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("relocate skill to .skills: %w", err)
	}
	return nil
}

func symlinkResolvesTo(linkPath, skillPath string) bool {
	link, err := os.Readlink(linkPath)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(link) {
		link = filepath.Join(filepath.Dir(linkPath), link)
	}
	resolved, err := filepath.EvalSymlinks(link)
	if err != nil {
		resolved = filepath.Clean(link)
	}
	want, err := filepath.EvalSymlinks(skillPath)
	if err != nil {
		want = filepath.Clean(skillPath)
	}
	return filepath.Clean(resolved) == filepath.Clean(want)
}
