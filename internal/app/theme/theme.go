package theme

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// DetectedMsg is sent once the terminal color profile has been detected.
type DetectedMsg struct {
	Dark bool
}

// DetectCmd returns a command that detects the terminal's color profile.
func DetectCmd() tea.Cmd {
	return func() tea.Msg {
		ConfigureColorProfile()
		return DetectedMsg{Dark: lipgloss.HasDarkBackground()}
	}
}

// ConfigureColorProfile detects and sets the terminal's color profile.
func ConfigureColorProfile() {
	p := colorprofile.Detect(os.Stdout, os.Environ())
	lipgloss.SetColorProfile(colorProfileToTermenv(p))
}

func colorProfileToTermenv(p colorprofile.Profile) termenv.Profile {
	switch p {
	case colorprofile.TrueColor:
		return termenv.TrueColor
	case colorprofile.ANSI256:
		return termenv.ANSI256
	case colorprofile.ANSI:
		return termenv.ANSI
	default:
		return termenv.Ascii
	}
}
