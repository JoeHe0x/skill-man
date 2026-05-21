package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	"github.com/JoeHe0x/skill-man/internal/domain/skill"
)

func TestInspectEsc_footerNotScanning(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "demo-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := mustModel(t, New(root, "/home/test"))
	m.status = "ready"
	m.scan.Pending = 0
	m.panels.Get(panel.TabSkills).ApplyScan(panel.SkillsScan(
		[]*skill.Skill{{
			BaseExtension: extension.BaseExtension{Name: "demo-skill", Path: skillDir},
		}}, nil))
	m.refreshActiveList()

	updated, _ := m.HandleInspectSelected()
	m2 := mustModel(t, updated)
	if m2.state != stateInspecting {
		t.Fatalf("state = %v, want inspecting", m2.state)
	}
	m2.status = "loading"

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m3 := mustModel(t, updated)

	if m3.state != stateListing {
		t.Fatalf("state = %v, want listing", m3.state)
	}
	if m3.status != "ready" {
		t.Fatalf("status = %q, want ready", m3.status)
	}
	if strings.Contains(m3.footerContext, "Scanning") {
		t.Fatalf("footer must not show scan loading after esc: %q", m3.footerContext)
	}
}
