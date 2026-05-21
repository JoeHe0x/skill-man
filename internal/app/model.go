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
	featbind "github.com/JoeHe0x/skill-man/internal/app/feature/bind"
	featconfirm "github.com/JoeHe0x/skill-man/internal/app/feature/confirm"
	feathelp "github.com/JoeHe0x/skill-man/internal/app/feature/help"
	featinstall "github.com/JoeHe0x/skill-man/internal/app/feature/install"
	"github.com/JoeHe0x/skill-man/internal/app/feature/overlay"
	featpalette "github.com/JoeHe0x/skill-man/internal/app/feature/palette"
	featprompt "github.com/JoeHe0x/skill-man/internal/app/feature/prompt"
	"github.com/JoeHe0x/skill-man/internal/app/list"
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

type Model struct {
	Core
	state     SessionState
	lastState SessionState

	install    *featinstall.Feature
	prompt     *featprompt.Feature
	confirm    *featconfirm.Feature
	bind       *featbind.Feature
	cmdPalette *featpalette.Feature
	helpScreen *feathelp.Feature
	list.Pane
	spinner spinner.Model
	help    help.Model

	styles     theme.Styles
	darkTheme  bool
	themeReady bool
	registry   *commands.Registry

	activeTab panel.Tab
	panels    *panel.Registry

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
		Pane:         list.NewPane(uiStyles),
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

	m.ConfigureKeys()

	m.install = featinstall.New(&m)
	m.prompt = featprompt.New(&m)
	m.confirm = featconfirm.New(&m)
	m.bind = featbind.New(&m)
	m.cmdPalette = featpalette.New(&m)
	m.helpScreen = feathelp.New(&m)
	m.features = []feature.Feature{
		m.prompt,
		m.install,
		m.cmdPalette,
		m.helpScreen,
		m.bind,
		overlay.NewInspect(&m),
		overlay.NewAgentFilter(&m),
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

func (m *Model) showPrompt(label, placeholder string, action featprompt.Action) tea.Cmd {
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
	width := m.Preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	item := panel.Item{Kind: panel.ItemSkill, Skill: skill}
	return panel.SyncPreviewCmd(m.activePanel(), item, width, &m.PreviewGen)
}

func (m *Model) setCommandItems() {
	items := commandListItems(m.registry.Specs())
	m.setMainListItems(items)
	if len(m.Main.Items()) > 0 {
		m.Main.Select(0)
	}
}

func (m *Model) refreshActiveList() {
	items := panelToListItems(m.activePanel().ListItems(m.agentIDs))
	m.setMainListItems(items)
	if len(m.Main.Items()) > 0 {
		m.Main.Select(0)
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
		m.Preview.SetContent(preview)
		m.PreviewBody = preview
		m.setFooterContext(m.footerStatsLine())
		return nil
	}

	m.setFooterContext(m.footerStatsLine())
	return m.SyncSelectionPreview()
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
