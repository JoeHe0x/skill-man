package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
)

type focusPane int

const (
	focusPaneList focusPane = iota
	focusPanePreview
)

type SessionState int

const (
	stateHome SessionState = iota
	stateListing
	stateSearching
	stateInstalling
	stateConfirming
	stateHelpOverlay
	stateBindingAgent
	stateFilteringAgent
	stateInspecting
	stateCommandPalette
)

type pendingAction struct {
	name       string
	skillName  string
	skill      *skilldomain.Skill
	mcpName    string
	mcp        *mcpdomain.Server
	mcpMembers []*mcpdomain.Server
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

	cwd            string
	home           string
	status         string
	errMsg         string
	footerFlash    string
	footerFlashTag int
	footerContext  string
	focusedPane    focusPane
	agentIDs       []string
	allAgents      []agent.Agent

	prompt            *promptModel
	installFlow       *installFlow
	installCancel     context.CancelFunc
	pending           *pendingAction
	palette           *commandPalette
	helpOverlay       helpOverlay
	list              list.Model
	listDelegate      *itemDelegate
	agentList         list.Model
	agentListDelegate *itemDelegate
	tree              fileTreeModel
	preview           viewport.Model
	spinner           spinner.Model
	help              help.Model

	styles     styles
	darkTheme  bool
	themeReady bool
	registry   *commands.Registry

	activeTab   panel.Tab
	panels      *panel.Registry
	binds       bindSession
	previewBody string
	previewGen  int // increments on each preview request; stale loads are dropped

	skillManager manager.ExtensionManager[*skilldomain.Skill]
	mcpManager   *servicemcp.Manager
}

func New(cwd, home string) *Model {
	allAgents := agent.DefaultAgents()
	registry := commands.NewRegistry()
	uiStyles := newStyles(true)

	mainDelegate := newItemDelegate(uiStyles)
	skillList := list.New([]list.Item{}, mainDelegate, 0, 0)
	skillList.Title = ""
	skillList.SetShowTitle(false)
	skillList.SetShowStatusBar(false)
	skillList.SetFilteringEnabled(false)
	skillList.SetShowHelp(false)
	skillList.DisableQuitKeybindings()

	agentDelegate := newItemDelegate(uiStyles)
	agentList := list.New([]list.Item{}, agentDelegate, 0, 0)
	agentList.Title = ""
	agentList.SetShowTitle(false)
	agentList.SetShowStatusBar(false)
	agentList.SetShowPagination(false)
	agentList.SetFilteringEnabled(false)
	agentList.SetShowHelp(false)
	agentList.DisableQuitKeybindings()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	preview := viewport.New(0, 0)
	preview.SetContent(welcomePreview)

	fileTree := newFileTreeModel(uiStyles)

	skillManager := manager.NewManager[*skilldomain.Skill](service.SkillScanStrategy{})
	panels := newPanelRegistry()

	m := Model{
		state:       stateHome,
		lastState:   stateHome,
		cwd:         cwd,
		home:        home,
		status:      "loading",
		focusedPane: focusPaneList,

		activeTab:         panel.TabSkills,
		panels:            panels,
		list:              skillList,
		listDelegate:      mainDelegate,
		agentList:         agentList,
		agentListDelegate: agentDelegate,
		tree:              fileTree,
		preview:           preview,
		spinner:           sp,
		help:              help.New(),
		styles:            uiStyles,
		darkTheme:         true,
		registry:          registry,
		agentIDs:          []string{"all"},
		allAgents:         allAgents,
		previewBody:       welcomePreview,
		skillManager:      skillManager,
		mcpManager:        servicemcp.NewManager(),
	}

	m.list.KeyMap.CursorUp = keys.Up
	m.list.KeyMap.CursorDown = keys.Down
	m.list.KeyMap.NextPage = keys.PgDown
	m.list.KeyMap.PrevPage = keys.PgUp
	m.agentList.KeyMap.CursorUp = keys.Up
	m.agentList.KeyMap.CursorDown = keys.Down
	m.agentList.KeyMap.NextPage = keys.PgDown
	m.agentList.KeyMap.PrevPage = keys.PgUp
	m.preview.KeyMap.PageUp = keys.PgUp
	m.preview.KeyMap.PageDown = keys.PgDown

	m.helpOverlay = newHelpOverlay()
	m.configureMainList()
	initHelpStyles(&m.help, uiStyles)
	m.syncSelectionPreview()
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.scanAllCmd(), detectTerminalThemeCmd())
}

func (m *Model) showPrompt(label, placeholder string, action func(m *Model, text string) tea.Cmd) tea.Cmd {
	m.prompt = newPromptModel(label, placeholder, action)
	return textinput.Blink
}

