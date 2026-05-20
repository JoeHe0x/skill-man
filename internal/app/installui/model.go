package installui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
	serviceinstall "github.com/JoeHe0x/skill-man/internal/service/install"
)

type step int

const (
	stepBrowse step = iota
	stepPaths
	stepConfirm
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
	targets  []dirChoice

	searching   bool
	installing  bool
	quitPending bool

	width  int
	height int

	recentQueries []string
	searchInput   textinput.Model
	resultList    list.Model
	pathsList     list.Model
	spinner       spinner.Model
	progress      progress.Model
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

	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(36))

	m := Model{
		cfg:         cfg,
		step:        stepBrowse,
		focus:       focusSearch,
		searchInput: ti,
		resultList:  resultList,
		pathsList:   pathsList,
		spinner:     sp,
		progress:    bar,
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
	m.progress = progress.New(progress.WithDefaultGradient(), progress.WithWidth(m.progress.Width))
}

func (m *Model) BeginInstall() tea.Cmd {
	m.installing = true
	m.quitPending = false
	m.progress = progress.New(progress.WithDefaultGradient(), progress.WithWidth(m.progress.Width))
	m.progress.ShowPercentage = true
	return tea.Batch(m.progress.SetPercent(0), progressTickCmd())
}

func (m *Model) EndInstall() {
	m.installing = false
	m.quitPending = false
}

func (m Model) FooterHint() string {
	if m.hostErrMsg() != "" && m.step == stepBrowse && len(m.results) == 0 {
		return ""
	}
	if m.installing {
		return fmt.Sprintf("Installing %s…  Esc twice to cancel", m.selected.Name)
	}
	switch m.step {
	case stepConfirm:
		return "Enter · run install   Esc · back to paths"
	case stepPaths:
		return "Space · toggle path   Enter · confirm   Esc · back to results"
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
	if m.installing {
		return []key.Binding{keys.Cancel}
	}
	switch m.step {
	case stepConfirm:
		return []key.Binding{keys.Enter, keys.Cancel}
	case stepPaths:
		return []key.Binding{keys.Toggle, keys.Enter, keys.Cancel}
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
	case progress.FrameMsg:
		if !m.installing {
			return m, nil
		}
		next, cmd := m.progress.Update(msg)
		m.progress = next.(progress.Model)
		return m, cmd
	case ProgressTickMsg:
		return m.handleProgressTick()
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
	if m.installing {
		return m.handleInstallingKeys(msg)
	}
	switch m.step {
	case stepBrowse:
		return m.handleBrowseKeys(msg)
	case stepPaths:
		return m.handlePathsKeys(msg)
	case stepConfirm:
		return m.handleConfirmKeys(msg)
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
		if len(selectedAgentIDs(m.targets)) == 0 {
			return m, func() tea.Msg { return HintMsg{Text: "Select at least one install path (Space to toggle)"} }
		}
		m.step = stepConfirm
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

func (m Model) handleConfirmKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		m.step = stepPaths
		return m, nil
	case key.Matches(msg, keys.Enter):
		agentIDs := selectedAgentIDs(m.targets)
		if len(agentIDs) == 0 {
			m.step = stepPaths
			return m, func() tea.Msg { return HintMsg{Text: "Select at least one install path (Space to toggle)"} }
		}
		return m, func() tea.Msg { return RequestInstallMsg{AgentIDs: agentIDs} }
	}
	return m, nil
}

func (m Model) handleInstallingKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel), key.Matches(msg, keys.Home):
		if !m.quitPending {
			m.quitPending = true
			return m, func() tea.Msg { return HintMsg{Text: "Install in progress — press Esc again to cancel"} }
		}
		m.installing = false
		m.quitPending = false
		return m, func() tea.Msg { return CancelInstallMsg{} }
	}
	return m, nil
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

func (m Model) handleProgressTick() (Model, tea.Cmd) {
	if !m.installing {
		return m, nil
	}
	var cmds []tea.Cmd
	if m.progress.Percent() < 0.9 {
		cmds = append(cmds, m.progress.IncrPercent(0.04))
	}
	cmds = append(cmds, progressTickCmd())
	return m, tea.Batch(cmds...)
}

func (m Model) selectedCandidate() (domaininstall.Candidate, bool) {
	item := m.resultList.SelectedItem()
	if item == nil {
		return domaininstall.Candidate{}, false
	}
	pi, ok := item.(panel.Item)
	if !ok {
		return domaininstall.Candidate{}, false
	}
	for _, c := range m.results {
		if c.Name == pi.Title && c.Source == pi.Desc {
			return c, true
		}
	}
	return domaininstall.Candidate{}, false
}

func progressTickCmd() tea.Cmd {
	return tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg {
		return ProgressTickMsg{}
	})
}

func (m Model) InstallCmd(ctx context.Context, agentIDs []string) tea.Cmd {
	candidate := m.selected
	provider := m.cfg.Provider
	cwd, home := m.cfg.CWD, m.cfg.Home
	return func() tea.Msg {
		name, err := provider.Install(ctx, cwd, home, candidate, agentIDs)
		return InstallDoneMsg{Name: name, Err: err}
	}
}

func (m Model) View() string {
	innerWidth := m.dialogWidth()
	listHeight := m.listHeight()
	var body string
	switch {
	case m.installing:
		body = m.renderInstalling(innerWidth)
	case m.step == stepConfirm:
		body = m.renderConfirm(innerWidth)
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

func (m Model) Searching() bool   { return m.searching }
func (m Model) Installing() bool  { return m.installing }
func (m Model) QuitPending() bool { return m.quitPending }
