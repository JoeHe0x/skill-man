package app

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

func TestBrowseShortHelp_includesDeleteOnSkillsTab(t *testing.T) {
	m := New("", "")
	m.activeTab = panel.TabSkills
	m.width = 120

	help := m.browseShortHelp(false)
	found := false
	for _, b := range help {
		if strings.Contains(b.Help().Key, "del") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected del binding in browse short help")
	}
}
