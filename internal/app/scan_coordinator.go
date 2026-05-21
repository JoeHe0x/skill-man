package app

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
)

// scanCoordinator tracks in-flight panel scans for a single rescan batch.
type scanCoordinator struct {
	Gen     uint64
	Pending int
}

func (c *scanCoordinator) begin(tabCount int) uint64 {
	c.Gen++
	c.Pending = tabCount
	if c.Pending == 0 {
		c.Pending = 1
	}
	return c.Gen
}

func (c *scanCoordinator) tag(gen uint64, msg tea.Msg) tea.Msg {
	if m, ok := msg.(panel.ScannedMsg); ok {
		m.Gen = gen
		return m
	}
	return msg
}

func (c *scanCoordinator) stale(gen uint64) bool {
	return gen != c.Gen || c.Pending <= 0
}

func (c *scanCoordinator) noteComplete(m *Model, gen uint64) tea.Cmd {
	if c.stale(gen) {
		return nil
	}
	c.Pending--
	if c.Pending > 0 {
		return nil
	}
	m.status = "ready"
	m.previewGen++
	m.updateFooterForState(m.state)
	if m.state == stateHome || m.state == stateListing || m.state == stateSearching {
		m.refreshActiveList()
		return m.syncSelectionPreview()
	}
	return nil
}

func (m *Model) beginScanAllCmd() tea.Cmd {
	gen := m.scan.begin(len(m.panels.Tabs()))
	m.status = "loading"
	m.updateFooterForState(m.state)

	cmds := make([]tea.Cmd, 0, len(m.panels.Tabs()))
	for _, tab := range m.panels.Tabs() {
		tab := tab
		scan := panel.ScanCmd(m.panels.Get(tab), m.cwd, m.home, slices.Clone(m.allAgents))
		cmds = append(cmds, func() tea.Msg {
			return m.scan.tag(gen, scan())
		})
	}
	return tea.Batch(cmds...)
}

func (m *Model) handleScanned(msg panel.ScannedMsg) (tea.Model, tea.Cmd) {
	if m.scan.stale(msg.Gen) {
		return m, nil
	}
	if msg.Err != nil {
		m.reportError(msg.Err)
		return m, m.scan.noteComplete(m, msg.Gen)
	}

	applied := m.panels.Get(msg.Tab).ApplyScan(msg)
	if msg.Tab == panel.TabMCP && !applied {
		return m, m.scan.noteComplete(m, msg.Gen)
	}

	m.clearError()
	if msg.Tab == panel.TabSkills && m.state == stateInstalling && m.install.flow != nil {
		return m, m.scan.noteComplete(m, msg.Gen)
	}

	var cmds []tea.Cmd
	if m.activeTab == msg.Tab && (m.state == stateHome || m.state == stateListing || m.state == stateSearching) {
		m.refreshActiveList()
		cmds = append(cmds, m.syncSelectionPreview())
	}
	cmds = append(cmds, m.scan.noteComplete(m, msg.Gen))
	return m, tea.Batch(cmds...)
}

// noteScanCompleted is used by tests to drive the coordinator without a full scan message.
func (m *Model) noteScanCompleted(gen uint64) tea.Cmd {
	return m.scan.noteComplete(m, gen)
}
