package app

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
)

// Feature wrapper structs own feature-specific state and implement feature.Feature
// so they can consume messages through dispatchToFeatures().

// --- install feature ---

type installFeature struct {
	flow *installui.Model
	bg   *installBackground
	m    *Model
}

func (f *installFeature) Name() string { return "install" }
func (f *installFeature) Active() bool {
	return f.flow != nil || f.bg != nil
}
func (f *installFeature) Init() tea.Cmd { return nil }
func (f *installFeature) View(width, height int) string {
	if f.flow == nil {
		return ""
	}
	return f.m.renderInstallDialogArea()
}

func (f *installFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.bg != nil {
		switch msg := msg.(type) {
		case installui.ProgressTickMsg:
			_, cmd := f.m.handleInstallProgressTick(msg)
			return cmd, true
		case progress.FrameMsg:
			if cmd, ok := f.bg.handleFrame(msg); ok {
				return cmd, true
			}
		case installui.InstallDoneMsg:
			_, cmd := f.m.handleInstallUIMsg(msg)
			return cmd, true
		}
	}
	if f.flow == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case installui.SearchDoneMsg,
		installui.InstallDoneMsg,
		installui.ClosedMsg,
		installui.HintMsg,
		installui.RequestInstallMsg:
		_, cmd := f.m.handleInstallUIMsg(msg)
		return cmd, true
	case spinner.TickMsg:
		if f.flow.Searching() {
			_, cmd := f.m.handleInstallUIMsg(msg)
			return cmd, true
		}
		return nil, false
	case tea.KeyMsg:
		_, cmd := f.m.handleInstallUIMsg(msg)
		return cmd, true
	}
	return nil, false
}

// --- palette feature ---

type paletteFeature struct {
	m *Model
}

func (f *paletteFeature) Name() string                  { return "palette" }
func (f *paletteFeature) Active() bool                  { return f.m.palette != nil }
func (f *paletteFeature) Init() tea.Cmd                 { return nil }
func (f *paletteFeature) View(width, height int) string { return "" }

func (f *paletteFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.m.palette == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handlePaletteKeys(msg)
		return cmd, true
	}
	return nil, false
}

// --- help feature ---

type helpFeature struct {
	m *Model
}

func (f *helpFeature) Name() string                  { return "help" }
func (f *helpFeature) Active() bool                  { return f.m.state == stateHelpOverlay }
func (f *helpFeature) Init() tea.Cmd                 { return nil }
func (f *helpFeature) View(width, height int) string { return "" }

func (f *helpFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleHelpOverlayKeys(msg)
		return cmd, true
	case tea.MouseMsg:
		_, cmd := f.m.handleHelpOverlayMouse(msg)
		return cmd, true
	}
	return nil, false
}

// --- bind feature ---

type bindFeature struct {
	m *Model
}

func (f *bindFeature) Name() string                  { return "bind" }
func (f *bindFeature) Active() bool                  { return f.m.state == stateBindingAgent }
func (f *bindFeature) Init() tea.Cmd                 { return nil }
func (f *bindFeature) View(width, height int) string { return "" }

func (f *bindFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleBindingKeys(msg)
		return cmd, true
	}
	return nil, false
}

// --- inspect feature ---

type inspectFeature struct {
	m *Model
}

func (f *inspectFeature) Name() string                  { return "inspect" }
func (f *inspectFeature) Active() bool                  { return f.m.state == stateInspecting }
func (f *inspectFeature) Init() tea.Cmd                 { return nil }
func (f *inspectFeature) View(width, height int) string { return "" }

func (f *inspectFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleInspectingKeys(msg)
		return cmd, true
	}
	return nil, false
}

// --- agent filter feature ---

type agentFilterFeature struct {
	m *Model
}

func (f *agentFilterFeature) Name() string                  { return "agentFilter" }
func (f *agentFilterFeature) Active() bool                  { return f.m.state == stateFilteringAgent }
func (f *agentFilterFeature) Init() tea.Cmd                 { return nil }
func (f *agentFilterFeature) View(width, height int) string { return "" }

func (f *agentFilterFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleAgentFilterUpdate(msg)
		return cmd, true
	}
	return nil, false
}

// --- confirm feature ---

type confirmFeature struct {
	m *Model
}

func (f *confirmFeature) Name() string                  { return "confirm" }
func (f *confirmFeature) Active() bool                  { return f.m.state == stateConfirming }
func (f *confirmFeature) Init() tea.Cmd                 { return nil }
func (f *confirmFeature) View(width, height int) string { return "" }

func (f *confirmFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !f.Active() {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		_, cmd := f.m.handleConfirmKeys(msg)
		return cmd, true
	}
	return nil, false
}
