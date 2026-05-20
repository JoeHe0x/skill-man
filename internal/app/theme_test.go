package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
)

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
	if m.listDelegate.styles.ItemTitle.GetForeground() != theme.NewStyles(false).ItemTitle.GetForeground() {
		t.Fatal("delegate styles not updated")
	}
}
