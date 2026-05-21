package installui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/domain/extension"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

type step int

const (
	stepBrowse step = iota
	stepPaths
)

type focus int

const (
	focusSearch focus = iota
	focusResults
)

// Config wires host dependencies into the install wizard.
type Config struct {
	Styles    theme.Styles
	Provider  serviceinstall.Provider
	AgentIDs  []string
	CWD       string
	Home      string
	GetErrMsg func() string
	SetErrMsg func(string)
	ClearErr  func()
}

// Model is a composable Bubble Tea sub-model for search & install (skills.sh).
type Model struct {
	cfg Config

	step     step
	focus    focus
	query    string
	results  []domaininstall.Candidate
	selected domaininstall.Candidate
	scope    extension.Scope
	targets  []dirChoice

	searching bool

	width  int
	height int

	recentQueries []string
	searchInput   textinput.Model
	resultList    list.Model
	pathsList     list.Model
	spinner       spinner.Model
	delegate      *itemDelegate
}

// New builds an install wizard focused on the search step.
func New(cfg Config) Model {
	ti := textinput.New()
	ti.Placeholder = "skills.sh"
	ti.CharLimit = 128
	ti.Prompt = "> "
	ti.Focus()

	delegate := newItemDelegate(cfg.Styles)
	resultList := newResultsList(delegate)
	pathsList := newPathsList(delegate)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	m := Model{
		cfg:         cfg,
		step:        stepBrowse,
		scope:       extension.ScopeProject,
		focus:       focusSearch,
		searchInput: ti,
		resultList:  resultList,
		pathsList:   pathsList,
		spinner:     sp,
		delegate:    delegate,
	}
	configureSearchInput(&m)
	return m
}

func newResultsList(delegate *itemDelegate) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetStatusBarItemName("skill", "skills")
	l.SetShowPagination(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = keys.Up
	l.KeyMap.CursorDown = keys.Down
	l.KeyMap.NextPage = keys.PgDown
	l.KeyMap.PrevPage = keys.PgUp
	l.KeyMap.Filter = keys.Find
	l.KeyMap.ClearFilter = keys.Home
	return l
}

func newPathsList(delegate *itemDelegate) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = keys.Up
	l.KeyMap.CursorDown = keys.Down
	l.KeyMap.NextPage = keys.PgDown
	l.KeyMap.PrevPage = keys.PgUp
	return l
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) ApplyTheme(styles theme.Styles) {
	m.cfg.Styles = styles
	m.delegate.styles = styles
}

func (m Model) FooterHint() string {
	if m.hostErrMsg() != "" && m.step == stepBrowse && len(m.results) == 0 {
		return ""
	}
	switch m.step {
	case stepPaths:
		return "Tab · project/global   Space · toggle path   Enter · install   Esc · back"
	default:
		if m.searching {
			return fmt.Sprintf("Searching skills.sh for %q…  Esc: cancel", m.query)
		}
		if len(m.results) > 0 {
			n := len(m.resultList.VisibleItems())
			if m.resultList.FilterValue() != "" {
				return fmt.Sprintf("%q · %d/%d · ↑↓ Enter · / · Esc", m.query, n, len(m.results))
			}
			return fmt.Sprintf("%q · %d · ↑↓ Enter · / · Esc", m.query, len(m.results))
		}
		return "skills.sh · Enter · Esc"
	}
}

func (m Model) ShortHelp() []key.Binding {
	switch m.step {
	case stepPaths:
		return []key.Binding{keys.Scope, keys.Toggle, keys.Enter, keys.Cancel}
	default:
		if m.searching {
			return []key.Binding{keys.Cancel}
		}
		if len(m.results) > 0 {
			return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.InstallSearch, keys.Cancel}
		}
		return []key.Binding{keys.Enter, keys.Cancel}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.updateKeys(msg)
	case SearchDoneMsg:
		return m.handleSearchDone(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	switch m.step {
	case stepBrowse:
		return m.updateBrowse(msg)
	case stepPaths:
		return m.updatePaths(msg)
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.Msg) (Model, tea.Cmd) {
	if m.focus == focusSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.resultList, cmd = m.resultList.Update(msg)
	return m, cmd
}

func (m Model) updatePaths(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.pathsList, cmd = m.pathsList.Update(msg)
	return m, cmd
}

func (m Model) updateKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.step {
	case stepBrowse:
		return m.handleBrowseKeys(msg)
	case stepPaths:
		return m.handlePathsKeys(msg)
	}
	return m, nil
}

func (m Model) handleBrowseKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.focus == focusSearch || len(m.results) == 0 {
			query := strings.TrimSpace(m.searchInput.Value())
			if query == "" {
				return m, func() tea.Msg { return HintMsg{Text: "Enter a search keyword, then press Enter"} }
			}
			m.rememberSearchQuery(query)
			m.query = query
			m.searching = true
			m.results = nil
			m.resultList.SetItems(nil)
			m.resultList.SetShowStatusBar(false)
			return m, m.searchCmd()
		}
		candidate, ok := m.selectedCandidate()
		if !ok {
			return m, func() tea.Msg { return HintMsg{Text: "Select a skill from the list"} }
		}
		m.selected = candidate
		m.step = stepPaths
		m.scope = extension.ScopeProject
		m.targets = newDirChoices(m.cfg.AgentIDs)
		items := dirChoicesToItems(m.targets)
		m.delegate.SetHeight(listHeightForItems(items))
		m.pathsList.SetItems(items)
		m.pathsList.Select(0)
		return m, nil

	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		if m.resultList.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.resultList, cmd = m.resultList.Update(msg)
			return m, cmd
		}
		return m, func() tea.Msg { return ClosedMsg{Hint: "Install cancelled"} }

	case msg.String() == "/" || key.Matches(msg, keys.InstallSearch):
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Up, keys.Down):
		if m.focus == focusSearch && len(m.results) == 0 {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
		if len(m.results) == 0 {
			return m, nil
		}
		m.focus = focusResults
		var cmd tea.Cmd
		m.resultList, cmd = m.resultList.Update(msg)
		return m, cmd
	}

	if m.focus == focusSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.resultList, cmd = m.resultList.Update(msg)
	return m, cmd
}

