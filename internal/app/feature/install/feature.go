package install

import (
	"context"
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	"github.com/JoeHe0x/skill-man/internal/app/strutil"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

// Feature owns the search-and-install wizard and background install progress.
type Feature struct {
	host Host
	flow *installui.Model
	bg   *background
}

// New returns an install feature wired to host.
func New(host Host) *Feature {
	return &Feature{host: host}
}

func (f *Feature) Name() string { return "install" }
func (f *Feature) Active() bool {
	return f.flow != nil || f.bg != nil
}
func (f *Feature) Init() tea.Cmd { return nil }

func (f *Feature) View(width, height int) string {
	if f.flow == nil {
		return ""
	}
	return f.RenderDialogArea()
}

func (f *Feature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.bg != nil {
		switch msg := msg.(type) {
		case installui.ProgressTickMsg:
			return f.handleProgressTick(), true
		case progress.FrameMsg:
			if cmd, ok := f.bg.handleFrame(msg); ok {
				return cmd, true
			}
		case installui.InstallDoneMsg:
			_, cmd := f.handleCompleted(uimsg.InstallCompleted{Name: msg.Name, Err: msg.Err})
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
		_, cmd := f.handleUIMsg(msg)
		return cmd, true
	case spinner.TickMsg:
		if f.flow.Searching() {
			_, cmd := f.handleUIMsg(msg)
			return cmd, true
		}
		return nil, false
	case tea.KeyMsg:
		_, cmd := f.handleUIMsg(msg)
		return cmd, true
	}
	return nil, false
}

// ClearWizard drops the open wizard without stopping a background install.
func (f *Feature) ClearWizard() {
	f.flow = nil
}

func (f *Feature) BackgroundActive() bool {
	return f.bg != nil
}

func (f *Feature) providerForTab(tab panel.Tab) (serviceinstall.Provider, bool) {
	switch tab {
	case panel.TabSkills:
		return serviceinstall.NewSkillsCLIProvider(), true
	default:
		return nil, false
	}
}

func (f *Feature) StartFlow() (tea.Model, tea.Cmd) {
	if !f.host.ActivePanelSearchInstall() {
		return f.host.TeaModel(), f.host.FlashFooter("Search & install is not available for this tab yet")
	}
	provider, ok := f.providerForTab(f.host.ActiveTab())
	if !ok {
		return f.host.TeaModel(), f.host.FlashFooter("Search & install is not available for this tab yet")
	}
	f.host.TransitionTo(session.Installing)
	flow := installui.New(installui.Config{
		Styles:    f.host.Styles(),
		Provider:  provider,
		AgentIDs:  f.host.AgentIDs(),
		CWD:       f.host.CWD(),
		Home:      f.host.Home(),
		GetErrMsg: func() string { return f.host.ErrMsg() },
		SetErrMsg: func(s string) { f.host.ReportError(errors.New(s)) },
		ClearErr: func() {
			f.host.ClearError()
			f.host.SetStatus("ready")
		},
	})
	flow.SetSize(f.host.Width(), f.host.Height())
	f.flow = &flow
	f.syncHint()
	return f.host.TeaModel(), textinput.Blink
}

func (f *Feature) SyncHint() {
	if f.flow == nil {
		if f.BackgroundActive() {
			f.host.SetFooterContext("Installing " + f.bg.skillName + " in background")
		}
		return
	}
	if hint := f.flow.FooterHint(); hint != "" {
		f.host.SetFooterContext(hint)
	}
}

func (f *Feature) CancelFlow(hint string) {
	f.flow = nil
	f.host.TransitionTo(session.Listing)
	if hint != "" {
		f.host.SetFooterContext(hint)
	}
}

func (f *Feature) ClearFlow() {
	f.flow = nil
}

func (f *Feature) HandleUIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	return f.handleUIMsg(msg)
}

func (f *Feature) handleUIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installui.ClosedMsg:
		if f.flow == nil {
			return f.host.TeaModel(), nil
		}
		f.CancelFlow(msg.Hint)
		return f.host.TeaModel(), nil
	case installui.HintMsg:
		f.host.SetFooterContext(msg.Text)
		return f.host.TeaModel(), nil
	case installui.RequestInstallMsg:
		return f.startSelected(msg.AgentIDs, msg.Scope)
	case installui.InstallDoneMsg:
		return f.handleCompleted(uimsg.InstallCompleted{Name: msg.Name, Err: msg.Err})
	case installui.SearchDoneMsg:
		if f.flow == nil {
			return f.host.TeaModel(), nil
		}
		next, cmd := f.flow.Update(msg)
		f.flow = &next
		if f.flow.Searching() {
			f.host.SetStatus("loading")
		}
		f.syncHint()
		return f.host.TeaModel(), cmd
	}
	if f.flow == nil {
		return f.host.TeaModel(), nil
	}
	next, cmd := f.flow.Update(msg)
	f.flow = &next
	if f.flow.Searching() {
		f.host.SetStatus("loading")
	}
	f.syncHint()
	return f.host.TeaModel(), cmd
}

