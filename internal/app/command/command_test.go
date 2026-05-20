package command

import (
	"context"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

func TestRemoveSkill_Execute(t *testing.T) {
	t.Parallel()

	mgr := &stubSkillManager{}
	skill := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "test-skill"}}
	cmd := &RemoveSkill{
		Skill:       skill,
		Manager:     mgr,
		ProjectRoot: "/tmp",
		Home:        "/home",
	}
	result := cmd.Execute(context.Background())
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.AffectedName != "test-skill" {
		t.Fatalf("expected affected name 'test-skill', got %q", result.AffectedName)
	}
	if !mgr.removeCalled {
		t.Fatal("expected Remove to be called")
	}
}

func TestToggleDisableSkill_Label(t *testing.T) {
	t.Parallel()

	enabled := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "s1"}}
	disabled := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "s2", Disabled: true}}

	if got := (&ToggleDisableSkill{Skill: enabled}).Label(); got != "enabled s1" {
		t.Fatalf("unexpected enabled label: %q", got)
	}
	if got := (&ToggleDisableSkill{Skill: disabled}).Label(); got != "disabled s2" {
		t.Fatalf("unexpected disabled label: %q", got)
	}
}

func TestRemoveMCPKey_Execute(t *testing.T) {
	t.Parallel()

	mgr := servicemcp.NewManager()
	members := []*mcpdomain.Server{
		{BaseExtension: extension.BaseExtension{Name: "test-mcp"}, ConfigKey: "test-mcp"},
	}
	cmd := &RemoveMCPKey{Members: members, Manager: mgr}
	result := cmd.Execute(context.Background())
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.AffectedName != "test-mcp" {
		t.Fatalf("expected affected name 'test-mcp', got %q", result.AffectedName)
	}
}

func TestCommandInterface(t *testing.T) {
	t.Parallel()

	var _ Cmd = (*RemoveSkill)(nil)
	var _ Cmd = (*ToggleDisableSkill)(nil)
	var _ Cmd = (*RemoveMCPKey)(nil)
	var _ Cmd = (*ToggleDisableMCPKey)(nil)
}

// stubSkillManager implements manager.ExtensionManager for testing.
type stubSkillManager struct {
	removeCalled bool
}

func (m *stubSkillManager) Scan(ctx context.Context, projectRoot, home string, agents []agent.Agent) ([]*skilldomain.Skill, error) {
	return nil, nil
}
func (m *stubSkillManager) Bind(ctx context.Context, ext *skilldomain.Skill, a agent.Agent, projectRoot, home string) error {
	return nil
}
func (m *stubSkillManager) Unbind(ctx context.Context, ext *skilldomain.Skill, a agent.Agent, projectRoot, home string) error {
	return nil
}
func (m *stubSkillManager) ToggleDisable(ctx context.Context, ext *skilldomain.Skill) error {
	return nil
}
func (m *stubSkillManager) Remove(ctx context.Context, ext *skilldomain.Skill, projectRoot, home string) error {
	m.removeCalled = true
	return nil
}
