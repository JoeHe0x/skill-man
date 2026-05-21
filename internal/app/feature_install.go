package app

import (
	"context"
	"errors"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

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
	return f.renderDialogArea()
}

func (f *installFeature) Update(msg tea.Msg) (tea.Cmd, bool) {
	if f.bg != nil {
		switch msg := msg.(type) {
		case installui.ProgressTickMsg:
			return f.handleProgressTick(), true
		case progress.FrameMsg:
			if cmd, ok := f.bg.handleFrame(msg); ok {
				return cmd, true
			}
		case installui.InstallDoneMsg:
			_, cmd := f.handleCompleted(installCompletedMsg{name: msg.Name, err: msg.Err})
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

func (f *installFeature) providerForTab(tab panel.Tab) (serviceinstall.Provider, bool) {
	switch tab {
	case panel.TabSkills:
		return serviceinstall.NewSkillsCLIProvider(), true
	default:
		return nil, false
	}
}

func (f *installFeature) startFlow() (tea.Model, tea.Cmd) {
	m := f.m
	if !m.activePanel().Capabilities().SearchInstall {
		return m, m.flashFooter("Search & install is not available for this tab yet")
	}
	provider, ok := f.providerForTab(m.activeTab)
	if !ok {
		return m, m.flashFooter("Search & install is not available for this tab yet")
	}
	m.transitionTo(stateInstalling)
	flow := installui.New(installui.Config{
		Styles:    m.styles,
		Provider:  provider,
		AgentIDs:  m.agentIDs,
		CWD:       m.cwd,
		Home:      m.home,
		GetErrMsg: func() string { return m.errMsg },
		SetErrMsg: func(s string) { m.reportError(errors.New(s)) },
		ClearErr: func() {
			m.clearError()
			m.status = "ready"
		},
	})
	flow.SetSize(m.width, m.height)
	f.flow = &flow
	f.syncHint()
	return m, textinput.Blink
}

func (f *installFeature) syncHint() {
	if f.flow == nil {
		if f.backgroundActive() {
			f.m.setFooterContext("Installing " + f.bg.skillName + " in background")
		}
		return
	}
	if hint := f.flow.FooterHint(); hint != "" {
		f.m.setFooterContext(hint)
	}
}

func (f *installFeature) cancelFlow(hint string) {
	f.flow = nil
	f.m.transitionTo(stateListing)
	if hint != "" {
		f.m.setFooterContext(hint)
	}
}

func (f *installFeature) clearFlow() {
	f.flow = nil
}

func (f *installFeature) handleUIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installui.ClosedMsg:
		if f.flow == nil {
			return f.m, nil
		}
		f.cancelFlow(msg.Hint)
		return f.m, nil
	case installui.HintMsg:
		f.m.setFooterContext(msg.Text)
		return f.m, nil
	case installui.RequestInstallMsg:
		return f.startSelected(msg.AgentIDs, msg.Scope)
	case installui.InstallDoneMsg:
		return f.handleCompleted(installCompletedMsg{name: msg.Name, err: msg.Err})
	case installui.SearchDoneMsg:
		if f.flow == nil {
			return f.m, nil
		}
		next, cmd := f.flow.Update(msg)
		f.flow = &next
		if f.flow.Searching() {
			f.m.status = "loading"
		}
		f.syncHint()
		return f.m, cmd
	}
	if f.flow == nil {
		return f.m, nil
	}
	next, cmd := f.flow.Update(msg)
	f.flow = &next
	if f.flow.Searching() {
		f.m.status = "loading"
	}
	f.syncHint()
	return f.m, cmd
}

func (f *installFeature) startSelected(agentIDs []string, scope extension.Scope) (tea.Model, tea.Cmd) {
	flow := f.flow
	if flow == nil {
		return f.m, nil
	}
	candidate := flow.Selected()
	if candidate.Name == "" {
		return f.m, nil
	}
	provider, ok := f.providerForTab(f.m.activeTab)
	if !ok {
		return f.m, f.m.flashFooter("Install provider unavailable")
	}

	leftWidth, _, _, _ := f.m.paneSizes()
	f.bg = newInstallBackground(candidate.Name, leftWidth, f.m.styles)
	f.flow = nil

	f.m.transitionTo(stateListing)
	f.m.status = "loading"
	f.syncHint()

	cwd, home := f.m.cwd, f.m.home
	installCmd := func() tea.Msg {
		name, err := provider.Install(context.Background(), cwd, home, candidate, agentIDs, scope)
		return installui.InstallDoneMsg{Name: name, Err: err}
	}
	return f.m, tea.Batch(f.bg.begin(), installCmd)
}

func (f *installFeature) renderDialogArea() string {
	if f.flow == nil {
		return ""
	}
	leftWidth, mainHeight, _, _ := f.m.paneSizes()
	f.flow.SetSize(leftWidth, f.m.height)
	return f.flow.PlaceInPane(leftWidth, mainHeight)
}

func (f *installFeature) backgroundActive() bool {
	return f.bg != nil
}

func (f *installFeature) handleProgressTick() tea.Cmd {
	if f.bg == nil {
		return nil
	}
	return f.bg.handleTick()
}

func (f *installFeature) renderBackgroundOverlay(main string, mainHeight int) string {
	if f.bg == nil {
		return main
	}
	leftWidth, _, _, _ := f.m.paneSizesFor(mainHeight)
	corner := lipgloss.NewStyle().Width(leftWidth).PaddingLeft(1).Render(f.bg.view(f.m.styles))
	progressH := lipgloss.Height(corner)
	contentH := max(4, mainHeight-progressH)
	top := clipLines(main, contentH)
	return lipgloss.JoinVertical(lipgloss.Left, top, corner)
}

func (f *installFeature) handleCompleted(msg installCompletedMsg) (tea.Model, tea.Cmd) {
	f.bg = nil
	f.clearFlow()
	if f.m.state == stateInstalling {
		f.m.transitionTo(stateListing)
	}
	if msg.err != nil {
		f.m.reportError(msg.err)
		return f.m, f.m.beginScanAllCmd()
	}
	f.m.clearError()
	f.m.status = "ready"
	return f.m, tea.Batch(
		f.m.flashFooter("✓ Installed "+msg.name+" — back in skill list"),
		tea.Sequence(
			f.m.beginScanAllCmd(),
			func() tea.Msg { return reselectSkillMsg{name: msg.name} },
		),
	)
}

func (m *Model) startInstallFlow() (tea.Model, tea.Cmd) {
	return m.install.startFlow()
}

func (m *Model) handleInstallingUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	return m.install.handleUIMsg(msg)
}

func (m *Model) backgroundInstallActive() bool {
	return m.install.backgroundActive()
}

func (m *Model) renderInstallDialogArea() string {
	return m.install.renderDialogArea()
}

func (m *Model) renderBackgroundInstallOverlay(main string, mainHeight int) string {
	return m.install.renderBackgroundOverlay(main, mainHeight)
}

func (m *Model) cancelInstallFlow(hint string) {
	m.install.cancelFlow(hint)
}

func (m *Model) clearInstallFlow() {
	m.install.clearFlow()
}

func (m *Model) syncInstallHint() {
	m.install.syncHint()
}
