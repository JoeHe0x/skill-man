package app

import (
	"testing"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
)

func TestInstallCancel_viaUIMsg(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))
	updated, _ := m.startInstallFlow()
	m = mustModel(t, updated)

	m.install.cancel = func() {}
	_ = m.install.flow.BeginInstall()

	updated, _ = m.handleInstallUIMsg(installui.CancelInstallMsg{})
	m = mustModel(t, updated)
	if m.install.flow.Installing() {
		t.Fatal("expected installing cleared after cancel msg")
	}
	if m.install.cancel != nil {
		t.Fatal("expected cancel func cleared")
	}
}
