package app

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoeHe0x/skill-man/internal/app/feature"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/theme"
	"github.com/JoeHe0x/skill-man/internal/commands"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
	"github.com/JoeHe0x/skill-man/internal/service/manager"
	servicemcp "github.com/JoeHe0x/skill-man/internal/service/mcp"
	service "github.com/JoeHe0x/skill-man/internal/service/skill"
	usecasebind "github.com/JoeHe0x/skill-man/internal/usecase/bind"
	usecase "github.com/JoeHe0x/skill-man/internal/usecase/extension"
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

type Model struct {
	Core
	state     SessionState
	lastState SessionState

	install    *installFeature
	prompt     *promptFeature
	confirm    *confirmFeature
	bind       *bindFeature
	cmdPalette *paletteFeature
	helpScreen *helpScreenFeature
	listPane
	spinner spinner.Model
	help    help.Model

	styles     theme.Styles
	darkTheme  bool
	themeReady bool
	registry   *commands.Registry

	activeTab   panel.Tab
	panels      *panel.Registry
	previewBody string
	previewGen  int // increments on each preview request; stale loads are dropped

	skillManager manager.ExtensionManager[*skilldomain.Skill]
	mcpManager   *servicemcp.Manager
	mutator      usecase.Mutator
	binder       usecasebind.Binder

	features []feature.Feature
}

func New(cwd, home string) *Model {
	registry := commands.NewRegistry()
	uiStyles := theme.NewStyles(true)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	skillManager := manager.NewManager[*skilldomain.Skill](service.SkillScanStrategy{})
	mcpManager := servicemcp.NewManager()
	panels := newPanelRegistry()

	m := Model{
		Core:      newCore(cwd, home),
		state:     stateHome,
		lastState: stateHome,

		activeTab:    panel.TabSkills,
		panels:       panels,
		listPane:     newListPane(uiStyles),
		spinner:      sp,
		help:         help.New(),
		styles:       uiStyles,
		darkTheme:    true,
		registry:     registry,
		skillManager: skillManager,
		mcpManager:   mcpManager,
		mutator:      usecase.NewMutator(skillManager, mcpManager, cwd, home),
		binder:       usecasebind.NewBinder(skillManager, mcpManager, cwd, home),
	}

	m.listPane.configureKeys()

	m.install = &installFeature{host: &m}
	m.prompt = &promptFeature{host: &m, model: &m}
	m.confirm = &confirmFeature{host: &m}
	m.bind = &bindFeature{host: &m}
	m.cmdPalette = &paletteFeature{host: &m}
	m.helpScreen = &helpScreenFeature{host: &m, overlay: newHelpScreenOverlay()}
	m.features = []feature.Feature{
		m.prompt,
		m.install,
		m.cmdPalette,
		m.helpScreen,
		m.bind,
		&inspectFeature{host: &m},
		&agentFilterFeature{host: &m},
		m.confirm,
	}
	m.configureMainList()
	initHelpStyles(&m.help, uiStyles)
	m.updateFooterForState(m.state)
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.beginScanAllCmd(), theme.DetectCmd())
}

func (m *Model) showPrompt(label, placeholder string, action func(m *Model, text string) tea.Cmd) tea.Cmd {
	return m.prompt.Show(label, placeholder, action)
}

func (m *Model) hidePrompt() {
	m.prompt.Hide()
}

func (m *Model) footerStatsLine() string {
	items := m.activePanel().ListItems(m.agentIDs)
	return fmt.Sprintf("%d %s · agents: %s",
		panel.VisibleListCount(items),
		m.activePanel().CountLabel(),
		m.agentDisplay())
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
	return panel.ScanCmd(m.panels.Get(panel.TabSkills), m.cwd, m.home, slices.Clone(m.allAgents))
}

func (m *Model) previewSkillCmd(skill *skilldomain.Skill) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	item := panel.Item{Kind: panel.ItemSkill, Skill: skill}
	return panel.SyncPreviewCmd(m.activePanel(), item, width, &m.previewGen)
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
		m.transitionTo(stateListing)
	}

	m.refreshActiveList()
	if preview := m.activePanel().StaticPreview(); preview != "" {
		m.preview.SetContent(preview)
		m.previewBody = preview
		m.setFooterContext(m.footerStatsLine())
		return nil
	}

	m.setFooterContext(m.footerStatsLine())
	return m.syncSelectionPreview()
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

func (m *Model) reportError(err error) {
	m.Core.reportError(err)
}

func (m *Model) clearError() {
	m.Core.clearError()
}

func (m *Model) statusView() string {
	switch {
	case m.errMsg != "":
		return m.styles.StatusError.Render(m.errMsg)
	case m.state == stateConfirming:
		return m.styles.StatusWarn.Render("confirm")
	case m.status == "ready":
		return m.styles.StatusReady.Render(m.status)
	default:
		return m.styles.StatusWarn.Render(m.status)
	}
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
