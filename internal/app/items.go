package app

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"

	"skill-man/internal/commands"
	skilldomain "skill-man/internal/domain/skill"
)

type itemKind int

const (
	itemKindCommand itemKind = iota
	itemKindSkill
	itemKindMessage
)

type listItem struct {
	kind    itemKind
	title   string
	desc    string
	meta    string
	command commands.Spec
	skill   *skilldomain.Skill
}

func (i listItem) FilterValue() string {
	return strings.ToLower(strings.Join([]string{i.title, i.desc, i.meta}, " "))
}

func (i listItem) Title() string {
	return i.title
}

func (i listItem) Description() string {
	return i.desc
}

func commandItems(specs []commands.Spec) []list.Item {
	items := make([]list.Item, 0, len(specs))
	for _, spec := range specs {
		meta := spec.Usage
		if spec.Dangerous {
			meta += " | dangerous"
		}
		items = append(items, listItem{
			kind:    itemKindCommand,
			title:   "/" + spec.Name,
			desc:    spec.Summary,
			meta:    meta,
			command: spec,
		})
	}
	return items
}

func skillItems(skills []*skilldomain.Skill, agentFilter []string) []list.Item {
	if len(skills) == 0 {
		return []list.Item{listItem{
			kind:  itemKindMessage,
			title: "No skills found",
			desc:  "Run /reload after adding local skills or change the workspace root.",
			meta:  "empty",
		}}
	}

	items := make([]list.Item, 0, len(skills))
	for _, skill := range skills {
		if !skillMatchesFilter(skill, agentFilter) {
			continue
		}

		tools := "no tools"
		if len(skill.Tools) > 0 {
			tools = strings.Join(skill.Tools, ", ")
		}

		agents := "no agents"
		if len(skill.GetAgents()) > 0 {
			agents = strings.Join(skill.GetAgents(), ", ")
		}

		management := "unmanaged"
		if skill.IsManaged() {
			management = skill.SourceKind
		}

		title := skill.GetName()
		if skill.GetScope() == skilldomain.ScopeGlobal {
			title = skill.GetName() + " [global]"
		}
		if skill.IsDisabled() {
			title = "[x] " + title
		}

		items = append(items, listItem{
			kind:  itemKindSkill,
			title: title,
			desc:  skill.GetDescription(),
			meta:  fmt.Sprintf("%s | agents: %s | %s | %s | %s", skill.GetScope(), agents, tools, management, skill.GetUpdatedAt().Format(time.DateOnly)),
			skill: skill,
		})
	}
	return items
}

func skillMatchesFilter(skill *skilldomain.Skill, agentFilter []string) bool {
	if len(agentFilter) == 0 || slices.Contains(agentFilter, "all") {
		return true
	}
	for _, id := range agentFilter {
		if slices.Contains(skill.GetAgents(), id) {
			return true
		}
	}
	return false
}
