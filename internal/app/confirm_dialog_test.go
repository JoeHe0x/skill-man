package app

import (
	"strings"
	"testing"
)

func TestRenderRemoveConfirmDialog_simple(t *testing.T) {
	m := New("", "")
	m.width = 80
	m.height = 24
	m.confirm.pending = &pendingAction{name: "remove", skillName: "demo-skill"}

	out := m.confirm.renderDialog()
	for _, want := range []string{"Remove demo-skill?", "[y/N]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in dialog, got:\n%s", want, out)
		}
	}
}
