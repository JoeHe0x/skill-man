package extension

import (
	"context"
	"testing"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
)

func TestRemoveSkill(t *testing.T) {
	t.Parallel()

	mgr := &stubSkillManager{}
	mut := NewMutator(mgr, servicemcp.NewManager(), "/tmp", "/home")
	skill := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "test-skill"}}

	out := mut.RemoveSkill(context.Background(), skill)
	if out.Err != nil {
		t.Fatalf("unexpected error: %v", out.Err)
	}
	if out.Kind != KindSkill {
		t.Fatalf("kind = %v, want KindSkill", out.Kind)
	}
	if out.AffectedName != "test-skill" {
		t.Fatalf("expected affected name 'test-skill', got %q", out.AffectedName)
	}
	if !mgr.removeCalled {
		t.Fatal("expected Remove to be called")
	}
}

func TestToggleDisableSkill_Message(t *testing.T) {
	t.Parallel()

	mut := NewMutator(&stubSkillManager{}, servicemcp.NewManager(), "/tmp", "/home")
	enabled := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "s1"}}
	disabled := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "s2", Disabled: true}}

	if out := mut.ToggleDisableSkill(context.Background(), enabled); out.Message != "disabled s1" {
		t.Fatalf("unexpected enabled message: %q", out.Message)
	}
	if out := mut.ToggleDisableSkill(context.Background(), disabled); out.Message != "enabled s2" {
		t.Fatalf("unexpected disabled message: %q", out.Message)
	}
}

func TestRemoveMCPKey(t *testing.T) {
	t.Parallel()

	mut := NewMutator(&stubSkillManager{}, servicemcp.NewManager(), "/tmp", "/home")
	members := []*mcpdomain.Server{
		{BaseExtension: extension.BaseExtension{Name: "test-mcp"}, ConfigKey: "test-mcp"},
	}
	out := mut.RemoveMCPKey(context.Background(), members)
	if out.Err != nil {
		t.Fatalf("unexpected error: %v", out.Err)
	}
	if out.Kind != KindMCP {
		t.Fatalf("kind = %v, want KindMCP", out.Kind)
	}
	if out.AffectedName != "test-mcp" {
		t.Fatalf("expected affected name 'test-mcp', got %q", out.AffectedName)
	}
}

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
