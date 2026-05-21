package prompt

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

type promptModel struct {
	input  textinput.Model
	label  string
	action Action
}

func newPromptModel(label, placeholder string, action Action) *promptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Prompt = ""
	ti.Focus()
	return &promptModel{
		input:  ti,
		label:  label,
		action: action,
	}
}

// Feature owns the temporary footer text input.
type Feature struct {
	host   Host
	prompt *promptModel
}

// New returns a prompt feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string { return "prompt" }
func (f *Feature) Active() bool { return f.prompt != nil }
func (f *Feature) Init() tea.Cmd {
	return nil
}
func (f *Feature) View(width, height int) string { return "" }

func (f *Feature) Show(label, placeholder string, action Action) tea.Cmd {
	f.prompt = newPromptModel(label, placeholder, action)
	return textinput.Blink
}

func (f *Feature) Hide() {
	f.prompt = nil
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.prompt == nil {
		return nil, false
	}
	keys := uikeys.Default
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Home):
			f.Hide()
			if f.host.State() == session.Installing {
				f.host.CancelInstallFlow("Install cancelled")
				return nil, true
			}
			f.host.SetFooterContext("Cancelled")
			return nil, true
		case key.Matches(msg, keys.Enter) || msg.Type == tea.KeyEnter:
			text := f.prompt.input.Value()
			act := f.prompt.action
			f.Hide()
			if act == nil {
				return nil, true
			}
			return act(text), true
		}
		var cmd tea.Cmd
		f.prompt.input, cmd = f.prompt.input.Update(msg)
		return cmd, true
	default:
		return nil, false
	}
}

// PromptLabel returns the active prompt label (for tests).
func (f *Feature) PromptLabel() string {
	if f.prompt == nil {
		return ""
	}
	return f.prompt.label
}

func (f *Feature) RenderFooter() string {
	if f.prompt == nil {
		return ""
	}
	styles := f.host.Styles()
	label := styles.HintBold.Render(f.prompt.label + ": ")
	input := f.prompt.input.View()
	helpLine := styles.Hint.Render("Enter=confirm  Esc=cancel")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, label, input),
		helpLine,
	)
}
