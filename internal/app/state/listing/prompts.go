package listing

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

// PromptHost extends Host with prompt and filter helpers used by listing flows.
type PromptHost interface {
	Host
	ShowPrompt(label, placeholder string, action func(text string) tea.Cmd) tea.Cmd
	HidePrompt()
	SetFooterContext(string)
	SetStatus(string)
	FlashFooter(string) tea.Cmd
	AgentIDs() []string
	SetAgentIDs([]string)
	RefreshActiveList()
	SetMainListItems([]panel.Item)
	ActiveAgents() []agent.Agent
	Mutator() usecase.Mutator
}

// ShowFindPrompt opens the search prompt for the active panel.
func ShowFindPrompt(h PromptHost) (tea.Model, tea.Cmd) {
	if !h.ActivePanel().Capabilities().Find {
		h.SetFooterContext("Find is not available for this tab")
		return h.TeaModel(), nil
	}
	return h.TeaModel(), h.ShowPrompt("Find", "search query...", func(text string) tea.Cmd {
		h.HidePrompt()
		text = strings.TrimSpace(text)
		h.TransitionTo(session.Searching)
		if text == "" {
			h.RefreshActiveList()
			return tea.Batch(h.FlashFooter("Search cancelled"), h.SyncSelectionPreview())
		}
		items := h.ActivePanel().SearchItems(text, h.AgentIDs())
		h.SetFooterContext(fmt.Sprintf("find: %q → %d result(s)", text, panel.VisibleListCount(items)))
		h.SetMainListItems(items)
		return h.SyncSelectionPreview()
	})
}

// ShowAddPrompt opens the add-skill source prompt.
func ShowAddPrompt(h PromptHost) (tea.Model, tea.Cmd) {
	return h.TeaModel(), h.ShowPrompt("Add source", "path or SKILL.md ...", func(text string) tea.Cmd {
		h.HidePrompt()
		source := strings.TrimSpace(text)
		if source == "" {
			return h.FlashFooter("Add cancelled")
		}
		h.SetStatus("loading")
		h.SetFooterContext(fmt.Sprintf("Installing from %s...", source))
		return command.Run(&command.AddSkill{Source: source, Agents: h.ActiveAgents(), Mutator: h.Mutator()})
	})
}

// ShowInitPrompt opens the new-skill template prompt.
func ShowInitPrompt(h PromptHost) (tea.Model, tea.Cmd) {
	return h.TeaModel(), h.ShowPrompt("Init name", "new-skill (enter for default)", func(text string) tea.Cmd {
		h.HidePrompt()
		name := strings.TrimSpace(text)
		if name == "" {
			name = "new-skill"
		}
		h.SetStatus("loading")
		h.SetFooterContext(fmt.Sprintf("Creating skill template: %s", name))
		return command.Run(&command.InitSkill{Name: name, Mutator: h.Mutator()})
	})
}

// SetAgentFilter applies a single agent filter id ("all" or agent id).
func SetAgentFilter(h PromptHost, id string) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" || id == "all" {
		h.SetAgentIDs([]string{"all"})
		return
	}
	if _, ok := agent.AgentByID(id); ok {
		h.SetAgentIDs([]string{id})
	}
}
