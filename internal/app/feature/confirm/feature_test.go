package confirm

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
)

type testHost struct {
	width, height int
}

func (h *testHost) IsConfirming() bool              { return true }
func (h *testHost) TransitionTo(session.State) bool { return true }
func (h *testHost) SetFooterContext(string)         {}
func (h *testHost) SetStatus(string)                {}
func (h *testHost) PaneSizes() (int, int, int, int) { return h.width, h.height, 40, h.height }
func (h *testHost) Styles() theme.Styles            { return theme.NewStyles(true) }
func (h *testHost) Mutator() usecase.Mutator        { return usecase.Mutator{} }
func (h *testHost) TeaModel() tea.Model             { return nil }

func TestRenderRemoveConfirmDialog_simple(t *testing.T) {
	h := &testHost{width: 80, height: 24}
	f := New(h)
	skill := &skilldomain.Skill{BaseExtension: extension.BaseExtension{Name: "demo-skill"}}
	f.RequestRemove(panel.RemoveEffect{Skill: skill})

	out := f.RenderDialog()
	for _, want := range []string{"Remove demo-skill?", "[y/N]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in dialog, got:\n%s", want, out)
		}
	}
}
