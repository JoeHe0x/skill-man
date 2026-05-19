package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const footerFlashLifetime = 3 * time.Second

type footerFlashTimeoutMsg struct {
	tag int
}

func (m *Model) setFooterContext(msg string) {
	m.footerContext = msg
}

func (m *Model) flashFooter(msg string) tea.Cmd {
	if msg == "" {
		return nil
	}
	m.footerFlash = msg
	m.footerFlashTag++
	tag := m.footerFlashTag
	return tea.Tick(footerFlashLifetime, func(time.Time) tea.Msg {
		return footerFlashTimeoutMsg{tag: tag}
	})
}

func (m *Model) handleFooterFlashTimeout(msg footerFlashTimeoutMsg) (tea.Model, tea.Cmd) {
	if msg.tag == m.footerFlashTag {
		m.footerFlash = ""
	}
	return m, nil
}
