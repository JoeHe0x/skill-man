package app

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

func TestAgentFilterDialogSelectsAgent(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := mustModel(t, New(root, home))
	if m.agentDisplay() != "all" {
		t.Fatalf("expected initial filter 'all', got %q", m.agentDisplay())
	}

	updated, _ := m.handleOpenAgentFilter()
	m2 := mustModel(t, updated)
	if m2.state != stateFilteringAgent {
		t.Fatalf("expected stateFilteringAgent, got %v", m2.state)
	}

	claudeIdx := -1
	for i, item := range m2.agentList.Items() {
		li, ok := item.(panel.Item)
		if ok && li.Meta == "claude-code" {
			claudeIdx = i
			break
		}
	}
	if claudeIdx < 0 {
		t.Fatal("expected claude-code in filter list when .claude/skills exists")
	}

	m2.agentList.Select(claudeIdx)
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m2.handleAgentFilterKeys(enter)
	m3 := mustModel(t, updated)

	if m3.state != stateHome {
		t.Fatalf("expected to return to home state, got %v", m3.state)
	}
	if m3.agentDisplay() != "claude-code" {
		t.Fatalf("expected claude-code filter, got %q", m3.agentDisplay())
	}
}

func TestAgentFilterDialogHidesMissingSkillDirs(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()

	m := mustModel(t, New(root, home))
	updated, _ := m.handleOpenAgentFilter()
	m2 := mustModel(t, updated)

	agentRows := 0
	for _, item := range m2.agentList.Items() {
		li, ok := item.(panel.Item)
		if ok && li.Meta != "all" {
			agentRows++
		}
	}
	if agentRows != 0 {
		t.Fatalf("expected no agents without local dirs, got %d", agentRows)
	}

	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	visible := agent.AgentsWithLocalSkillDir(m2.allAgents, root, home)
	if len(visible) == 0 {
		t.Fatal("expected claude-code after creating .claude/skills")
	}
}

func TestAgentFilterDialogCancel(t *testing.T) {
	m := mustModel(t, New("/tmp", "/home/test"))

	updated, _ := m.handleOpenAgentFilter()
	m2 := mustModel(t, updated)
	m2.agentList.Select(5)

	esc := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ = m2.handleAgentFilterKeys(esc)
	m3 := mustModel(t, updated)

	if m3.state != stateHome {
		t.Fatalf("expected to return to home, got %v", m3.state)
	}
	if m3.agentDisplay() != "all" {
		t.Fatalf("expected filter unchanged after cancel, got %q", m3.agentDisplay())
	}
}
