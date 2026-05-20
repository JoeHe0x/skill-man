package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/extension"
)

type stubExt struct {
	extension.BaseExtension
}

func (s stubExt) GetTools() []string { return nil }

func TestToggleDisable_diskState(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	enabled := filepath.Join(dir, "SKILL.md")
	disabled := enabled + ".disabled"

	write := func(path string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("---\nname: x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("enable from disabled file", func(t *testing.T) {
		write(disabled)
		ext := stubExt{BaseExtension: extension.BaseExtension{
			ConfigPath: enabled, // stale: UI thinks enabled path
			Disabled:   false,
		}}
		if err := ToggleDisable(ext); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(enabled); err != nil {
			t.Fatalf("want SKILL.md: %v", err)
		}
		if _, err := os.Stat(disabled); !os.IsNotExist(err) {
			t.Fatalf("want no .disabled: %v", err)
		}
	})

	t.Run("disable from enabled file", func(t *testing.T) {
		write(enabled)
		ext := stubExt{BaseExtension: extension.BaseExtension{
			ConfigPath: disabled, // stale: UI thinks disabled path
			Disabled:   true,
		}}
		if err := ToggleDisable(ext); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(disabled); err != nil {
			t.Fatalf("want SKILL.md.disabled: %v", err)
		}
	})

	t.Run("both exist", func(t *testing.T) {
		write(enabled)
		write(disabled)
		ext := stubExt{BaseExtension: extension.BaseExtension{ConfigPath: enabled}}
		if err := ToggleDisable(ext); err == nil {
			t.Fatal("expected error when both files exist")
		}
	})
}
