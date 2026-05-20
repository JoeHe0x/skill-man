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
		return
	}
	if hint := m.install.flow.FooterHint(); hint != "" {
		m.setFooterContext(hint)
	}
}

func (m *Model) cancelInstallFlow(hint string) {
	m.transitionTo(m.lastState)
	m.abortInstallRun()
	m.install.flow = nil
	if hint != "" {
		m.setFooterContext(hint)
	}
}

func (m *Model) abortInstallRun() {
	if m.install.cancel != nil {
		m.install.cancel()
		m.install.cancel = nil
	}
}

func (m *Model) clearInstallFlow() {
	m.abortInstallRun()
	m.install.flow = nil
}

func (m *Model) handleInstallUIMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		m.transitionTo(m.lastState)
		return m, nil
	}
	switch msg := msg.(type) {
	case installui.ClosedMsg:
		m.cancelInstallFlow(msg.Hint)
		return m, nil
	case installui.HintMsg:
		m.setFooterContext(msg.Text)
		return m, nil
	case installui.CancelInstallMsg:
		if m.install.cancel != nil {
			m.install.cancel()
			m.install.cancel = nil
		}
		m.install.flow.EndInstall()
		m.status = "ready"
		m.syncInstallHint()
		m.setFooterContext("Cancelling install…")
		return m, nil
	case installui.RequestInstallMsg:
		return m.startInstallSelected(msg.AgentIDs)
	case installui.InstallDoneMsg:
		return m.handleInstallCompleted(installCompletedMsg{name: msg.Name, err: msg.Err})
	case installui.SearchDoneMsg, installui.ProgressTickMsg:
		next, cmd := m.install.flow.Update(msg)
		m.install.flow = &next
		if m.install.flow.Searching() {
			m.status = "loading"
		}
		m.syncInstallHint()
		return m, cmd
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
		m.transitionTo(m.lastState)
		return m, nil
	}
	return m.handleInstallUIMsg(msg)
}

func (m *Model) startInstallSelected(agentIDs []string) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	if m.install.cancel != nil {
		m.install.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.install.cancel = cancel
	begin := m.install.flow.BeginInstall()
	installCmd := m.install.flow.InstallCmd(ctx, agentIDs)
	m.status = "loading"
	m.syncInstallHint()
	return m, tea.Batch(begin, installCmd)
}

func (m *Model) renderInstallDialogArea() string {
	if m.install.flow == nil {
		return ""
	}
	leftWidth, mainHeight, _, _ := m.paneSizes()
	m.install.flow.SetSize(leftWidth, m.height)
	return m.install.flow.PlaceInPane(leftWidth, mainHeight)
}
