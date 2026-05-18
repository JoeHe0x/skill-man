package app

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"skill-man/internal/commands"
	"skill-man/internal/domain"
	"skill-man/internal/service"
)

type SessionState int

const (
	stateHome SessionState = iota
	stateListing
	stateSearching
	stateConfirming
	stateViewingHelp
	stateBindingAgent
	stateInspecting
)

type pendingAction struct {
	name      string
	skillName string
	skill     domain.Skill
}

// promptModel is a temporary text input shown on demand for commands that
// need user text (find query, add path, init name).
type promptModel struct {
	input  textinput.Model
	label  string
	action func(m *Model, text string) tea.Cmd
}

func newPromptModel(label, placeholder string, action func(m *Model, text string) tea.Cmd) *promptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Prompt = ""
	ti.Focus()
	return &promptModel{
		input:  ti,
		label:  label,
		action: action,
	}
}

type Model struct {
	state     SessionState
	lastState SessionState

	width  int
	height int

	cwd       string
	home      string
	status    string
	hint      string
	errMsg    string
	agentIDs  []string
	allAgents []domain.Agent

	prompt    *promptModel
	list      list.Model
	agentList list.Model
	tree      fileTreeModel
	preview   viewport.Model
	spinner   spinner.Model

	styles   styles
	registry *commands.Registry

	skills       []domain.Skill
	filtered     []domain.Skill
	bindingSkill domain.Skill
	logs         []string
	previewBody  string
	previewGen   int // increments on each preview request; stale loads are dropped

	pending *pendingAction
}

func (m *Model) updateHint() {
	if m.errMsg != "" {
		return // leave error messages visible
	}

	switch m.state {
	case stateHome:
		m.hint = "?/F1:Help  Ctrl+L:List  Ctrl+F:Find  Ctrl+A:Agent  Ctrl+R:Reload  Ctrl+U:Update  Ctrl+C:Quit"
	case stateListing, stateSearching:
		selected, ok := m.list.SelectedItem().(listItem)
		if ok && selected.kind == itemKindSkill {
			m.hint = "?/F1:Help  Enter:Inspect  X:Toggle  B:Bind  Del:Remove  Ctrl+L:List  Ctrl+F:Find  Ctrl+A:Agent  Ctrl+R:Reload  Ctrl+U:Update  Ctrl+C:Quit"
		} else {
			m.hint = "?/F1:Help  Ctrl+L:List  Ctrl+F:Find  Ctrl+A:Agent  Ctrl+R:Reload  Ctrl+U:Update  Ctrl+C:Quit"
		}
	}
}

func New(cwd, home string) *Model {
	allAgents := domain.DefaultAgents()
	registry := commands.NewRegistry()
	uiStyles := newStyles()

	skillList := list.New([]list.Item{}, newItemDelegate(uiStyles), 0, 0)
	skillList.Title = ""
	skillList.SetShowTitle(false)
	skillList.SetShowStatusBar(false)
	skillList.SetFilteringEnabled(false)
	skillList.SetShowHelp(false)
	skillList.DisableQuitKeybindings()

	agentList := list.New([]list.Item{}, newItemDelegate(uiStyles), 0, 0)
	agentList.Title = ""
	agentList.SetShowTitle(false)
	agentList.SetShowStatusBar(false)
	agentList.SetFilteringEnabled(false)
	agentList.SetShowHelp(false)
	agentList.DisableQuitKeybindings()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	preview := viewport.New(0, 0)
	preview.SetContent(welcomePreview)

	fileTree := newFileTreeModel(uiStyles)

	m := Model{
		state:     stateHome,
		lastState: stateHome,
		cwd:       cwd,
		home:      home,
		status:    "loading",
		hint:      "?/F1:Help  Ctrl+L:List  Ctrl+F:Find  Ctrl+A:Agent  Ctrl+R:Reload  Ctrl+U:Update  Ctrl+C:Quit",

		list:        skillList,
		agentList:   agentList,
		tree:        fileTree,
		preview:     preview,
		spinner:     sp,
		styles:      uiStyles,
		registry:    registry,
		agentIDs:    []string{"all"},
		allAgents:   allAgents,
		previewBody: welcomePreview,
	}

	m.list.KeyMap.CursorUp = keys.Up
	m.list.KeyMap.CursorDown = keys.Down
	m.list.KeyMap.NextPage = keys.PgDown
	m.list.KeyMap.PrevPage = keys.PgUp
	m.preview.KeyMap.PageUp = keys.PgUp
	m.preview.KeyMap.PageDown = keys.PgDown

	m.syncSelectionPreview()
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.scanSkillsCmd())
}

