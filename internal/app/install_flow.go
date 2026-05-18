package app

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

type installStep int

const (
	installStepBrowse installStep = iota // search input + registry results inside dialog
	installStepAgents
)

type installFocus int

const (
	installFocusSearch installFocus = iota
	installFocusList
)

// installDirChoice is one install destination (skill dir) and the agents that use it.
type installDirChoice struct {
	skillDir string
	agents   []agent.Agent
	desired  bool
}

// installFlow is the generic search-and-install dialog (skills today; MCP later).
type installFlow struct {
	kind        domaininstall.Kind
	provider    serviceinstall.Provider
	step        installStep
	focus       installFocus
	query       string
	results     []domaininstall.Candidate
	selected    domaininstall.Candidate
	targets     []installDirChoice
	searching   bool
	searchInput textinput.Model
	resultList  list.Model
}

func (m *Model) installProviderForTab(tab panel.Tab) (serviceinstall.Provider, bool) {
	switch tab {
	case panel.TabSkills:
		return serviceinstall.NewSkillsCLIProvider(), true
	default:
		return nil, false
	}
}

func newInstallFlow(provider serviceinstall.Provider, delegate *itemDelegate) *installFlow {
	ti := textinput.New()
	ti.Placeholder = "search skills.sh (e.g. react, testing)..."
	ti.CharLimit = 128
	ti.Prompt = "🔍 "
	ti.Focus()

	resultList := list.New([]list.Item{}, delegate, 0, 0)
	resultList.Title = ""
	resultList.SetShowTitle(false)
	resultList.SetShowStatusBar(false)
	resultList.SetFilteringEnabled(false)
	resultList.SetShowHelp(false)
	resultList.DisableQuitKeybindings()
	resultList.KeyMap.CursorUp = keys.Up
	resultList.KeyMap.CursorDown = keys.Down
	resultList.KeyMap.NextPage = keys.PgDown
	resultList.KeyMap.PrevPage = keys.PgUp

	return &installFlow{
		kind:        provider.Kind(),
		provider:    provider,
		step:        installStepBrowse,
		focus:       installFocusSearch,
		searchInput: ti,
		resultList:  resultList,
	}
}

func (m *Model) startInstallFlow() (tea.Model, tea.Cmd) {
	if !m.activePanel().Capabilities().SearchInstall {
		m.hint = "Search & install is not available for this tab yet"
		return m, nil
	}
	provider, ok := m.installProviderForTab(m.activeTab)
	if !ok {
		m.hint = "Search & install is not available for this tab yet"
		return m, nil
	}

	m.lastState = m.state
	m.state = stateInstalling
	m.installFlow = newInstallFlow(provider, m.listDelegate)
	m.syncInstallHint()
	return m, textinput.Blink
}

// syncInstallHint updates the footer hint for the install dialog.
func (m *Model) syncInstallHint() {
	if m.installFlow == nil {
		return
	}
	if m.errMsg != "" && m.installFlow.step == installStepBrowse && len(m.installFlow.results) == 0 {
		return // renderHintFooter shows errMsg
	}
	switch m.installFlow.step {
	case installStepAgents:
		m.hint = fmt.Sprintf("Install %s | Space: toggle install path | Enter: confirm | Esc: back to results",
			m.installFlow.selected.Name)
	default:
		if m.installFlow.searching {
			m.hint = fmt.Sprintf("Searching skills.sh for %q…  Esc: cancel", m.installFlow.query)
			return
		}
		if len(m.installFlow.results) > 0 {
			m.hint = fmt.Sprintf("Search %q → %d results | ↑↓: select skill | Enter: choose paths | /: new search | Esc: close",
				m.installFlow.query, len(m.installFlow.results))
			return
		}
		m.hint = "Search & Install | Type keyword · Enter: search skills.sh · Esc: close"
	}
}

func (m *Model) cancelInstallFlow(hint string) {
	m.installFlow = nil
	m.state = m.lastState
	if hint != "" {
		m.hint = hint
	}
}

func (m *Model) clearInstallFlow() {
	m.installFlow = nil
}

// newInstallDirChoices groups agents by skill directory (actual install path).
// Default selection follows the header agent filter (Ctrl+A dialog).
func newInstallDirChoices(agentFilter []string) []installDirChoice {
	byDir := map[string][]agent.Agent{}
	for _, a := range agent.DefaultAgents() {
		dir := a.EntityDirs[agent.EntitySkill]
		if dir == "" {
			continue
		}
		byDir[dir] = append(byDir[dir], a)
	}

	dirs := make([]string, 0, len(byDir))
	for dir := range byDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	wantAll := len(agentFilter) == 0 || slices.Contains(agentFilter, "all")
	choices := make([]installDirChoice, 0, len(dirs))
	for _, dir := range dirs {
		agents := byDir[dir]
		slices.SortFunc(agents, func(a, b agent.Agent) int {
			return strings.Compare(a.Name, b.Name)
		})
		desired := false
		if !wantAll {
			for _, id := range agentFilter {
				if slices.ContainsFunc(agents, func(a agent.Agent) bool { return a.ID == id }) {
					desired = true
					break
				}
			}
		}
		choices = append(choices, installDirChoice{
			skillDir: dir,
			agents:   agents,
			desired:  desired,
		})
	}
	return choices
}