func (m *Model) hidePrompt() {
	m.prompt = nil
}

func (m *Model) scanAllCmd() tea.Cmd {
	return m.panels.ScanAllCmd(m.cwd, m.home, slices.Clone(m.allAgents))
}

// mcpMembersForConfigKey returns every scanned server row for a config key (authoritative for bind UI).
func (m *Model) mcpMembersForConfigKey(key string) []*mcpdomain.Server {
	if key == "" {
		return nil
	}
	var out []*mcpdomain.Server
	for _, srv := range m.panels.MCPServers() {
		k := srv.ConfigKey
		if k == "" {
			k = srv.GetName()
		}
		if strings.EqualFold(k, key) {
			out = append(out, srv)
		}
	}
	return out
}

func (m *Model) scanSkillsCmd() tea.Cmd {
	return m.panels.Get(panel.TabSkills).ScanCmd(m.cwd, m.home, slices.Clone(m.allAgents))
}

func (m *Model) previewSkillCmd(skill *skilldomain.Skill) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	item := panel.Item{Kind: panel.ItemSkill, Skill: skill}
	return m.activePanel().SyncPreview(item, width, &m.previewGen)
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

func (m *Model) removeSkillCmd(skill *skilldomain.Skill) tea.Cmd {
	return func() tea.Msg {
		if err := m.skillManager.Remove(context.Background(), skill, m.cwd, m.home); err != nil {
			return mutationCompletedMsg{err: err}
		}
		return mutationCompletedMsg{message: fmt.Sprintf("removed %s", skill.GetName())}
	}
}

func (m *Model) toggleDisableMCPCmd(srv *mcpdomain.Server) tea.Cmd {
	return m.toggleDisableMCPKeyCmd([]*mcpdomain.Server{srv})
}

func (m *Model) toggleDisableMCPKeyCmd(members []*mcpdomain.Server) tea.Cmd {
	label := "disabled"
	if !mcpKeyDisabled(members) {
		label = "enabled"
	}
	return m.mcpKeyMutationCmd(members, m.mcpManager.ToggleDisable, label)
}

func (m *Model) removeMCPCmd(srv *mcpdomain.Server) tea.Cmd {
	return m.removeMCPKeyCmd([]*mcpdomain.Server{srv})
}

func (m *Model) removeMCPKeyCmd(members []*mcpdomain.Server) tea.Cmd {
	return m.mcpKeyMutationCmd(members, m.mcpManager.Remove, "removed")
}

func (m *Model) mcpKeyMutationCmd(members []*mcpdomain.Server, apply func(*mcpdomain.Server) error, label string) tea.Cmd {
	cp := slices.Clone(members)
	return func() tea.Msg {
		var errs []error
		for _, s := range cp {
			if err := apply(s); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return mutationCompletedMsg{err: errors.Join(errs...), targetTab: panel.TabMCP}
		}
		key := cp[0].ConfigKey
		if key == "" {
			key = cp[0].GetName()
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("%s MCP `%s` (%d locations)", label, key, len(cp)),
			selectName: key,
			targetTab:  panel.TabMCP,
		}
	}
}

