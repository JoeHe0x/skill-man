package confirm

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/command"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/strutil"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
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

// Feature owns the remove confirmation dialog.
type Feature struct {
	host    Host
	pending *pendingAction
}

// New returns a confirm feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string { return "confirm" }
func (f *Feature) Active() bool {
	return f.host.IsConfirming() && f.pending != nil
}
func (f *Feature) Init() tea.Cmd                 { return nil }
func (f *Feature) View(width, height int) string { return "" }

func (f *Feature) Clear() {
	f.pending = nil
}

func (f *Feature) RequestRemove(eff panel.RemoveEffect) (tea.Model, tea.Cmd) {
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
	f.host.TransitionTo(session.Confirming)
	return f.host.TeaModel(), nil
}

func (f *Feature) BeginRemoveConfirm() {
	f.host.SetFooterContext("y confirm · n/Esc cancel")
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.host.IsConfirming() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := uikeys.Default
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

func (f *Feature) cancel() (tea.Model, tea.Cmd) {
	f.Clear()
	f.host.TransitionTo(session.Listing)
	f.host.SetFooterContext("Cancelled")
	return f.host.TeaModel(), nil
}

func (f *Feature) executeRemove() (tea.Model, tea.Cmd) {
	if f.pending == nil || f.pending.name != "remove" {
		f.Clear()
		f.host.TransitionTo(session.Listing)
		return f.host.TeaModel(), nil
	}
	mut := f.host.Mutator()
	if len(f.pending.mcpMembers) > 0 {
		members := f.pending.mcpMembers
		name := f.pending.mcpName
		f.Clear()
		f.host.TransitionTo(session.Listing)
		f.host.SetStatus("loading")
		f.host.SetFooterContext(fmt.Sprintf("Removing MCP `%s`...", name))
		return f.host.TeaModel(), command.Run(&command.RemoveMCPKey{Members: members, Mutator: mut})
	}
	skill := f.pending.skill
	f.Clear()
	f.host.TransitionTo(session.Listing)
	f.host.SetStatus("loading")
	f.host.SetFooterContext(fmt.Sprintf("Removing %s...", skill.GetName()))
	return f.host.TeaModel(), command.Run(&command.RemoveSkill{Skill: skill, Mutator: mut})
}

// HasPending reports whether a confirm action is queued.
func (f *Feature) HasPending() bool { return f.pending != nil }

// RenderDialog renders the confirm box (for tests and overlay).
func (f *Feature) RenderDialog() string {
	return f.renderDialog()
}

func (f *Feature) RenderMainOverlay() string {
	leftWidth, mainHeight, _, _ := f.host.PaneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, f.renderDialog())
}

func (f *Feature) renderDialog() string {
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
		styles.PanelTitle.Render("Remove "+strutil.Truncate(target, dialogWidth-8)+"?"),
		styles.Hint.Render("[y/N]"),
	)

	return styles.ModalDanger.
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		Render(body)
}