func installDirChoicesToListItems(choices []installDirChoice) []list.Item {
	items := make([]list.Item, 0, len(choices))
	for _, c := range choices {
		items = append(items, listItem{
			kind:  itemKindMessage,
			title: installDirTitle(c.skillDir, c.desired),
			desc:  formatInstallDirAgents(c.agents),
			meta:  c.skillDir,
		})
	}
	return items
}

func installDirTitle(skillDir string, checked bool) string {
	if checked {
		return "✓ " + skillDir
	}
	return "  " + skillDir
}

func formatInstallDirAgents(agents []agent.Agent) string {
	if len(agents) == 0 {
		return ""
	}
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	if len(names) <= 5 {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:5], ", ") + fmt.Sprintf(" +%d more", len(names)-5)
}

func selectedInstallAgentIDs(targets []installDirChoice) []string {
	seen := map[string]bool{}
	var ids []string
	for _, t := range targets {
		if !t.desired {
			continue
		}
		for _, a := range t.agents {
			if seen[a.ID] {
				continue
			}
			seen[a.ID] = true
			ids = append(ids, a.ID)
		}
	}
	return ids
}

func installResultsToListItems(results []domaininstall.Candidate) []list.Item {
	items := make([]list.Item, 0, len(results))
	for _, c := range results {
		meta := c.Installs
		if c.Local {
			meta = "local"
		}
		items = append(items, listItem{
			kind:  itemKindMessage,
			title: c.Name,
			desc:  c.Source,
			meta:  meta,
		})
	}
	return items
}

func (m *Model) handleInstallingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.installFlow == nil {
		m.state = m.lastState
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleInstallingKeys(msg)
	case installSearchCompletedMsg:
		return m.handleInstallSearchCompleted(msg)
	}

	if m.installFlow.step == installStepBrowse && m.installFlow.focus == installFocusSearch {
		var cmd tea.Cmd
		m.installFlow.searchInput, cmd = m.installFlow.searchInput.Update(msg)
		return m, cmd
	}
	if m.installFlow.step == installStepBrowse && m.installFlow.focus == installFocusList {
		var cmd tea.Cmd
		m.installFlow.resultList, cmd = m.installFlow.resultList.Update(msg)
		return m, cmd
	}
	if m.installFlow.step == installStepAgents {
		var cmd tea.Cmd
		m.agentList, cmd = m.agentList.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) handleInstallSearchCompleted(msg installSearchCompletedMsg) (tea.Model, tea.Cmd) {
	if m.installFlow == nil {
		return m, nil
	}
	m.installFlow.searching = false
	if msg.err != nil {
		m.reportError(msg.err)
		m.status = "ready"
		m.installFlow.focus = installFocusSearch
		m.installFlow.results = nil
		m.installFlow.resultList.SetItems(nil)
		m.syncInstallHint()
		return m, nil
	}
	m.clearError()
	m.status = "ready"
	m.installFlow.results = msg.results
	m.installFlow.focus = installFocusList
	items := installResultsToListItems(msg.results)
	m.listDelegate.SetHeight(listHeightForItems(items))
	m.installFlow.resultList.SetItems(items)
	if len(items) > 0 {
		m.installFlow.resultList.Select(0)
	}
	m.syncInstallHint()
	return m, nil
}

func (m *Model) runInstallSearch(query string) tea.Cmd {
	m.installFlow.query = query
	m.installFlow.searching = true
	m.status = "loading"
	m.syncInstallHint()
	provider := m.installFlow.provider
	return func() tea.Msg {
		results, err := provider.Search(query)
		return installSearchCompletedMsg{results: results, err: err}
	}
}

func (m *Model) handleInstallingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.installFlow == nil {
		return m, nil
	}

	switch m.installFlow.step {
	case installStepBrowse:
		return m.handleInstallBrowseKeys(msg)
	case installStepAgents:
		return m.handleInstallAgentsKeys(msg)
	}
	return m, nil
}

