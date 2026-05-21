package install

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
	"github.com/JoeHe0x/skill-man/internal/app/strutil"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

type background struct {
	skillName string
	progress  progress.Model
}

func newBackground(skillName string, barWidth int, styles theme.Styles) *background {
	w := min(48, max(24, barWidth/2))
	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(w))
	bar.ShowPercentage = true
	_ = styles
	return &background{skillName: skillName, progress: bar}
}

func (b *background) begin() tea.Cmd {
	return tea.Batch(b.progress.SetPercent(0), progressTickCmd())
}

func nextInstallProgressPercent(current float64) float64 {
	const cap = 0.97
	if current >= cap {
		return current
	}
	delta := (1.0 - current) * 0.06
	if delta < 0.002 {
		delta = 0.002
	}
	next := current + delta
	if next > cap {
		return cap
	}
	return next
}

func (b *background) handleTick() tea.Cmd {
	p := b.progress.Percent()
	next := nextInstallProgressPercent(p)
	var cmds []tea.Cmd
	if next > p {
		cmds = append(cmds, b.progress.SetPercent(next))
	}
	cmds = append(cmds, progressTickCmd())
	return tea.Batch(cmds...)
}

func (b *background) handleFrame(msg progress.FrameMsg) (tea.Cmd, bool) {
	if b == nil {
		return nil, false
	}
	next, cmd := b.progress.Update(msg)
	b.progress = next.(progress.Model)
	return cmd, true
}

func (b *background) view(styles theme.Styles) string {
	title := styles.PanelTitle.Render("Installing " + strutil.Truncate(b.skillName, 28))
	bar := b.progress.View()
	hint := styles.Hint.Render("Estimated progress · downloading from skills.sh")
	body := lipgloss.JoinVertical(lipgloss.Left, title, bar, hint)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(0, 1).
		Render(body)
}

func progressTickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg {
		return installui.ProgressTickMsg{}
	})
}
