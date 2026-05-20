package app

import (
	"context"
	"errors"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/installui"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

func (m *Model) installProviderForTab(tab panel.Tab) (serviceinstall.Provider, bool) {
	switch tab {
	case panel.TabSkills:
		return serviceinstall.NewSkillsCLIProvider(), true
	default:
		return nil, false
	}
}

func (m *Model) startInstallFlow() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().SearchInstall {
		return m, m.flashFooter("Search & install is not available for this tab yet")
	}
	provider, ok := m.installProviderForTab(m.activeTab)
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
	m.install.flow = &flow
	m.syncInstallHint()
	return m, textinput.Blink
}

func (m *Model) syncInstallHint() {
	if m.install.flow == nil {
		if m.backgroundInstallActive() {
			m.setFooterContext("Installing " + m.install.bg.skillName + " in background")
		}
		return
	}
	if hint := m.install.flow.FooterHint(); hint != "" {
		m.setFooterContext(hint)
	}
}

func (m *Model) cancelInstallFlow(hint string) {
	m.install.flow = nil
	m.transitionTo(stateListing)
	if hint != "" {
		m.setFooterContext(hint)
	}
}

func (m *Model) clearInstallFlow() {
	m.install.flow = nil
}

func (m *Model) handleInstallUIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installui.ClosedMsg:
		if m.install.flow == nil {
			return m, nil
		}
		m.cancelInstallFlow(msg.Hint)
		return m, nil
	case installui.HintMsg:
		m.setFooterContext(msg.Text)
		return m, nil
	case installui.RequestInstallMsg:
		return m.startInstallSelected(msg.AgentIDs)
	case installui.InstallDoneMsg:
		return m.handleInstallCompleted(installCompletedMsg{name: msg.Name, err: msg.Err})
	case installui.SearchDoneMsg:
		if m.install.flow == nil {
			return m, nil
		}
		next, cmd := m.install.flow.Update(msg)
		m.install.flow = &next
		if m.install.flow.Searching() {
			m.status = "loading"
		}
		m.syncInstallHint()
		return m, cmd
	}
	if m.install.flow == nil {
		return m, nil
	}
	next, cmd := m.install.flow.Update(msg)
	m.install.flow = &next
	if m.install.flow.Searching() {
		m.status = "loading"
	}
	m.syncInstallHint()
	return m, cmd
}

func (m *Model) handleInstallingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	return m.handleInstallUIMsg(msg)
}

func (m *Model) startInstallSelected(agentIDs []string) (tea.Model, tea.Cmd) {
	flow := m.install.flow
	if flow == nil {
		return m, nil
	}
	candidate := flow.Selected()
	if candidate.Name == "" {
		return m, nil
	}
	provider, ok := m.installProviderForTab(m.activeTab)
	if !ok {
		return m, m.flashFooter("Install provider unavailable")
	}

	leftWidth, _, _, _ := m.paneSizes()
	m.install.bg = newInstallBackground(candidate.Name, leftWidth, m.styles)
	m.install.flow = nil

	m.transitionTo(stateListing)
	m.status = "loading"
	m.syncInstallHint()

	cwd, home := m.cwd, m.home
	installCmd := func() tea.Msg {
		name, err := provider.Install(context.Background(), cwd, home, candidate, agentIDs)
		return installui.InstallDoneMsg{Name: name, Err: err}
	}
	return m, tea.Batch(m.install.bg.begin(), installCmd)
}

func (m *Model) renderInstallDialogArea() string {
	if m.install.flow == nil {
		return ""
	}
	leftWidth, mainHeight, _, _ := m.paneSizes()
	m.install.flow.SetSize(leftWidth, m.height)
	return m.install.flow.PlaceInPane(leftWidth, mainHeight)
}
