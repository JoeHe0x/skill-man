package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

type pendingAction struct {
	name       string
	skillName  string
	skill      *skilldomain.Skill
	mcpName    string
	mcp        *mcpdomain.Server
	mcpMembers []*mcpdomain.Server
}

type confirmFeature struct {
	host    confirmHost
	pending *pendingAction
}

func (f *confirmFeature) Name() string { return "confirm" }
func (f *confirmFeature) Active() bool {
	return f.host.IsConfirming() && f.pending != nil
}
func (f *confirmFeature) Init() tea.Cmd                 { return nil }
func (f *confirmFeature) View(width, height int) string { return "" }

func (f *confirmFeature) Clear() {
	f.pending = nil
}

func (f *confirmFeature) requestRemove(eff panel.RemoveEffect) (tea.Model, tea.Cmd) {
	if eff.Skill != nil {
		f.pending = &pendingAction{
			name:      "remove",
			skillName: eff.Skill.GetName(),
			skill:     eff.Skill,
		}
	} else {
		f.pending = &pendingAction{
			name:       "remove",
			mcpName:    eff.MCPName,
			mcpMembers: eff.MCPMembers,
		}
	}
	f.host.TransitionTo(stateConfirming)
	return f.host.TeaModel(), nil
}

func (f *confirmFeature) beginRemoveConfirm() {
	f.host.SetFooterContext("y confirm · n/Esc cancel")
}

func (f *confirmFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.host.IsConfirming() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm):
			_, cmd := f.executeRemove()
			return cmd, true
		case key.Matches(msg, keys.Cancel):
			_, cmd := f.cancel()
			return cmd, true
		}
	}
	return nil, false
}

func (f *confirmFeature) cancel() (tea.Model, tea.Cmd) {
	f.Clear()
	f.host.TransitionTo(stateListing)
	f.host.SetFooterContext("Cancelled")
	return f.host.TeaModel(), nil
}

func (f *confirmFeature) executeRemove() (tea.Model, tea.Cmd) {
	if f.pending == nil || f.pending.name != "remove" {
		f.Clear()
		f.host.TransitionTo(stateListing)
		return f.host.TeaModel(), nil
	}
	mut := f.host.Mutator()
	if len(f.pending.mcpMembers) > 0 {
		members := f.pending.mcpMembers
		name := f.pending.mcpName
		f.Clear()
		f.host.TransitionTo(stateListing)
		f.host.SetStatus("loading")
		f.host.SetFooterContext(fmt.Sprintf("Removing MCP `%s`...", name))
		return f.host.TeaModel(), runCommand(&command.RemoveMCPKey{Members: members, Mutator: mut})
	}
	skill := f.pending.skill
	f.Clear()
	f.host.TransitionTo(stateListing)
	f.host.SetStatus("loading")
	f.host.SetFooterContext(fmt.Sprintf("Removing %s...", skill.GetName()))
	return f.host.TeaModel(), runCommand(&command.RemoveSkill{Skill: skill, Mutator: mut})
}

func (f *confirmFeature) renderMainOverlay() string {
	leftWidth, mainHeight, _, _ := f.host.PaneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, f.renderDialog())
}

func (f *confirmFeature) renderDialog() string {
	if f.pending == nil {
		return ""
	}
	leftWidth, _, _, _ := f.host.PaneSizes()
	dialogWidth := min(max(36, leftWidth-4), 52)
	if dialogWidth > leftWidth-2 {
		dialogWidth = max(24, leftWidth-2)
	}

	target := f.pending.skillName
	if f.pending.mcpName != "" {
		target = "MCP " + f.pending.mcpName
	}

	styles := f.host.Styles()
	body := lipgloss.JoinVertical(lipgloss.Left,
		styles.PanelTitle.Render("Remove "+truncate(target, dialogWidth-8)+"?"),
		styles.Hint.Render("[y/N]"),
	)

	return styles.ModalDanger.
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		Render(body)
}