func (f *Feature) startSelected(agentIDs []string, scope extension.Scope) (tea.Model, tea.Cmd) {
	flow := f.flow
	if flow == nil {
		return f.host.TeaModel(), nil
	}
	candidate := flow.Selected()
	if candidate.Name == "" {
		return f.host.TeaModel(), nil
	}
	provider, ok := f.providerForTab(f.host.ActiveTab())
	if !ok {
		return f.host.TeaModel(), f.host.FlashFooter("Install provider unavailable")
	}

	leftWidth, _, _, _ := f.host.PaneSizes()
	f.bg = newBackground(candidate.Name, leftWidth, f.host.Styles())
	f.flow = nil

	f.host.TransitionTo(session.Listing)
	f.host.SetStatus("loading")
	f.syncHint()

	cwd, home := f.host.CWD(), f.host.Home()
	installCmd := func() tea.Msg {
		name, err := provider.Install(context.Background(), cwd, home, candidate, agentIDs, scope)
		return installui.InstallDoneMsg{Name: name, Err: err}
	}
	return f.host.TeaModel(), tea.Batch(f.bg.begin(), installCmd)
}

func (f *Feature) RenderDialogArea() string {
	if f.flow == nil {
		return ""
	}
	leftWidth, mainHeight, _, _ := f.host.PaneSizes()
	f.flow.SetSize(leftWidth, f.host.Height())
	return f.flow.PlaceInPane(leftWidth, mainHeight)
}

// HandleBackgroundFrame forwards progress frame messages to the background bar.
func (f *Feature) HandleBackgroundFrame(msg progress.FrameMsg) (tea.Cmd, bool) {
	if f.bg == nil {
		return nil, false
	}
	return f.bg.handleFrame(msg)
}

func (f *Feature) handleProgressTick() tea.Cmd {
	if f.bg == nil {
		return nil
	}
	return f.bg.handleTick()
}

func (f *Feature) RenderBackgroundOverlay(main string, mainHeight int) string {
	if f.bg == nil {
		return main
	}
	leftWidth, _, _, _ := f.host.PaneSizesFor(mainHeight)
	corner := lipgloss.NewStyle().Width(leftWidth).PaddingLeft(1).Render(f.bg.view(f.host.Styles()))
	progressH := lipgloss.Height(corner)
	contentH := max(4, mainHeight-progressH)
	top := strutil.ClipLines(main, contentH)
	return lipgloss.JoinVertical(lipgloss.Left, top, corner)
}

func (f *Feature) handleCompleted(msg uimsg.InstallCompleted) (tea.Model, tea.Cmd) {
	f.bg = nil
	f.ClearFlow()
	if f.host.State() == session.Installing {
		f.host.TransitionTo(session.Listing)
	}
	if msg.Err != nil {
		f.host.ReportError(msg.Err)
		return f.host.TeaModel(), f.host.BeginScanAllCmd()
	}
	f.host.ClearError()
	f.host.SetStatus("ready")
	return f.host.TeaModel(), tea.Batch(
		f.host.FlashFooter("✓ Installed "+msg.Name+" — back in skill list"),
		tea.Sequence(
			f.host.BeginScanAllCmd(),
			func() tea.Msg { return uimsg.ReselectSkill{Name: msg.Name} },
		),
	)
}

func (f *Feature) syncHint() { f.SyncHint() }

// WizardOpen reports whether the install wizard is visible.
func (f *Feature) WizardOpen() bool { return f.flow != nil }

// WizardSearching reports whether the wizard is waiting on registry search.
func (f *Feature) WizardSearching() bool {
	return f.flow != nil && f.flow.Searching()
}

// ApplyTheme forwards theme to the open wizard.
func (f *Feature) ApplyTheme(styles theme.Styles) {
	if f.flow != nil {
		f.flow.ApplyTheme(styles)
	}
}

// ShortHelp returns footer key hints for the install wizard.
func (f *Feature) ShortHelp() []key.Binding {
	if f.flow == nil {
		return nil
	}
	return f.flow.ShortHelp()
}

// StartSelected begins installing the wizard's selected candidate (exported for tests).
func (f *Feature) StartSelected(agentIDs []string, scope extension.Scope) (tea.Model, tea.Cmd) {
	return f.startSelected(agentIDs, scope)
}

// PrepareWizardSelected sets the selected candidate on an open wizard (tests only).
func (f *Feature) PrepareWizardSelected(candidate domaininstall.Candidate) {
	if f.flow == nil {
		return
	}
	prepared := f.flow.WithSelected(candidate)
	f.flow = &prepared
}
