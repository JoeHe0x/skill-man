package app

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
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
	installStepConfirm // summary before download
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
	kind          domaininstall.Kind
	provider      serviceinstall.Provider
	step          installStep
	focus         installFocus
	query         string
	results       []domaininstall.Candidate
	selected      domaininstall.Candidate
	targets       []installDirChoice
	searching     bool
	installing    bool
	quitPending   bool
	progress      progress.Model
	recentQueries []string
	searchInput   textinput.Model
	resultList    list.Model
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

	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(36))

	flow := &installFlow{
		kind:        provider.Kind(),
		provider:    provider,
		step:        installStepBrowse,
		focus:       installFocusSearch,
		progress:    bar,
		searchInput: ti,
		resultList:  resultList,
	}
	configureInstallSearchInput(flow)
	return flow
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
	m.install.flow = newInstallFlow(provider, m.listDelegate)
	m.syncInstallHint()
	return m, textinput.Blink
}

// syncInstallHint updates the footer hint for the install dialog.
func (m *Model) syncInstallHint() {
	if m.install.flow == nil {
		return
	}
	if m.errMsg != "" && m.install.flow.step == installStepBrowse && len(m.install.flow.results) == 0 {
		return // renderHintFooter shows errMsg
	}
	if m.install.flow.installing {
		m.setFooterContext(fmt.Sprintf("Installing %s…  Esc twice to cancel", m.install.flow.selected.Name))
		return
	}
	switch m.install.flow.step {
	case installStepConfirm:
		m.setFooterContext(fmt.Sprintf("Confirm install %s | Enter: install | Esc: back to paths",
			m.install.flow.selected.Name))
	case installStepAgents:
		m.setFooterContext(fmt.Sprintf("Install %s | Space: toggle path | Enter: review | Esc: back to results",
			m.install.flow.selected.Name))
	default:
		if m.install.flow.searching {
			m.setFooterContext(fmt.Sprintf("Searching skills.sh for %q…  Esc: cancel", m.install.flow.query))
			return
		}
		if len(m.install.flow.results) > 0 {
			m.setFooterContext(fmt.Sprintf("Search %q → %d results | ↑↓: select skill | Enter: choose paths | /: new search | Esc: close",
				m.install.flow.query, len(m.install.flow.results)))
			return
		}
		m.setFooterContext("Search & Install | Tab: complete · ↑↓: suggestion or list · Enter: search · Esc: close")
	}
}

func (m *Model) cancelInstallFlow(hint string) {
	m.transitionTo(m.lastState)
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
	m.install.flow = nil
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
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: installDirTitle(c.skillDir, c.desired),
			Desc:  formatInstallDirAgents(c.agents),
			Meta:  c.skillDir,
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
		items = append(items, panel.Item{
			Kind:  panel.ItemMessage,
			Title: c.Name,
			Desc:  c.Source,
			Meta:  meta,
		})
	}
	return items
}

func (m *Model) handleInstallingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		m.transitionTo(m.lastState)
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleInstallingKeys(msg)
	case installSearchCompletedMsg:
		return m.handleInstallSearchCompleted(msg)
	}

	if m.install.flow.step == installStepBrowse && m.install.flow.focus == installFocusSearch {
		var cmd tea.Cmd
		m.install.flow.searchInput, cmd = m.install.flow.searchInput.Update(msg)
		return m, cmd
	}
	if m.install.flow.step == installStepBrowse && m.install.flow.focus == installFocusList {
		var cmd tea.Cmd
		m.install.flow.resultList, cmd = m.install.flow.resultList.Update(msg)
		return m, cmd
	}
	if m.install.flow.step == installStepAgents {
		var cmd tea.Cmd
		m.agentList, cmd = m.agentList.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) handleInstallSearchCompleted(msg installSearchCompletedMsg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	m.install.flow.searching = false
	if msg.err != nil {
		m.reportError(msg.err)
		m.status = "ready"
		m.install.flow.focus = installFocusSearch
		m.install.flow.results = nil
		m.install.flow.resultList.SetItems(nil)
		m.syncInstallHint()
		return m, nil
	}
	m.clearError()
	m.status = "ready"
	m.install.flow.results = msg.results
	m.install.flow.syncSearchSuggestions()
	m.install.flow.focus = installFocusList
	items := installResultsToListItems(msg.results)
	m.listDelegate.SetHeight(listHeightForItems(items))
	m.install.flow.resultList.SetItems(items)
	if len(items) > 0 {
		m.install.flow.resultList.Select(0)
	}
	m.syncInstallHint()
	return m, nil
}