func (m *Model) toggleDisableSkillCmd(skill *skilldomain.Skill) tea.Cmd {
	return func() tea.Msg {
		if err := m.skillManager.ToggleDisable(context.Background(), skill); err != nil {
			return mutationCompletedMsg{err: err}
		}
		action := "disabled"
		if skill.IsDisabled() {
			action = "enabled"
		}
		return mutationCompletedMsg{
			message:    fmt.Sprintf("%s %s", action, skill.GetName()),
			selectName: skill.GetName(), // trigger reselect
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

func (m *Model) updateSkillCmd(skill *skilldomain.Skill) tea.Cmd {
	return func() tea.Msg {
		result, err := service.UpdateSkill(*skill)
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
	skills := slices.Clone(m.panels.Skills())
	return func() tea.Msg {
		var g errgroup.Group
		var mu sync.Mutex
		updated := 0
		firstName := ""

		for _, skill := range skills {
			if !skill.IsManaged() || skill.SourceKind != "local" {
				continue
			}
			g.Go(func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic updating skill %s: %v", skill.GetName(), r)
					}
				}()
				if _, err := service.UpdateSkill(*skill); err != nil {
					return err
				}
				mu.Lock()
				updated++
				if firstName == "" {
					firstName = skill.GetName()
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
	items := commandListItems(m.registry.Specs())
	m.setMainListItems(items)
	if len(m.list.Items()) > 0 {
		m.list.Select(0)
	}
}

func (m *Model) refreshActiveList() {
	items := panelToListItems(m.activePanel().ListItems(m.agentIDs))
	m.setMainListItems(items)
	if len(m.list.Items()) > 0 {
		m.list.Select(0)
	}
}

func (m *Model) setMainListItems(items []list.Item) {
	m.listDelegate.SetHeight(listHeightForItems(items))
	m.list.SetItems(items)
}

func (m *Model) setAgentListItems(items []list.Item) {
	m.agentListDelegate.SetHeight(listHeightForItems(items))
	m.agentList.SetItems(items)
}

func (m *Model) switchExtensionTab(reverse bool) tea.Cmd {
	next := m.activeTab.Next()
	if reverse {
		next = m.activeTab.Prev()
	}
	return m.setActiveTab(next)
}

func (m *Model) setActiveTab(tab panel.Tab) tea.Cmd {
	if m.activeTab == tab {
		return nil
	}
	m.activeTab = tab
	m.clearError()

	if m.state == stateInspecting || m.state == stateBindingAgent || m.state == stateFilteringAgent || m.state == stateConfirming || m.state == stateInstalling {
		m.state = stateListing
		m.clearInstallFlow()
	}

	m.refreshActiveList()
	if preview := m.activePanel().StaticPreview(); preview != "" {
		m.preview.SetContent(preview)
		m.previewBody = preview
		m.setFooterContext(fmt.Sprintf("%d %s · agents: %s", m.activePanel().Count(), m.activePanel().CountLabel(), m.agentDisplay()))
		return nil
	}

	m.setFooterContext(fmt.Sprintf("%d %s · agents: %s", m.activePanel().Count(), m.activePanel().CountLabel(), m.agentDisplay()))
	return m.syncSelectionPreview()
}

func (m *Model) selectSkillByName(name string) bool {
	skill, ok := m.findSkillByName(name)
	if !ok {
		return false
	}

	m.state = stateListing
	m.lastState = stateListing
	m.refreshActiveList()
	for idx, item := range m.list.Items() {
		li, ok := item.(listItem)
		if ok && li.kind == itemKindSkill && strings.EqualFold(li.skill.GetName(), skill.GetName()) {
			m.list.Select(idx)
			break
		}
	}
	return true
}

func (m *Model) selectMCPByName(name string) bool {
	m.refreshActiveList()
	for idx, item := range m.list.Items() {
		li, ok := item.(listItem)
		if !ok || li.kind != itemKindMCP {
			continue
		}
		if strings.EqualFold(li.mcpKey, name) ||
			strings.EqualFold(li.mcp.GetName(), name) ||
			strings.EqualFold(li.mcp.ConfigKey, name) {
			m.list.Select(idx)
			return true
		}
	}
	return false
}

func (m *Model) findMCPByName(name string) (*mcpdomain.Server, bool) {
	for _, srv := range m.panels.MCPServers() {
		if strings.EqualFold(srv.GetName(), name) {
			return srv, true
		}
	}
	return nil, false
}

func (m *Model) findSkillByName(name string) (*skilldomain.Skill, bool) {
	for _, skill := range m.panels.Skills() {
		if strings.EqualFold(skill.GetName(), name) {
			return skill, true
		}
	}
	return nil, false
}

func (m *Model) syncSelectionPreview() tea.Cmd {
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok {
		m.preview.SetContent(m.styles.emptyPreview.Render("No selection"))
		return nil
	}

	if selected.kind == itemKindCommand {
		m.previewBody = service.RenderCommandPreview(
			selected.command.Name,
			selected.command.Usage,
			selected.command.Summary,
			selected.command.Implemented,
		)
		m.preview.SetContent(m.previewBody)
		return nil
	}

	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	return m.activePanel().SyncPreview(listItemToPanel(selected), width, &m.previewGen)
}

// reportError surfaces a failure in the status bar and footer (single handling site).
func (m *Model) reportError(err error) {
	if err == nil {
		return
	}
	m.status = "error"
	m.errMsg = err.Error()
}

func (m *Model) clearError() {
	m.errMsg = ""
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

func (m *Model) activeAgents() []agent.Agent {
	if len(m.agentIDs) == 0 || slices.Contains(m.agentIDs, "all") {
		return m.allAgents
	}
	var out []agent.Agent
	for _, id := range m.agentIDs {
		if a, ok := agent.AgentByID(id); ok {
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

func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return ""
}

const welcomePreview = `# Welcome to skill-man

Tab      skills / mcp tabs
Ctrl+P   command palette
?        expand footer key hints
F1       command reference (this screen)
Ctrl+L   list          Ctrl+F  find
Ctrl+A   agent         Ctrl+D  install
Ctrl+N   init          Ctrl+R  reload
Ctrl+U   update
Enter    inspect       Del     remove
Ctrl+C   quit

Footer shows context keys; green flashes confirm actions.`
