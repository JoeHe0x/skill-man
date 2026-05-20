package installui

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

type dirChoice struct {
	skillDir string
	agents   []agent.Agent
	desired  bool
}

func newDirChoices(agentFilter []string) []dirChoice {
	byDir := map[string][]agent.Agent{}
	for _, a := range agent.DefaultAgents() {
		dir := a.EntityDirs[agent.EntitySkill]
		if dir == "" {
			continue
		}
		byDir[dir] = append(byDir[dir], a)
	}
	dirs := make([]string, 0, len(byDir))
	for dir := range byDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	wantAll := len(agentFilter) == 0 || slices.Contains(agentFilter, "all")
	choices := make([]dirChoice, 0, len(dirs))
	for _, dir := range dirs {
		agents := byDir[dir]
		slices.SortFunc(agents, func(a, b agent.Agent) int {
			return strings.Compare(a.Name, b.Name)
		})
		desired := false
		if !wantAll {
			for _, id := range agentFilter {
				if slices.ContainsFunc(agents, func(a agent.Agent) bool { return a.ID == id }) {
					desired = true
					break
				}
			}
		}
		choices = append(choices, dirChoice{skillDir: dir, agents: agents, desired: desired})
	}
	return choices
}

func dirChoicesToItems(choices []dirChoice) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: dirTitle(c.skillDir, c.desired),
			Desc:  formatDirAgents(c.agents),
			Meta:  c.skillDir,
		})
	}
	return items
}

func dirTitle(skillDir string, checked bool) string {
	if checked {
		return "✓ " + skillDir
	}
	return "  " + skillDir
}

func formatDirAgents(agents []agent.Agent) string {
	if len(agents) == 0 {
		return ""
	}
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	if len(names) <= 5 {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:5], ", ") + fmt.Sprintf(" +%d more", len(names)-5)
}

func selectedAgentIDs(targets []dirChoice) []string {
	seen := map[string]bool{}
	var ids []string
	for _, t := range targets {
		if !t.desired {
			continue
		}
		for _, a := range t.agents {
			if seen[a.ID] {
				continue
			}
			seen[a.ID] = true
			ids = append(ids, a.ID)
		}
	}
	return ids
}

func resultsToItems(results []domaininstall.Candidate) []list.Item {
	items := make([]list.Item, 0, len(results))
	for _, c := range results {
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: c.Name,
			Desc:  c.Source,
			Meta:  resultInstallMeta(c),
		})
	}
	return items
}

func resultInstallMeta(c domaininstall.Candidate) string {
	if c.Local {
		return "local"
	}
	return strings.TrimSpace(c.Installs)
}

func listHeightForItems(items []list.Item) int {
	h := 3
	allMessages := len(items) > 0
	for _, it := range items {
		if li, ok := it.(panel.Item); !ok || li.Kind != panel.ItemMessage {
			allMessages = false
			break
		}
	}
	if allMessages {
		return 1
	}
	for _, it := range items {
		li, ok := it.(panel.Item)
		if !ok || len(li.DetailLines) == 0 {
			continue
		}
		need := 2 + len(li.DetailLines)
		if need > h {
			h = need
		}
	}
	return h
}