func (m *Model) runInstallSearch(query string) tea.Cmd {
	m.install.flow.rememberSearchQuery(query)
	m.install.flow.query = query
	m.install.flow.searching = true
	m.status = "loading"
	m.syncInstallHint()
	provider := m.install.flow.provider
	return func() tea.Msg {
		results, err := provider.Search(query)
		return installSearchCompletedMsg{results: results, err: err}
	}
}

func (m *Model) handleInstallingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	if m.install.flow.installing {
		return m.handleInstallRunningKeys(msg)
	}

	switch m.install.flow.step {
	case installStepBrowse:
		return m.handleInstallBrowseKeys(msg)
	case installStepAgents:
		return m.handleInstallAgentsKeys(msg)
	case installStepConfirm:
		return m.handleInstallConfirmKeys(msg)
	}
	return m, nil
}

func (m *Model) handleInstallBrowseKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.install.flow.focus == installFocusSearch || len(m.install.flow.results) == 0 {
			query := strings.TrimSpace(m.install.flow.searchInput.Value())
			if query == "" {
				m.setFooterContext("Enter a search keyword, then press Enter")
				return m, nil
			}
			return m, m.runInstallSearch(query)
		}
		idx := m.install.flow.resultList.Index()
		if idx < 0 || idx >= len(m.install.flow.results) {
			m.setFooterContext("Select a skill from the list")
			return m, nil
		}
		m.install.flow.selected = m.install.flow.results[idx]
		m.install.flow.step = installStepAgents
		m.install.flow.targets = newInstallDirChoices(m.agentIDs)
		m.setAgentListItems(installDirChoicesToListItems(m.install.flow.targets))
		m.agentList.Select(0)
		m.syncInstallHint()
		return m, nil

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		m.cancelInstallFlow("Install cancelled")
		return m, nil

	case msg.String() == "/" || key.Matches(msg, keys.Find):
		m.install.flow.focus = installFocusSearch
		m.install.flow.searchInput.Focus()
		m.syncInstallHint()
		return m, textinput.Blink

	case key.Matches(msg, keys.Up, keys.Down):
		if m.install.flow.focus == installFocusSearch && len(m.install.flow.results) == 0 {
			var cmd tea.Cmd
			m.install.flow.searchInput, cmd = m.install.flow.searchInput.Update(msg)
			return m, cmd
		}
		if len(m.install.flow.results) == 0 {
			return m, nil
		}
		m.install.flow.focus = installFocusList
		var cmd tea.Cmd
		m.install.flow.resultList, cmd = m.install.flow.resultList.Update(msg)
		return m, cmd
	}

	if m.install.flow.focus == installFocusSearch {
		var cmd tea.Cmd
		m.install.flow.searchInput, cmd = m.install.flow.searchInput.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.install.flow.resultList, cmd = m.install.flow.resultList.Update(msg)
	return m, cmd
}

func (m *Model) handleInstallAgentsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.install.flow.step = installStepBrowse
		m.install.flow.focus = installFocusList
		m.syncInstallHint()
		return m, nil

	case key.Matches(msg, keys.Enter):
		agentIDs := selectedInstallAgentIDs(m.install.flow.targets)
		if len(agentIDs) == 0 {
			m.setFooterContext("Select at least one install path (Space to toggle)")
			return m, nil
		}
		m.install.flow.step = installStepConfirm
		m.syncInstallHint()
		return m, nil

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		if idx >= 0 && idx < len(m.install.flow.targets) {
			m.install.flow.targets[idx].desired = !m.install.flow.targets[idx].desired
			m.setAgentListItems(installDirChoicesToListItems(m.install.flow.targets))
			m.agentList.Select(idx)
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) startInstallSelected(agentIDs []string) (tea.Model, tea.Cmd) {
	if m.install.flow == nil {
		return m, nil
	}
	candidate := m.install.flow.selected
	if candidate.Source == "" {
		return m, nil
	}
	cwd := m.cwd
	home := m.home
	provider := m.install.flow.provider

	if m.install.cancel != nil {
		m.install.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.install.cancel = cancel

	m.install.flow.installing = true
	m.install.flow.quitPending = false
	m.install.flow.progress = progress.New(progress.WithDefaultGradient(), progress.WithWidth(36))
	m.install.flow.progress.ShowPercentage = true
	m.status = "loading"
	m.syncInstallHint()

	start := m.install.flow.progress.SetPercent(0)
	tick := installProgressTickCmd()

	return m, tea.Batch(start, tick, func() tea.Msg {
		name, err := provider.Install(ctx, cwd, home, candidate, agentIDs)
		return installCompletedMsg{name: name, err: err}
	})
}

func installProgressTickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg {
		return installProgressTickMsg{}
	})
}

