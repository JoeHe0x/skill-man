package app

import (
	"strings"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
)

func TestHandleMutationCompleted_clearsRemovingFooter(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.state = stateListing
	m.status = "loading"
	m.setFooterContext("Removing demo-skill...")

	updated, _ := m.handleMutationCompleted(uimsg.MutationCompleted{
		Message: "removed demo-skill",
	})
	m = mustModel(t, updated)

	if strings.Contains(m.footerContext, "Removing") {
		t.Fatalf("footer still shows remove progress: %q", m.footerContext)
	}
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading (rescan after mutation)", m.status)
	}
}

func TestHandleMutationCompleted_clearsRemovingFooterOnError(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	m.state = stateListing
	m.status = "loading"
	m.setFooterContext("Removing demo-skill...")

	updated, _ := m.handleMutationCompleted(uimsg.MutationCompleted{
		Err: errTestMutation,
	})
	m = mustModel(t, updated)

	if strings.Contains(m.footerContext, "Removing") {
		t.Fatalf("footer still shows remove progress: %q", m.footerContext)
	}
	// beginScanAllCmd runs after a failed mutation; loading takes precedence in the footer.
	if m.status != "loading" {
		t.Fatalf("status = %q, want loading (rescan after mutation error)", m.status)
	}
}

var errTestMutation = errString("remove failed")

type errString string

func (e errString) Error() string { return string(e) }
