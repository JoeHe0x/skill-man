package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	featprompt "github.com/JoeHe0x/skill-man/internal/app/feature/prompt"
	applist "github.com/JoeHe0x/skill-man/internal/app/list"
	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/app/session"
	statefallback "github.com/JoeHe0x/skill-man/internal/app/state/fallback"
	statefiltering "github.com/JoeHe0x/skill-man/internal/app/state/filtering"
	statenspect "github.com/JoeHe0x/skill-man/internal/app/state/inspect"
	stateinstalling "github.com/JoeHe0x/skill-man/internal/app/state/installing"
	statelistfilter "github.com/JoeHe0x/skill-man/internal/app/state/listfilter"
	statelisting "github.com/JoeHe0x/skill-man/internal/app/state/listing"
	"github.com/JoeHe0x/skill-man/internal/app/uimsg"
	"github.com/JoeHe0x/skill-man/internal/domain/agent"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

// Model state-host bridges: thin adapters so state/* packages stay Bubble Tea–free
// of Model field access. Inspect keys live in state/inspect; agent filter in state/filtering.

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statelisting.HandleKeys(m, msg)
}

func (m *Model) OpenHelpOverlay() (tea.Model, tea.Cmd) { return m.openHelpOverlay() }

func (m *Model) OpenCommandPalette() (tea.Model, tea.Cmd) { return m.openCommandPalette() }

func (m *Model) SwitchExtensionTab(reverse bool) tea.Cmd { return m.switchExtensionTab(reverse) }

func (m *Model) MainFilterState() list.FilterState { return m.Main.FilterState() }

func (m *Model) MainUpdate(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.Main, cmd = m.Main.Update(msg)
	return cmd
}

func (m *Model) StaticPreview() string { return m.activePanel().StaticPreview() }

func (m *Model) SetPreviewContent(s string) { m.Preview.SetContent(s) }

func (m *Model) ToggleHelpAll() { m.help.ShowAll = !m.help.ShowAll }

func (m *Model) SetFocusedList() { m.focusedPane = focusPaneList }

func (m *Model) SetFocusedPreview() { m.focusedPane = focusPanePreview }

func (m *Model) ShowPrompt(label, placeholder string, action func(text string) tea.Cmd) tea.Cmd {
	return m.showPrompt(label, placeholder, featprompt.Action(action))
}

func (m *Model) HidePrompt() { m.hidePrompt() }

func (m *Model) RefreshActiveList() { m.refreshActiveList() }

func (m *Model) SetMainListItems(items []panel.Item) {
	applist.SetMainItemsFromPanel(m, items)
}

func (m *Model) setMainListItems(items []list.Item) {
	m.SetMainItems(items)
}

func (m *Model) setAgentListItems(items []list.Item) {
	m.SetOverlayItems(items)
}

func (m *Model) ListPane() *applist.Pane { return &m.Pane }

func (m *Model) AppWidth() int { return m.width }

func (m *Model) FindSkillByName(name string) (*skilldomain.Skill, bool) {
	return m.findSkillByName(name)
}

var _ applist.BridgeHost = (*Model)(nil)

func (m *Model) SetAgentIDs(ids []string) { m.agentIDs = ids }

func (m *Model) ActiveAgents() []agent.Agent { return m.activeAgents() }

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowFindPrompt(m) }

func (m *Model) showAddPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowAddPrompt(m) }

func (m *Model) showInitPrompt() (tea.Model, tea.Cmd) { return statelisting.ShowInitPrompt(m) }

func (m *Model) setAgentFilter(id string) { statelisting.SetAgentFilter(m, id) }

func (m *Model) handleMutationCompleted(msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	return statefallback.HandleMutationCompleted(m, msg)
}

func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleWindowResize(m, msg)
}

func (m *Model) handleMouseDispatch(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleMouse(m, msg)
}

func (m *Model) handlePreviewLoaded(msg panel.PreviewLoadedMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandlePreviewLoaded(m, msg)
}

func (m *Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleSpinnerTick(m, msg)
}

func (m *Model) handleProgressFrame(msg progress.FrameMsg) (tea.Model, tea.Cmd) {
	return statefallback.HandleProgressFrame(m, msg)
}

func (m *Model) handleReselectMCP(msg uimsg.ReselectMCP) (tea.Model, tea.Cmd) {
	return statefallback.HandleReselectMCP(m, msg)
}

func (m *Model) handleReselectSkill(msg uimsg.ReselectSkill) (tea.Model, tea.Cmd) {
	return statefallback.HandleReselectSkill(m, msg)
}

func (m *Model) handleFallthroughMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	return statefallback.HandleFallthrough(m, msg)
}

func (m *Model) ApplyMutationResult(msg uimsg.MutationCompleted) (tea.Model, tea.Cmd) {
	return m.applyMutationResult(msg)
}

func (m *Model) SetWindowSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) ResizeComponents() { m.resizeComponents() }