func (m *Model) showPrompt(label, placeholder string, action func(m *Model, text string) tea.Cmd) tea.Cmd {
	m.prompt = newPromptModel(label, placeholder, action)
	return textinput.Blink
}

func (m *Model) hidePrompt() {
	m.prompt = nil
}

func (m *Model) scanSkillsCmd() tea.Cmd {
	cwd := m.cwd
	home := m.home
	agents := slices.Clone(m.allAgents)
	return func() tea.Msg {
		skills, err := service.ScanSkills(context.Background(), cwd, home, agents)
		return skillsScannedMsg{skills: skills, err: err}
	}
}

func (m *Model) previewSkillCmd(skill domain.Skill) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	m.previewGen++
	gen := m.previewGen
	return func() tea.Msg {
		content, err := service.RenderSkillPreview(skill, width)
		return previewLoadedMsg{content: content, err: err, gen: gen}
	}
}

func (m *Model) initSkillCmd(name string) tea.Cmd {
	root := m.cwd
	return func() tea.Msg {
		path, createdName, err := service.InitializeSkill(root, name)
		if err != nil {
			return mutationCompletedMsg{err: err}
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("created skill template at %s", path),
			selectName: createdName,
		}
	}
}

func (m *Model) removeSkillCmd(skill domain.Skill) tea.Cmd {
	return func() tea.Msg {
		if err := service.RemoveSkill(skill, m.cwd, m.home); err != nil {
			return mutationCompletedMsg{err: err}
		}
		return mutationCompletedMsg{message: fmt.Sprintf("removed %s", skill.Name)}
	}
}

func (m *Model) toggleDisableSkillCmd(skill domain.Skill) tea.Cmd {
	return func() tea.Msg {
		if err := service.ToggleDisableSkill(skill); err != nil {
			return mutationCompletedMsg{err: err}
		}
		action := "disabled"
		if skill.Disabled {
			action = "enabled"
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("%s %s", action, skill.Name),
			selectName: skill.Name, // trigger reselect
		}
	}
}

func (m *Model) addSkillCmd(source string) tea.Cmd {
	cwd := m.cwd
	agents := m.activeAgents()
	return func() tea.Msg {
		result, err := service.InstallLocalSkill(cwd, source, agents)
		if err != nil {
			return mutationCompletedMsg{err: err}
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("installed %s -> %s", result.Name, result.TargetPath),
			selectName: result.Name,
		}
	}
}

func (m *Model) updateSkillCmd(skill domain.Skill) tea.Cmd {
	return func() tea.Msg {
		result, err := service.UpdateSkill(skill)
		if err != nil {
			return mutationCompletedMsg{err: err}
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("updated %s from %s", result.Name, result.SourcePath),
			selectName: result.Name,
		}
	}
}

