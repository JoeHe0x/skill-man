package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

// bindSession holds per-bind-flow ephemeral state (agent binding dialog).
type bindSession struct {
	skill      *skilldomain.Skill
	mcp        *mcpdomain.Server   // template for bind/unbind mutations
	mcpMembers []*mcpdomain.Server // all files for the selected config key
	agents     []agentBindChoice
}

func (s *bindSession) clear() {
	s.skill = nil
	s.mcp = nil
	s.mcpMembers = nil
	s.agents = nil
}

// agentBindChoice tracks desired vs initial bind state for one row in the bind UI.
// Skill rows group every agent that shares a skills directory (e.g. .agents/skills).
type agentBindChoice struct {
	agent      agent.Agent   // display label; MCP rows use this agent only
	agents     []agent.Agent // skill rows: all agents in the shared directory
	skillDir   string        // skill rows: EntityDirs[EntitySkill] for index lookup
	scope      extension.Scope
	configPath string // MCP only; destination config file for this row
	initial    bool
	desired    bool
}

func skillBindGroupAgents(c agentBindChoice) []agent.Agent {
	if len(c.agents) > 0 {
		return c.agents
	}
	return []agent.Agent{c.agent}
}

func newMCPBindChoices(members []*mcpdomain.Server, projectRoot, home string) []agentBindChoice {
	targets := servicemcp.ListBindTargets(projectRoot, home)
	choices := make([]agentBindChoice, 0, len(targets))
	for _, t := range targets {
		bound := mcpTargetBoundFromMembers(members, t)
		choices = append(choices, agentBindChoice{
			agent:      t.Agent,
			scope:      t.Scope,
			configPath: t.ConfigPath,
			initial:    bound,
			desired:    bound,
		})
	}
	return choices
}

// mcpTargetBoundFromMembers reports whether configKey exists in the bind target's config file.
// Each scanned member is one on-disk file; presence in scan results means the key is in that file.
func mcpTargetBoundFromMembers(members []*mcpdomain.Server, t servicemcp.BindTarget) bool {
	key := mcpConfigKeyFromMembers(members)
	for _, srv := range members {
		if !servicemcp.ConfigPathsEqual(srv.ConfigPath, t.ConfigPath) {
			continue
		}
		srvKey := srv.ConfigKey
		if srvKey == "" {
			srvKey = srv.GetName()
		}
		if key != "" && srvKey != key {
			continue
		}
		return true
	}
	return false
}

func mcpConfigKeyFromMembers(members []*mcpdomain.Server) string {
	if len(members) == 0 {
		return ""
	}
	if k := members[0].ConfigKey; k != "" {
		return k
	}
	return members[0].GetName()
}

// mcpBindTemplate picks a member to copy command/args/url from when writing new bindings.
func mcpBindTemplate(members []*mcpdomain.Server) *mcpdomain.Server {
	for _, srv := range members {
		if srv.Command != "" || srv.URL != "" {
			cp := *srv
			cp.ConfigKey = mcpConfigKeyFromMembers(members)
			return &cp
		}
	}
	if len(members) == 0 {
		return nil
	}
	cp := *members[0]
	cp.ConfigKey = mcpConfigKeyFromMembers(members)
	return &cp
}

func mcpTargetBound(srv *mcpdomain.Server, t servicemcp.BindTarget) bool {
	key := srv.ConfigKey
	if key == "" {
		key = srv.GetName()
	}

	// Explicitly empty Bindings must not fall back to top-level Agents (can be stale after dedupe).
	if srv.Bindings != nil && len(srv.Bindings) == 0 {
		return false
	}

	bindings := srv.AllBindings()
	if len(bindings) == 0 {
		return servicemcp.ConfigPathsEqual(srv.ConfigPath, t.ConfigPath) &&
			slices.Contains(srv.Agents, t.Agent.ID)
	}

	targetPath := filepath.Clean(t.ConfigPath)
	for _, b := range bindings {
		if !servicemcp.ConfigPathsEqual(b.ConfigPath, targetPath) || b.ConfigKey != key {
			continue
		}
		// Match the config file, not binding scope (Windsurf always uses a global path).
		if len(b.Agents) == 0 {
			return true
		}
		if slices.Contains(b.Agents, t.Agent.ID) {
			return true
		}
	}
	return false
}

func newSkillBindChoices(skill *skilldomain.Skill, projectRoot, home string) []agentBindChoice {
	dirs := agent.UniqueSkillDirs(agent.DefaultAgents())
	choices := make([]agentBindChoice, 0, len(dirs))
	for _, dir := range dirs {
		groupAgents := agent.AgentBySkillsDir(dir)
		if len(groupAgents) == 0 {
			continue
		}
		bound := skillDirGroupBoundOnDisk(skill, groupAgents[0], projectRoot, home)
		choices = append(choices, agentBindChoice{
			agents:   groupAgents,
			skillDir: dir,
			agent:    skillBindDisplayAgent(groupAgents),
			initial:  bound,
			desired:  bound,
		})
	}
	return choices
}

