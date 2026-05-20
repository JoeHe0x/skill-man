package app

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

// installBackground tracks a skills CLI install that continues after the wizard closes.
type installBackground struct {
	skillName string
	progress  progress.Model
}

func newInstallBackground(skillName string, barWidth int, styles theme.Styles) *installBackground {
	w := min(48, max(24, barWidth/2))
	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(w))
	bar.ShowPercentage = true
	_ = styles
	return &installBackground{skillName: skillName, progress: bar}
}

func (b *installBackground) begin() tea.Cmd {
	return tea.Batch(b.progress.SetPercent(0), installProgressTickCmd())
}

func (b *installBackground) handleTick() tea.Cmd {
	var cmds []tea.Cmd
	if b.progress.Percent() < 0.9 {
		cmds = append(cmds, b.progress.IncrPercent(0.04))
	}
	cmds = append(cmds, installProgressTickCmd())
	return tea.Batch(cmds...)
}

func (b *installBackground) handleFrame(msg progress.FrameMsg) (tea.Cmd, bool) {
	if !b.active() {
		return nil, false
	}
	next, cmd := b.progress.Update(msg)
	b.progress = next.(progress.Model)
	return cmd, true
}

func (b *installBackground) active() bool {
	return b != nil
}

func (b *installBackground) view(styles theme.Styles) string {
	title := styles.PanelTitle.Render("Installing " + truncate(b.skillName, 28))
	bar := b.progress.View()
	hint := styles.Hint.Render("Running in background — browse while you wait")
	body := lipgloss.JoinVertical(lipgloss.Left, title, bar, hint)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(0, 1).
		Render(body)
}

func installProgressTickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg {
		return installui.ProgressTickMsg{}
	})
}

func (m *Model) backgroundInstallActive() bool {
	return m.install != nil && m.install.bg != nil
}

func (m *Model) handleInstallProgressTick(installui.ProgressTickMsg) (tea.Model, tea.Cmd) {
	if m.install.bg == nil {
		return m, nil
	}
	return m, m.install.bg.handleTick()
}

func (m *Model) renderBackgroundInstallOverlay(main string, mainHeight int) string {
	if m.install.bg == nil {
		return main
	}
	leftWidth, _, _, _ := m.paneSizesFor(mainHeight)
	corner := lipgloss.NewStyle().Width(leftWidth).PaddingLeft(1).Render(m.install.bg.view(m.styles))
	progressH := lipgloss.Height(corner)
	contentH := max(4, mainHeight-progressH)
	top := clipLines(main, contentH)
	return lipgloss.JoinVertical(lipgloss.Left, top, corner)
}
