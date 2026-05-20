package theme

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewStyles_lightDiffersFromDark(t *testing.T) {
	dark := NewStyles(true)
	light := NewStyles(false)
	if dark.ItemTitle.GetForeground() == light.ItemTitle.GetForeground() {
		t.Fatal("light and dark item title colors should differ")
	}
}

func TestColorProfileToTermenv(t *testing.T) {
	if got := colorProfileToTermenv(colorprofile.TrueColor); got != termenv.TrueColor {
		t.Fatalf("expected TrueColor, got %v", got)
	}
	if got := colorProfileToTermenv(colorprofile.ANSI256); got != termenv.ANSI256 {
		t.Fatalf("expected ANSI256, got %v", got)
	}
	if got := colorProfileToTermenv(colorprofile.ANSI); got != termenv.ANSI {
		t.Fatalf("expected ANSI, got %v", got)
	}
	if got := colorProfileToTermenv(colorprofile.Ascii); got != termenv.Ascii {
		t.Fatalf("expected Ascii, got %v", got)
	}
}

func TestConfigureColorProfile_doesNotPanic(t *testing.T) {
	ConfigureColorProfile()
	_ = lipgloss.ColorProfile()
}
