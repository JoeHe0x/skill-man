package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/install"
)

func TestStartInstallSelected_returnsToListingWithBackgroundJob(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.startInstallFlow()
	m = mustModel(t, updated)
	if m.install.flow == nil {
		t.Fatal("expected install flow")
	}

	prepared := m.install.flow.WithSelected(install.Candidate{
		Name:   "demo",
		Source: "owner/repo@demo",
	})
	m.install.flow = &prepared

	updated, cmd := m.install.startSelected([]string{"cursor"}, extension.ScopeProject)
	m = mustModel(t, updated)

	if m.state != stateListing {
		t.Fatalf("expected listing after starting install, got %v", m.state)
	}
	if m.install.flow != nil {
		t.Fatal("wizard should close when background install starts")
	}
	if m.install.bg == nil {
		t.Fatal("expected background progress job")
	}
	if cmd == nil {
		t.Fatal("expected progress + install commands")
	}
}
