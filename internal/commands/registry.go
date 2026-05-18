package commands

import (
	"sort"
)

type Spec struct {
	Name        string
	Aliases     []string
	Usage       string
	Summary     string
	Dangerous   bool
	Implemented bool
}

type Registry struct {
	specs   []Spec
	byName  map[string]Spec
	ordered []string
}

func NewRegistry() *Registry {
	specs := []Spec{
		{Name: "help", Aliases: []string{"h"}, Usage: "/help", Summary: "Show commands and keybindings.", Implemented: true},
		{Name: "list", Aliases: []string{"ls"}, Usage: "/list", Summary: "List installed skills.", Implemented: true},
		{Name: "find", Usage: "/find [query]", Summary: "Search installed skills.", Implemented: true},
		{Name: "inspect", Usage: "/inspect <skill>", Summary: "Focus a skill and open its preview.", Implemented: true},
		{Name: "reload", Usage: "/reload", Summary: "Rescan local skills.", Implemented: true},
		{Name: "add", Usage: "/add <source>", Summary: "Install a skill source.", Implemented: true},
		{Name: "remove", Aliases: []string{"rm"}, Usage: "/remove <skill>", Summary: "Remove an installed skill.", Dangerous: true, Implemented: true},
		{Name: "update", Usage: "/update [skill]", Summary: "Update installed skills.", Implemented: true},
		{Name: "init", Usage: "/init [name]", Summary: "Create a new SKILL.md template.", Implemented: true},
		{Name: "agent", Usage: "/agent [id|all]", Summary: "Filter skills by agent.", Implemented: true},
		{Name: "quit", Aliases: []string{"q"}, Usage: "/quit", Summary: "Exit skill-man.", Implemented: true},
	}

	byName := make(map[string]Spec, len(specs)*2)
	ordered := make([]string, 0, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
		ordered = append(ordered, spec.Name)
		for _, alias := range spec.Aliases {
			byName[alias] = spec
		}
	}
	sort.Strings(ordered)

	return &Registry{
		specs:   specs,
		byName:  byName,
		ordered: ordered,
	}
}

func (r *Registry) Specs() []Spec {
	out := make([]Spec, len(r.specs))
	copy(out, r.specs)
	return out
}
