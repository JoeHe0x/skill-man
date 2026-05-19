package app

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewStyles_lightDiffersFromDark(t *testing.T) {
	dark := newStyles(true)
	light := newStyles(false)
	if dark.itemTitle.GetForeground() == light.itemTitle.GetForeground() {
		t.Fatal("light and dark item title colors should differ")
	}
}

func TestColorProfileToTermenv_mapsTrueColor(t *testing.T) {
	p := colorProfileToTermenv(colorprofile.TrueColor)
	if p != termenv.TrueColor {
		t.Fatalf("expected TrueColor, got %v", p)
	}
}

func TestApplyTheme_updatesDelegates(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.listDelegate = newItemDelegate(m.styles)
	m.agentListDelegate = newItemDelegate(m.styles)
	m.list.SetDelegate(m.listDelegate)
	m.agentList.SetDelegate(m.agentListDelegate)

	_ = m.applyTheme(false)
	if m.darkTheme {
		t.Fatal("expected light theme")
	}
	if m.listDelegate.styles.itemTitle.GetForeground() != newLightStyles().itemTitle.GetForeground() {
		t.Fatal("delegate styles not updated")
	}
}

func TestConfigureTerminalColorProfile_doesNotPanic(t *testing.T) {
	configureTerminalColorProfile()
	_ = lipgloss.ColorProfile()
}
