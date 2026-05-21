package inspect

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
)

// EnterFromItem opens inspect mode or syncs MCP preview for the selected item.
func EnterFromItem(h Host, item panel.Item) (tea.Model, tea.Cmd) {
	if !h.ActivePanel().Capabilities().Inspect || !item.CanInspect() {
		if !h.ActivePanel().Capabilities().Inspect {
			h.SetFooterContext("Inspect is not available for this tab")
		}
		return h.TeaModel(), nil
	}
	eff, ok := item.InspectEffect()
	if !ok {
		return h.TeaModel(), nil
	}
	if eff.SkillPath != "" {
		h.TransitionTo(session.Inspecting)
		h.SetTreeRoot(eff.SkillPath)
		h.SetFooterContext("Inspecting skill files")
		sel := h.TreeSelected()
		if sel.Path != "" && !sel.IsDir {
			return h.TeaModel(), h.PreviewFileCmd(sel.Path)
		}
		return h.TeaModel(), nil
	}
	width := h.PreviewWidth()
	if width == 0 {
		width = max(40, h.AppWidth()/2)
	}
	pi := panel.Item{
		Kind:       panel.ItemMCP,
		MCPKey:     eff.MCPKey,
		MCPMembers: eff.MCPMembers,
	}
	return h.TeaModel(), panel.SyncPreviewCmd(h.ActivePanel(), pi, width, h.PreviewGenPtr())
}