func (m *Model) handleInstallProgressTick() (tea.Model, tea.Cmd) {
	if m.install.flow == nil || !m.install.flow.installing {
		return m, nil
	}
	var cmds []tea.Cmd
	if m.install.flow.progress.Percent() < 0.9 {
		cmds = append(cmds, m.install.flow.progress.IncrPercent(0.04))
	}
	cmds = append(cmds, installProgressTickCmd())
	return m, tea.Batch(cmds...)
}

func (m *Model) handleInstallRunningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		return m.handleInstallQuitAttempt()
	}
	return m, nil
}

func (m *Model) handleInstallQuitAttempt() (tea.Model, tea.Cmd) {
	if m.install.flow == nil || !m.install.flow.installing {
		return m, nil
	}
	if !m.install.flow.quitPending {
		m.install.flow.quitPending = true
		m.setFooterContext("Install in progress — press Esc again to cancel")
		return m, nil
	}
	if m.install.cancel != nil {
		m.install.cancel()
		m.install.cancel = nil
	}
	m.install.flow.installing = false
	m.install.flow.quitPending = false
	m.status = "ready"
	m.syncInstallHint()
	m.setFooterContext("Cancelling install…")
	return m, nil
}

func (m *Model) renderInstallDialog() string {
	flow := m.install.flow
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
	if flow.installing {
		flow.progress.Width = max(20, innerWidth-2)
		body = lipgloss.JoinVertical(lipgloss.Left,
			m.styles.PanelTitle.Render("Installing"),
			m.styles.Hint.Render(fmt.Sprintf("%s  (%s)", flow.selected.Name, flow.selected.Source)),
			flow.progress.View(),
			m.styles.Hint.Render("Running skills CLI — progress is approximate"),
		)
	} else {
		crumbs := m.styles.Hint.Render(installStepBreadcrumb(flow.step))
		switch flow.step {
		case installStepConfirm:
			body = lipgloss.JoinVertical(lipgloss.Left, crumbs, m.renderInstallConfirm(innerWidth))
		case installStepAgents:
			m.agentList.SetSize(innerWidth, listHeight)
			body = lipgloss.JoinVertical(lipgloss.Left,
				crumbs,
				m.styles.PanelTitle.Render("Install paths"),
				m.styles.Hint.Render(fmt.Sprintf("Skill: %s  (%s)", flow.selected.Name, flow.selected.Source)),
				m.styles.Hint.Render("Each row is a skills directory; agents listed share that path"),
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
					listView = m.styles.StatusError.Render("  " + truncate(m.errMsg, innerWidth-2))
				} else {
					listView = m.styles.EmptyPreview.Render("  Type a keyword and press Enter to search skills.sh")
				}
			}
			body = lipgloss.JoinVertical(lipgloss.Left,
				crumbs,
				m.styles.PanelTitle.Render("Search & Install"),
				searchLine,
				listView,
			)
		}
	}

	box := m.styles.Modal.Width(dialogWidth).Render(body)
	return box
}

// renderInstallDialogArea places the install dialog in the left main pane.
func (m *Model) renderInstallDialogArea() string {
	leftWidth, mainHeight, _, _ := m.paneSizes()
	return lipgloss.Place(leftWidth, mainHeight, lipgloss.Left, lipgloss.Top, m.renderInstallDialog())
}
