package panel

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

// ScanCmd wraps Core.Scan as a Bubble Tea command.
func ScanCmd(p Core, cwd, home string, agents []agent.Agent) tea.Cmd {
	return func() tea.Msg {
		return p.Scan(context.Background(), cwd, home, agents)
	}
}

// ScanAllCmd triggers scan for every panel in the registry.
func ScanAllCmd(r *Registry, cwd, home string, agents []agent.Agent) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(r.Tabs()))
	for _, tab := range r.Tabs() {
		cmds = append(cmds, ScanCmd(r.Get(tab), cwd, home, agents))
	}
	return tea.Batch(cmds...)
}

// SyncPreviewCmd loads preview markdown asynchronously and returns PreviewLoadedMsg.
func SyncPreviewCmd(p Core, selected Item, width int, previewGen *int) tea.Cmd {
	if previewGen == nil {
		return nil
	}
	tab := p.Tab()
	switch {
	case selected.Kind == ItemSkill && p.SelectedSkill(selected):
	case selected.Kind == ItemMCP && p.SelectedMCP(selected):
	default:
		return nil
	}
	*previewGen++
	gen := *previewGen
	return func() tea.Msg {
		content, err := p.PreviewMarkdown(selected, width)
		return PreviewLoadedMsg{Tab: tab, Content: content, Err: err, Gen: gen}
	}
}