func (m Model) handlePathsKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.step = stepBrowse
		m.focus = focusResults
		return m, nil
	case key.Matches(msg, keys.Enter):
		agentIDs := selectedAgentIDs(m.targets)
		if len(agentIDs) == 0 {
			return m, func() tea.Msg { return HintMsg{Text: "Select at least one install path (Space to toggle)"} }
		}
		if m.scope == extension.ScopeGlobal && m.cfg.Home == "" {
			return m, func() tea.Msg { return HintMsg{Text: "HOME is not set; cannot install globally"} }
		}
		return m, func() tea.Msg { return RequestInstallMsg{AgentIDs: agentIDs, Scope: m.scope} }
	case key.Matches(msg, keys.Scope):
		if m.scope == extension.ScopeGlobal {
			m.scope = extension.ScopeProject
		} else {
			m.scope = extension.ScopeGlobal
		}
		return m, nil
	case key.Matches(msg, keys.Toggle):
		idx := m.pathsList.Index()
		if idx >= 0 && idx < len(m.targets) {
			m.targets[idx].desired = !m.targets[idx].desired
			items := dirChoicesToItems(m.targets)
			m.pathsList.SetItems(items)
			m.pathsList.Select(idx)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.pathsList, cmd = m.pathsList.Update(msg)
	return m, cmd
}

func (m Model) handleSearchDone(msg SearchDoneMsg) (Model, tea.Cmd) {
	m.searching = false
	if msg.Err != nil {
		if m.cfg.SetErrMsg != nil {
			m.cfg.SetErrMsg(msg.Err.Error())
		}
		m.focus = focusSearch
		m.results = nil
		m.resultList.SetItems(nil)
		m.resultList.SetShowStatusBar(false)
		return m, nil
	}
	if m.cfg.ClearErr != nil {
		m.cfg.ClearErr()
	}
	m.results = msg.Results
	m.syncSearchSuggestions()
	m.focus = focusResults
	items := resultsToItems(msg.Results)
	m.delegate.SetHeight(listHeightForItems(items))
	m.resultList.SetItems(items)
	m.resultList.SetShowStatusBar(len(items) > 0)
	if len(items) > 0 {
		m.resultList.Select(0)
	}
	return m, nil
}

func (m Model) searchCmd() tea.Cmd {
	provider := m.cfg.Provider
	query := m.query
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		results, err := provider.Search(query)
		return SearchDoneMsg{Results: results, Err: err}
	})
}

func (m Model) selectedCandidate() (domaininstall.Candidate, bool) {
	item := m.resultList.SelectedItem()
	if item == nil {
		return domaininstall.Candidate{}, false
	}
	row, ok := item.(Row)
	if !ok {
		return domaininstall.Candidate{}, false
	}
	for _, c := range m.results {
		if c.Name == row.Title && c.Source == row.Desc {
			return c, true
		}
	}
	return domaininstall.Candidate{}, false
}

func (m Model) View() string {
	innerWidth := m.dialogWidth()
	listHeight := m.listHeight()
	var body string
	switch {
	case m.step == stepPaths:
		body = m.renderPaths(innerWidth, listHeight)
	default:
		body = m.renderBrowse(innerWidth, listHeight)
	}
	return body
}

func (m Model) PlaceInPane(paneWidth, paneHeight int) string {
	return lipgloss.Place(paneWidth, paneHeight, lipgloss.Center, lipgloss.Top, m.View())
}

// Selected returns the skill chosen for install.
func (m Model) Selected() domaininstall.Candidate {
	return m.selected
}

// WithSelected prepares the paths step for a chosen candidate (used by tests).
func (m Model) WithSelected(c domaininstall.Candidate) Model {
	m.selected = c
	m.step = stepPaths
	m.scope = extension.ScopeProject
	m.targets = newDirChoices(m.cfg.AgentIDs)
	return m
}

func (m Model) dialogWidth() int {
	w := m.width
	if w <= 0 {
		w = 60
	}
	return min(max(32, w-6), 64)
}

func (m Model) listHeight() int {
	h := min(max(8, m.height-10), 22)
	if h < 4 {
		return 4
	}
	return h
}

func (m Model) styles() theme.Styles {
	return m.cfg.Styles
}

func (m Model) hostErrMsg() string {
	if m.cfg.GetErrMsg != nil {
		return m.cfg.GetErrMsg()
	}
	return ""
}

func (m Model) Searching() bool { return m.searching }