// skillDirGroupBoundOnDisk is true when the skill lives in the shared dir or a symlink there points at it.
func skillDirGroupBoundOnDisk(skill *skilldomain.Skill, rep agent.Agent, projectRoot, home string) bool {
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

func skillBindDisplayAgent(agents []agent.Agent) agent.Agent {
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	rep := agents[0]
	rep.Name = strings.Join(names, ", ")
	return rep
}

func bindChoicesToListItems(choices []agentBindChoice, projectRoot, home string) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		title := bindAgentTitle(c.agent.Name, c.desired)
		desc := bindAgentDesc(c.agent)
		if c.scope != "" {
			title = bindAgentTitle(mcpBindRowTitle(c.agent.Name, c.scope), c.desired)
			desc = servicemcp.ShortPath(home, c.configPath)
		}
		meta := c.agent.ID
		if c.skillDir != "" {
			meta = c.skillDir
		}
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: title,
			Desc:  desc,
			Meta:  meta,
		})
	}
	return items
}

func mcpBindRowTitle(name string, scope extension.Scope) string {
	return fmt.Sprintf("%s (%s)", name, scope)
}

func applyMCPBindChoices(mgr *servicemcp.Manager, srv *mcpdomain.Server, choices []agentBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		if c.scope == "" || c.configPath == "" {
			continue
		}
		label := mcpBindRowTitle(c.agent.Name, c.scope) + " → " + servicemcp.ShortPath(home, c.configPath)
		target := servicemcp.BindTarget{Agent: c.agent, Scope: c.scope, ConfigPath: c.configPath}
		var err error
		switch {
		case c.desired && !c.initial:
			err = mgr.BindAtTarget(srv, target, projectRoot, home)
		case !c.desired && c.initial:
			err = mgr.UnbindAtTarget(srv, target, projectRoot, home)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", label, err))
		}
	}
	return errors.Join(errs...)
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

// applySkillDirBind creates one symlink for the whole shared directory (e.g. .agents/skills).
func applySkillDirBind(ctx context.Context, mgr manager.ExtensionManager[*skilldomain.Skill], skill *skilldomain.Skill, agents []agent.Agent, projectRoot, home string) error {
	if len(agents) == 0 {
		return nil
	}
	rep := agents[0]
	if skillDirGroupBoundOnDisk(skill, rep, projectRoot, home) {
		return nil
	}
	if err := mgr.Bind(ctx, skill, rep, projectRoot, home); err != nil {
		return fmt.Errorf("%s: %w", rep.EntityDirs[agent.EntitySkill], err)
	}
	return nil
}

// applySkillDirUnbind removes the shared symlink, or moves the skill out of the shared dir if installed there.
func applySkillDirUnbind(ctx context.Context, mgr manager.ExtensionManager[*skilldomain.Skill], skill *skilldomain.Skill, agents []agent.Agent, projectRoot, home string) error {
	if len(agents) == 0 {
		return nil
	}
	rep := agents[0]
	skillPath := filepath.Clean(skill.GetPath())
	linkPath := filepath.Clean(skillBindTargetPath(skill, rep, projectRoot, home))
	if skillPath == linkPath {
		return relocateSkillOutOfSharedDir(skill, projectRoot, home)
	}
	if !skillDirGroupBoundOnDisk(skill, rep, projectRoot, home) {
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

func applySkillBindChoices(ctx context.Context, mgr manager.ExtensionManager[*skilldomain.Skill], skill *skilldomain.Skill, choices []agentBindChoice, projectRoot, home string) error {
	var errs []error
	for _, c := range choices {
		if c.skillDir != "" {
			agents := skillBindGroupAgents(c)
			var err error
			if c.desired {
				err = applySkillDirBind(ctx, mgr, skill, agents, projectRoot, home)
			} else {
				err = applySkillDirUnbind(ctx, mgr, skill, agents, projectRoot, home)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", c.skillDir, err))
			}
			continue
		}
		for _, a := range skillBindGroupAgents(c) {
			var err error
			if c.desired {
				err = mgr.Bind(ctx, skill, a, projectRoot, home)
			} else {
				err = mgr.Unbind(ctx, skill, a, projectRoot, home)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", a.Name, err))
			}
		}
	}
	return errors.Join(errs...)
}

func bindAgentTitle(name string, checked bool) string {
	if checked {
		return "✓ " + name
	}
	return "  " + name
}

func bindAgentDesc(a agent.Agent) string {
	if dir := agent.MCPEntityDir(a); dir != "" {
		return dir
	}
	return a.EntityDirs[agent.EntitySkill]
}
