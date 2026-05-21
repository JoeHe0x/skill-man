package app

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

func TestRenderRemoveConfirmDialog_simple(t *testing.T) {
	m := New("", "")
	m.width = 80
	m.height = 24
	skill := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "demo-skill"}}
	m.confirm.RequestRemove(panel.RemoveEffect{Skill: skill})

	out := m.confirm.RenderDialog()
	for _, want := range []string{"Remove demo-skill?", "[y/N]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in dialog, got:\n%s", want, out)
		}
	}
}
