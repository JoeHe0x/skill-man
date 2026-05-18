package skill

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	skilldomain "skill-man/internal/domain/skill"
)

func TestInitializeInstallScanAndRemoveSkill(t *testing.T) {
	workspace := t.TempDir()
	sourceRoot := filepath.Join(workspace, "source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	sourcePath, sourceName, err := InitializeSkill(sourceRoot, "My Local Skill")
	if err != nil {
		t.Fatalf("InitializeSkill returned error: %v", err)
	}
	if sourceName != "my-local-skill" {
		t.Fatalf("unexpected source skill name: %s", sourceName)
	}

	result, err := InstallLocalSkill(workspace, sourcePath, nil)
	if err != nil {
		t.Fatalf("InstallLocalSkill returned error: %v", err)
	}
	if result.Name != "my-local-skill" {
		t.Fatalf("unexpected installed skill name: %s", result.Name)
	}

	if _, err := os.Stat(filepath.Join(result.TargetPath, "SKILL.md")); err != nil {
		t.Fatalf("installed SKILL.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.TargetPath, metadataFileName)); err != nil {
		t.Fatalf("install metadata missing: %v", err)
	}

	skills, err := ScanSkills(context.Background(), workspace, "", nil)
	if err != nil {
		t.Fatalf("ScanSkills returned error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 scanned skill (installed), got %d", len(skills))
	}

	var installedPath string
	for _, skill := range skills {
		if skill.Path == result.TargetPath {
			installedPath = skill.Path
			if !skill.Managed {
				t.Fatal("expected installed skill to be marked managed")
			}
			if skill.SourceKind != "local" {
				t.Fatalf("unexpected source kind: %s", skill.SourceKind)
			}
			if skill.SourcePath != sourcePath {
				t.Fatalf("unexpected source path: %s", skill.SourcePath)
			}
			if err := RemoveSkill(skill, workspace, ""); err != nil {
				t.Fatalf("RemoveSkill returned error: %v", err)
			}
			break
		}
	}
	if installedPath == "" {
		t.Fatal("did not find installed skill in scan results")
	}

	if _, err := os.Stat(installedPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected installed path removed, stat err=%v", err)
	}

	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("expected source path to remain, stat err=%v", err)
	}
}

func TestInstallLocalSkillRejectsMissingSkillFile(t *testing.T) {
	workspace := t.TempDir()
	source := filepath.Join(workspace, "empty")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := InstallLocalSkill(workspace, source, nil); err == nil {
		t.Fatal("expected install to fail for missing SKILL.md")
	}
}

func TestUpdateSkillRefreshesInstalledContents(t *testing.T) {
	workspace := t.TempDir()
	sourceRoot := filepath.Join(workspace, "source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	sourcePath, _, err := InitializeSkill(sourceRoot, "Updater")
	if err != nil {
		t.Fatalf("InitializeSkill returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcePath, "extra.txt"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := InstallLocalSkill(workspace, sourcePath, nil)
	if err != nil {
		t.Fatalf("InstallLocalSkill returned error: %v", err)
	}

	skills, err := ScanSkills(context.Background(), workspace, "", nil)
	if err != nil {
		t.Fatalf("ScanSkills returned error: %v", err)
	}

	var installedSkill skilldomain.Skill
	found := false
	for _, skill := range skills {
		if skill.Path == result.TargetPath {
			installedSkill = skill
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find installed skill")
	}

	if err := os.Remove(filepath.Join(sourcePath, "extra.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourcePath, "fresh.txt"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(result.TargetPath, "stale.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	updateResult, err := UpdateSkill(installedSkill)
	if err != nil {
		t.Fatalf("UpdateSkill returned error: %v", err)
	}
	if updateResult.TargetPath != result.TargetPath {
		t.Fatalf("unexpected update target path: %s", updateResult.TargetPath)
	}

	if _, err := os.Stat(filepath.Join(result.TargetPath, "fresh.txt")); err != nil {
		t.Fatalf("expected fresh file after update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.TargetPath, "stale.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected stale file removed after update, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(result.TargetPath, "extra.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected removed source file to disappear after update, got err=%v", err)
	}
}
