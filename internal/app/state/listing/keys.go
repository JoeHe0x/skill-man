package listing

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/uikeys"
)

// HandleKeys routes keys in listing/home/searching states.
func HandleKeys(h Host, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m := h.TeaModel()
	keys := uikeys.Default

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Home):
		h.ClearError()
		if h.MainFilterState() != list.Unfiltered {
			return m, h.MainUpdate(msg)
		}
		h.TransitionTo(session.Home)
		if preview := h.StaticPreview(); preview != "" {
			h.SetPreviewContent(preview)
			return m, nil
		}
		return m, h.SyncSelectionPreview()

	case key.Matches(msg, keys.HelpToggle):
		h.ToggleHelpAll()
		return m, nil

	case key.Matches(msg, keys.HelpScreen):
		return h.OpenHelpOverlay()

	case key.Matches(msg, keys.Palette):
		return h.OpenCommandPalette()

	case key.Matches(msg, keys.Tab):
		h.SetFocusedList()
		return m, h.SwitchExtensionTab(false)

	case key.Matches(msg, keys.ShiftTab):
		h.SetFocusedList()
		return m, h.SwitchExtensionTab(true)

	case key.Matches(msg, keys.Down):
		h.SetFocusedList()
		return m, tea.Batch(h.MainUpdate(msg), h.SyncSelectionPreview())

	case key.Matches(msg, keys.Up):
		h.SetFocusedList()
		return m, tea.Batch(h.MainUpdate(msg), h.SyncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		h.SetFocusedPreview()
		return m, h.PreviewUpdate(msg)

	case key.Matches(msg, keys.List):
		h.TransitionTo(session.Listing)
		return m, h.SyncSelectionPreview()

	case key.Matches(msg, keys.Find), key.Matches(msg, keys.Filter):
		return h.StartListFilter()

	case key.Matches(msg, keys.Agent):
		return h.OpenAgentFilter()

	case key.Matches(msg, keys.Reload):
		return m, h.BeginScanAllCmd()

	case key.Matches(msg, keys.Update):
		return h.HandleUpdate()

	case key.Matches(msg, keys.Enter):
		return h.HandleInspectSelected()

	case key.Matches(msg, keys.Bind):
		return h.HandleBindSelected()

	case key.Matches(msg, keys.Disable):
		return h.HandleDisableSelected()

	case key.Matches(msg, keys.Delete):
		return h.HandleRemoveSelected()

	case key.Matches(msg, keys.Add):
		return h.StartInstallFlow()

	case key.Matches(msg, keys.Init):
		if h.ActiveTab() == panel.TabSkills && h.ActivePanel().Capabilities().Init {
			return h.ShowInitPrompt()
		}
		h.SetFooterContext("Init is only available on the Skills tab")
		return m, nil
	}

	return m, nil
}