func (m *Model) handleInstallBrowseKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.installFlow.focus == installFocusSearch || len(m.installFlow.results) == 0 {
			query := strings.TrimSpace(m.installFlow.searchInput.Value())
			if query == "" {
				m.hint = "Enter a search keyword, then press Enter"
				return m, nil
			}
			return m, m.runInstallSearch(query)
		}
		idx := m.installFlow.resultList.Index()
		if idx < 0 || idx >= len(m.installFlow.results) {
			m.hint = "Select a skill from the list"
			return m, nil
		}
		m.installFlow.selected = m.installFlow.results[idx]
		m.installFlow.step = installStepAgents
		m.installFlow.targets = newInstallDirChoices(m.agentIDs)
		m.agentList.SetItems(installDirChoicesToListItems(m.installFlow.targets))
		m.agentList.Select(0)
		m.syncInstallHint()
		return m, nil

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		m.cancelInstallFlow("Install cancelled")
		return m, nil

	case msg.String() == "/" || key.Matches(msg, keys.Find):
		m.installFlow.focus = installFocusSearch
		m.installFlow.searchInput.Focus()
		m.syncInstallHint()
		return m, textinput.Blink

	case key.Matches(msg, keys.Up, keys.Down):
		if len(m.installFlow.results) == 0 {
			return m, nil
		}
		m.installFlow.focus = installFocusList
		var cmd tea.Cmd
		m.installFlow.resultList, cmd = m.installFlow.resultList.Update(msg)
		return m, cmd
	}

	if m.installFlow.focus == installFocusSearch {
		var cmd tea.Cmd
		m.installFlow.searchInput, cmd = m.installFlow.searchInput.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.installFlow.resultList, cmd = m.installFlow.resultList.Update(msg)
	return m, cmd
}

func (m *Model) handleInstallAgentsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.installFlow.step = installStepBrowse
		m.installFlow.focus = installFocusList
		m.syncInstallHint()
		return m, nil

	case key.Matches(msg, keys.Enter):
		agentIDs := selectedInstallAgentIDs(m.installFlow.targets)
		if len(agentIDs) == 0 {
			m.hint = "Select at least one install path (Space to toggle)"
			return m, nil
		}
		return m.startInstallSelected(agentIDs)

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		if idx >= 0 && idx < len(m.installFlow.targets) {
			m.installFlow.targets[idx].desired = !m.installFlow.targets[idx].desired
			cmd := m.agentList.SetItems(installDirChoicesToListItems(m.installFlow.targets))
			m.agentList.Select(idx)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) startInstallSelected(agentIDs []string) (tea.Model, tea.Cmd) {
	if m.installFlow == nil {
		return m, nil
	}
	candidate := m.installFlow.selected
	if candidate.Source == "" {
		return m, nil
	}
	cwd := m.cwd
	home := m.home
	provider := m.installFlow.provider
	m.status = "loading"
	m.hint = fmt.Sprintf("Installing %s...", candidate.Name)

	return m, func() tea.Msg {
		name, err := provider.Install(cwd, home, candidate, agentIDs)
		return installCompletedMsg{name: name, err: err}
	}
}

func (m *Model) renderInstallDialog() string {
	flow := m.installFlow
	if flow == nil {
		return ""
	}

	leftWidth, _, _, _ := m.paneSizes()
	dialogWidth := min(max(44, leftWidth-4), 76)
	if dialogWidth > leftWidth-2 {
		dialogWidth = max(20, leftWidth-2)
	}
	dialogHeight := min(max(16, m.height-8), 28)
	innerWidth := dialogWidth - 4
	listHeight := dialogHeight - 10
	if listHeight < 4 {
		listHeight = 4
	}

	var body string
	switch flow.step {
	case installStepAgents:
		m.agentList.SetSize(innerWidth, listHeight)
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.styles.panelTitle.Render("Install paths"),
			m.styles.hint.Render(fmt.Sprintf("Skill: %s  (%s)", flow.selected.Name, flow.selected.Source)),
			m.styles.hint.Render("Each row is a skills directory; agents listed share that path"),
			m.agentList.View(),
		)
	default:
		flow.resultList.SetSize(innerWidth, listHeight)
		searchLine := flow.searchInput.View()
		if flow.searching {
			searchLine += "  " + m.spinner.View()
		}
		listView := flow.resultList.View()
		if len(flow.results) == 0 && !flow.searching {
			if m.errMsg != "" {
				listView = m.styles.statusError.Render("  " + truncate(m.errMsg, innerWidth-2))
			} else {
				listView = m.styles.emptyPreview.Render("  Type a keyword and press Enter to search skills.sh")
			}
		}
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.styles.panelTitle.Render("Search & Install"),
			searchLine,
			listView,
		)
	}

	box := m.styles.modal.Width(dialogWidth).Render(body)
	return box
}

// renderInstallDialogArea places the install dialog in the left main pane.
func (m *Model) renderInstallDialogArea() string {
	leftWidth, mainHeight, _, _ := m.paneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, m.renderInstallDialog())
}
