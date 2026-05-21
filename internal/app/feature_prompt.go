package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// promptModel is a temporary text input for commands that need user text.
type promptModel struct {
	input  textinput.Model
	label  string
	action func(m *Model, text string) tea.Cmd
}

func newPromptModel(label, placeholder string, action func(m *Model, text string) tea.Cmd) *promptModel {
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

type promptFeature struct {
	m      *Model
	prompt *promptModel
}

func (f *promptFeature) Name() string { return "prompt" }
func (f *promptFeature) Active() bool { return f.prompt != nil }
func (f *promptFeature) Init() tea.Cmd {
	return nil
}
func (f *promptFeature) View(width, height int) string { return "" }

func (f *promptFeature) Show(label, placeholder string, action func(m *Model, text string) tea.Cmd) tea.Cmd {
	f.prompt = newPromptModel(label, placeholder, action)
	return textinput.Blink
}

func (f *promptFeature) Hide() {
	f.prompt = nil
}

func (f *promptFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.prompt == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Home):
			f.Hide()
			if f.m.state == stateInstalling {
				f.m.cancelInstallFlow("Install cancelled")
				return nil, true
			}
			f.m.setFooterContext("Cancelled")
			return nil, true
		case key.Matches(msg, keys.Enter) || msg.Type == tea.KeyEnter:
			text := f.prompt.input.Value()
			act := f.prompt.action
			f.Hide()
			if act == nil {
				return nil, true
			}
			return act(f.m, text), true
		}
		var cmd tea.Cmd
		f.prompt.input, cmd = f.prompt.input.Update(msg)
		return cmd, true
	default:
		return nil, false
	}
}

func (f *promptFeature) renderFooter() string {
	if f.prompt == nil {
		return ""
	}
	label := f.m.styles.HintBold.Render(f.prompt.label + ": ")
	input := f.prompt.input.View()
	helpLine := f.m.styles.Hint.Render("Enter=confirm  Esc=cancel")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, label, input),
		helpLine,
	)
}