func (m *Model) ResizePaletteInput() { m.cmdPalette.ResizeInput() }

func (m *Model) HandleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	return m.handleMouseMsg(msg)
}

func (m *Model) PreviewGeneration() int { return m.PreviewGen }

func (m *Model) SetPreviewError(err string) {
	m.Preview.SetContent("Preview failed:\n\n" + err)
}

func (m *Model) SetPreviewBody(body string) { m.PreviewBody = body }

func (m *Model) ClearStaleLoadingIfIdle() { m.clearStaleLoadingIfIdle() }

// clearStaleLoadingIfIdle resets status after non-scan work (e.g. inspect file preview)
// when no panel scan batch is in flight.
func (m *Model) clearStaleLoadingIfIdle() {
	if m.scan.Pending == 0 && m.status == "loading" {
		m.status = "ready"
		m.updateFooterForState(m.state)
	}
}

func (m *Model) SpinnerTick(msg spinner.TickMsg) tea.Cmd {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m *Model) InstallWizardSearching() bool {
	return m.state == stateInstalling && m.install.WizardOpen() && m.install.WizardSearching()
}

func (m *Model) InstallHandleUIMsg(msg tea.Msg) tea.Cmd {
	_, cmd := m.install.HandleUIMsg(msg)
	return cmd
}

func (m *Model) InstallHandleBackgroundFrame(msg progress.FrameMsg) (tea.Cmd, bool) {
	return m.install.HandleBackgroundFrame(msg)
}

func (m *Model) SelectMCPByName(name string) bool { return applist.SelectMCPByName(m, name) }

func (m *Model) SelectSkillByName(name string) bool { return applist.SelectSkillByName(m, name) }

func (m *Model) MainFallthrough(msg tea.Msg) (tea.Cmd, tea.Cmd) {
	var listCmd, previewCmd tea.Cmd
	m.Main, listCmd = m.Main.Update(msg)
	m.Preview, previewCmd = m.Preview.Update(msg)
	return listCmd, previewCmd
}

func (m *Model) SyncSelectionPreview() tea.Cmd {
	return applist.SyncSelectionPreview(m)
}

func (m *Model) PreviewFileCmd(path string) tea.Cmd {
	return applist.PreviewFileCmd(m, path)
}

func (m *Model) PreviewUpdate(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.Preview, cmd = m.Preview.Update(msg)
	return cmd
}

func (m *Model) TreeUpdate(msg tea.Msg) (applist.FileTree, tea.Cmd) {
	var cmd tea.Cmd
	m.Tree, cmd = m.Tree.Update(msg)
	return m.Tree, cmd
}

func (m *Model) TreeSelected() applist.TreeNode {
	return m.Tree.SelectedNode()
}

func (m *Model) handleAgentFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statefiltering.HandleKeys(m, msg)
}

func (m *Model) LastState() session.State { return m.lastState }

func (m *Model) AgentSelectedItem() (panel.Item, bool) {
	item, ok := m.Agent.SelectedItem().(panel.Item)
	return item, ok
}

func (m *Model) ApplyAgentFilter(id string) { statelisting.SetAgentFilter(m, id) }

func (m *Model) AgentDisplay() string { return m.agentDisplay() }

func (m *Model) AgentFilterListUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.Agent, cmd = m.Agent.Update(msg)
	return cmd
}

func (m *Model) handleListFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return statelistfilter.HandleKeys(m, msg)
}

func (m *Model) ListFilterStatusLine() string {
	n := visiblePanelListCount(m.Main.VisibleItems())
	if m.Main.FilterValue() != "" {
		return fmt.Sprintf("filter %q → %d item(s)", m.Main.FilterValue(), n)
	}
	return fmt.Sprintf("%d item(s)", n)
}

// listFilterActive reports whether inline main-list filtering should consume keys.
func (m *Model) listFilterActive() bool {
	if m.state == stateInstalling || m.state == stateBindingAgent ||
		m.state == stateFilteringAgent || m.state == stateConfirming ||
		m.state == stateInspecting || m.state == stateCommandPalette || m.prompt.Active() {
		return false
	}
	return m.Main.FilterState() == list.Filtering
}

func (m *Model) handleInstallingUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return stateinstalling.HandleKeys(m, msg)
}

func (m *Model) InstallWizardOpen() bool {
	return m.install.WizardOpen()
}

func (m *Model) InstallWizardHandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.install.HandleUIMsg(msg)
}

var (
	_ statelisting.Host       = (*Model)(nil)
	_ statelisting.PromptHost = (*Model)(nil)
	_ statefallback.Host      = (*Model)(nil)
	_ statenspect.Host        = (*Model)(nil)
	_ statefiltering.Host     = (*Model)(nil)
	_ statelistfilter.Host    = (*Model)(nil)
	_ stateinstalling.Host    = (*Model)(nil)
)
