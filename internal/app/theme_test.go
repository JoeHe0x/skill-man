package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/list"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

func TestApplyTheme_updatesDelegates(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.MainDel = list.NewDelegate(m.styles)
	m.AgentDel = list.NewDelegate(m.styles)
	m.Main.SetDelegate(m.MainDel)
	m.Agent.SetDelegate(m.AgentDel)

	_ = m.applyTheme(false)
	if m.darkTheme {
		t.Fatal("expected light theme")
	}
	if m.MainDel.Styles().ItemTitle.GetForeground() != theme.NewStyles(false).ItemTitle.GetForeground() {
		t.Fatal("delegate styles not updated")
	}
}
