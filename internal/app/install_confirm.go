package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) handleInstallConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.install.flow.step = installStepAgents
		m.syncInstallHint()
		return m, nil
	case key.Matches(msg, keys.Enter):
		agentIDs := selectedInstallAgentIDs(m.install.flow.targets)
		if len(agentIDs) == 0 {
			m.install.flow.step = installStepAgents
			m.setFooterContext("Select at least one install path (Space to toggle)")
			return m, nil
		}
		return m.startInstallSelected(agentIDs)
	}
	return m, nil
}

func installStepBreadcrumb(step installStep) string {
	const (
		sBrowse  = "1 Search"
		sPick    = "2 Pick"
		sPaths   = "3 Paths"
		sConfirm = "4 Confirm"
	)
	switch step {
	case installStepAgents:
		return lipgloss.JoinHorizontal(lipgloss.Left,
			dimStep(sBrowse), sep(), activeStep(sPick), sep(), dimStep(sPaths), sep(), dimStep(sConfirm),
		)
	case installStepConfirm:
		return lipgloss.JoinHorizontal(lipgloss.Left,
			dimStep(sBrowse), sep(), dimStep(sPick), sep(), dimStep(sPaths), sep(), activeStep(sConfirm),
		)
	default:
		return lipgloss.JoinHorizontal(lipgloss.Left,
			activeStep(sBrowse), sep(), dimStep(sPick), sep(), dimStep(sPaths), sep(), dimStep(sConfirm),
		)
	}
}

func sep() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" → ")
}

func activeStep(label string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(label)
}

func dimStep(label string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(label)
}

func (m *Model) renderInstallConfirm(innerWidth int) string {
	flow := m.install.flow
	agentIDs := selectedInstallAgentIDs(flow.targets)
	var paths []string
	for _, t := range flow.targets {
		if t.desired {
			paths = append(paths, t.skillDir)
		}
	}

	lines := []string{
		m.styles.PanelTitle.Render("Confirm install"),
		m.styles.Hint.Render(installStepBreadcrumb(installStepConfirm)),
		"",
		m.styles.HintBold.Render("Skill: ") + flow.selected.Name,
		m.styles.Hint.Render("Source: " + truncate(flow.selected.Source, innerWidth-8)),
		m.styles.Hint.Render(fmt.Sprintf("Agents (%d): %s", len(agentIDs), strings.Join(agentIDs, ", "))),
		m.styles.Hint.Render("Paths:"),
	}
	for _, p := range paths {
		lines = append(lines, m.styles.Hint.Render("  • "+truncate(p, innerWidth-4)))
	}
	if flow.selected.Local {
		lines = append(lines, "", m.styles.Hint.Render("Local skill — will copy/link into selected agent paths."))
	} else {
		lines = append(lines, "", m.styles.Hint.Render("Registry install via skills CLI (npx skills add)."))
	}
	lines = append(lines, "", m.styles.Hint.Render("Enter to install · Esc to go back"))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