func (m *Model) updateAllSkillsCmd() tea.Cmd {
	skills := slices.Clone(m.skills)
	return func() tea.Msg {
		var g errgroup.Group
		var mu sync.Mutex
		updated := 0
		firstName := ""

		for _, skill := range skills {
			if !skill.Managed || skill.SourceKind != "local" {
				continue
			}
			sk := skill
			g.Go(func() error {
				if _, err := service.UpdateSkill(sk); err != nil {
					return err
				}
				mu.Lock()
				updated++
				if firstName == "" {
					firstName = sk.Name
				}
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return mutationCompletedMsg{err: err}
		}

		if updated == 0 {
			return mutationCompletedMsg{message: "no managed local skills available to update"}
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("updated %d skill(s)", updated),
			selectName: firstName,
		}
	}
}

func (m *Model) setCommandItems() {
	m.list.SetItems(commandItems(m.registry.Specs()))
	if len(m.list.Items()) > 0 {
		m.list.Select(0)
	}
}

func (m *Model) setSkillItems(skills []domain.Skill) {
	m.filtered = slices.Clone(skills)
	m.list.SetItems(skillItems(skills, m.agentIDs))
	if len(m.list.Items()) > 0 {
		m.list.Select(0)
	}
}

func (m *Model) selectSkillByName(name string) bool {
	skill, ok := m.findSkillByName(name)
	if !ok {
		return false
	}

	m.state = stateListing
	m.lastState = stateListing
	m.setSkillItems(m.skills)
	for idx, item := range m.list.Items() {
		li, ok := item.(listItem)
		if ok && li.kind == itemKindSkill && strings.EqualFold(li.skill.Name, skill.Name) {
			m.list.Select(idx)
			break
		}
	}
	return true
}

func (m *Model) findSkillByName(name string) (domain.Skill, bool) {
	for _, skill := range m.skills {
		if strings.EqualFold(skill.Name, name) {
			return skill, true
		}
	}
	return domain.Skill{}, false
}

func (m *Model) syncSelectionPreview() tea.Cmd {
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.preview.SetContent(m.styles.emptyPreview.Render("No selection"))
		return nil
	}

	switch selected.kind {
	case itemKindSkill:
		return m.previewSkillCmd(selected.skill)
	case itemKindCommand:
		m.previewBody = service.RenderCommandPreview(
			selected.command.Name,
			selected.command.Usage,
			selected.command.Summary,
			selected.command.Implemented,
		)
		m.preview.SetContent(m.previewBody)
	default:
		m.preview.SetContent(selected.title + "\n\n" + selected.desc)
	}
	return nil
}

func (m *Model) logf(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
	if len(m.logs) > 50 {
		m.logs = m.logs[len(m.logs)-50:]
	}
}

func (m *Model) statusView() string {
	switch {
	case m.errMsg != "":
		return m.styles.statusError.Render(m.errMsg)
	case m.state == stateConfirming:
		return m.styles.statusWarn.Render("confirm")
	case m.status == "ready":
		return m.styles.statusReady.Render(m.status)
	default:
		return m.styles.statusWarn.Render(m.status)
	}
}

func (m *Model) activeAgents() []domain.Agent {
	if len(m.agentIDs) == 0 || slices.Contains(m.agentIDs, "all") {
		return m.allAgents
	}
	var out []domain.Agent
	for _, id := range m.agentIDs {
		if a, ok := domain.AgentByID(id); ok {
			out = append(out, a)
		}
	}
	return out
}

func (m *Model) agentDisplay() string {
	if slices.Contains(m.agentIDs, "all") || len(m.agentIDs) == 0 {
		return "all"
	}
	return strings.Join(m.agentIDs, ",")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) headerHeight() int {
	if m.width >= 80 && m.height >= 24 {
		return 8
	}
	return 2
}

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

const welcomePreview = `# Welcome to skill-man

?/F1  help     Ctrl+L  list      Ctrl+F  find
Ctrl+A  agent    Ctrl+D  add       Ctrl+N  init
Ctrl+R  reload   Ctrl+U  update
Enter   inspect  Del     remove    Ctrl+C  quit

Use the keybindings above to get started.`

const asciiLogo = `
████ █  █ ███ █    ▓      ▓   ▒  ▒▒  ▒  ▒
█    █ █   █  ▓    ▓      ▓▓ ▒▒ ▒  ▒ ▒▒ ▒
███  ██    █  ▓    ▓      ▓ ▒ ▒ ▒▒▒▒ ▒ ▒▒
   █ █ █   █  ▓    ▓      ▓   ▒ ▒  ▒ ▒  ▒
████ █  █ █▓▓ ▓▓▓▓ ▓▓▓▓   ▒   ▒ ▒  ▒ ▒  ░`
