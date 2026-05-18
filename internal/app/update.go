package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"skill-man/internal/domain"
	"skill-man/internal/service"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()
		return m, m.syncSelectionPreview()

	case tea.KeyMsg:
		if m.prompt != nil {
			return m.handlePromptKeys(msg)
		}
		if m.state == stateConfirming {
			return m.handleConfirmKeys(msg)
		}
		if m.state == stateBindingAgent {
			return m.handleBindingKeys(msg)
		}
		if m.state == stateInspecting {
			return m.handleInspectingKeys(msg)
		}
		return m.handleKeyMsg(msg)

	case skillsScannedMsg:
		if msg.err != nil {
			m.status = "error"
			m.errMsg = msg.err.Error()
			m.logf("scan failed: %v", msg.err)
			return m, nil
		}
		m.skills = msg.skills
		m.status = "ready"
		m.errMsg = ""
		m.logf("scanned %d skill(s)", len(msg.skills))
		m.setSkillItems(m.skills)
		if m.state == stateHome || m.state == stateListing || m.state == stateSearching {
			m.setSkillItems(m.skills)
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case previewLoadedMsg:
		if msg.gen != m.previewGen {
			return m, nil // stale, drop
		}
		if msg.err != nil {
			m.preview.SetContent("Preview failed:\n\n" + msg.err.Error())
			return m, nil
		}
		m.previewBody = msg.content
		m.preview.SetContent(msg.content)
		return m, nil

	case mutationCompletedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.status = "error"
			m.logf("mutation failed: %v", msg.err)
			return m, m.scanSkillsCmd()
		}
		m.errMsg = ""
		m.status = "ready"
		if msg.message != "" {
			m.hint = msg.message
			m.logf("%s", msg.message)
		}
		if msg.selectName != "" {
			return m, tea.Sequence(
				m.scanSkillsCmd(),
				func() tea.Msg { return reselectSkillMsg{name: msg.selectName} },
			)
		}
		return m, m.scanSkillsCmd()

	case reselectSkillMsg:
		if m.selectSkillByName(msg.name) {
			m.hint = fmt.Sprintf("selected %s", msg.name)
			return m, m.syncSelectionPreview()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var (
		listCmd    tea.Cmd
		previewCmd tea.Cmd
	)
	m.list, listCmd = m.list.Update(msg)
	m.preview, previewCmd = m.preview.Update(msg)

	m.updateHint()

	return m, tea.Batch(listCmd, previewCmd)
}

// --- key dispatch -----------------------------------------------------------

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Home):
		m.errMsg = ""
		m.state = stateHome
		m.lastState = stateHome
		m.setSkillItems(m.skills)

		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.Help):
		return m.handleHelp()

	case key.Matches(msg, keys.Down):
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.Up):
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, tea.Batch(cmd, m.syncSelectionPreview())

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd

	case key.Matches(msg, keys.List):
		return m.handleList()

	case key.Matches(msg, keys.Find):
		return m.showFindPrompt()

	case key.Matches(msg, keys.Agent):
		return m.handleCycleAgent()

	case key.Matches(msg, keys.Reload):
		return m.handleReload()

	case key.Matches(msg, keys.Update):
		return m.handleUpdate()

	case key.Matches(msg, keys.Enter):
		return m.handleInspectSelected()

	case key.Matches(msg, keys.Bind):
		return m.handleBindSelected()

	case key.Matches(msg, keys.Disable):
		return m.handleDisableSelected()

	case key.Matches(msg, keys.Delete):
		return m.handleRemoveSelected()
	}

	return m, nil
}

// --- command handlers -------------------------------------------------------

func (m *Model) handleHelp() (tea.Model, tea.Cmd) {
	m.state = stateViewingHelp
	m.lastState = stateViewingHelp
	m.setCommandItems()
	m.preview.SetContent(helpPreview)
	m.hint = "Ctrl+A:cycle agent  Ctrl+L:list  Ctrl+R:reload  Esc:back"
	return m, nil
}

func (m *Model) handleList() (tea.Model, tea.Cmd) {
	m.state = stateListing
	m.lastState = stateListing
	m.hint = fmt.Sprintf("%d skill(s) | agents: %s | %s", len(m.skills), m.agentDisplay(), m.cwd)
	m.setSkillItems(m.skills)
	return m, m.syncSelectionPreview()
}

func (m *Model) handleReload() (tea.Model, tea.Cmd) {
	m.status = "loading"
	m.hint = "Rescanning local skills..."
	return m, m.scanSkillsCmd()
}

func (m *Model) handleInspectSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		m.lastState = m.state
		m.state = stateInspecting
		m.tree.setRoot(selected.skill.Path)
		m.hint = "Space/Enter: toggle folder | Esc: back to skills"

		// load preview for the selected tree item if it's a file
		sel := m.tree.SelectedItem()
		if sel.path != "" && !sel.isDir {
			m.status = "loading"
			return m, m.previewFileCmd(sel.path)
		}
	}
	return m, nil
}

func (m *Model) handleDisableSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok || selected.kind != itemKindSkill {
		m.hint = "Select a skill first, then press 'x' to toggle disable"
		return m, nil
	}
	skill := selected.skill
	m.status = "loading"
	action := "Disabling"
	if skill.Disabled {
		action = "Enabling"
	}
	m.hint = fmt.Sprintf("%s %s...", action, skill.Name)
	return m, m.toggleDisableSkillCmd(skill)
}

func (m *Model) handleRemoveSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok || selected.kind != itemKindSkill {
		m.hint = "Select a skill first, then press Delete to remove"
		return m, nil
	}
	skill := selected.skill
	m.pending = &pendingAction{name: "remove", skillName: skill.Name, skill: skill}
	m.lastState = m.state
	m.state = stateConfirming
	return m, nil
}

func (m *Model) handleUpdate() (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(listItem)
	if ok && selected.kind == itemKindSkill {
		skill := selected.skill
		m.status = "loading"
		m.hint = fmt.Sprintf("Updating %s...", skill.Name)
		return m, m.updateSkillCmd(skill)
	}
	m.status = "loading"
	m.hint = "Updating all managed local skills..."
	return m, m.updateAllSkillsCmd()
}

func (m *Model) handleCycleAgent() (tea.Model, tea.Cmd) {
	allIDs := []string{"all"}
	for _, a := range m.allAgents {
		allIDs = append(allIDs, a.ID)
	}
	current := "all"
	if len(m.agentIDs) > 0 {
		current = m.agentIDs[0]
	}
	for i, id := range allIDs {
		if strings.EqualFold(id, current) {
			next := allIDs[(i+1)%len(allIDs)]
			m.setAgentFilter(next)
			break
		}
	}
	m.hint = fmt.Sprintf("Agent filter: %s", m.agentDisplay())
	m.setSkillItems(m.skills)
	return m, m.syncSelectionPreview()
}

// --- prompt handlers --------------------------------------------------------

func (m *Model) showFindPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Find", "search query...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		text = strings.TrimSpace(text)
		m.state = stateSearching
		m.lastState = stateSearching
		if text == "" {
			m.hint = "Search cancelled"
			m.setSkillItems(m.skills)
			return m.syncSelectionPreview()
		}
		results := m.searchSkills(text)
		m.hint = fmt.Sprintf("find: %q -> %d result(s)", text, len(results))
		m.setSkillItems(results)
		return m.syncSelectionPreview()
	})
}

func (m *Model) showAddPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Add source", "path or SKILL.md ...", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		source := strings.TrimSpace(text)
		if source == "" {
			m.hint = "Add cancelled"
			return nil
		}
		m.status = "loading"
		m.hint = fmt.Sprintf("Installing from %s...", source)
		return m.addSkillCmd(source)
	})
}

func (m *Model) showInitPrompt() (tea.Model, tea.Cmd) {
	return m, m.showPrompt("Init name", "new-skill (enter for default)", func(m *Model, text string) tea.Cmd {
		m.hidePrompt()
		name := strings.TrimSpace(text)
		if name == "" {
			name = "new-skill"
		}
		m.status = "loading"
		m.hint = fmt.Sprintf("Creating skill template: %s", name)
		return m.initSkillCmd(name)
	})
}

func (m *Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.hidePrompt()
		m.hint = "Cancelled"
		return m, nil
	case key.Matches(msg, keys.Enter):
		text := strings.TrimSpace(m.prompt.input.Value())
		cmd := m.prompt.action(m, text)
		return m, cmd
	}
	var cmd tea.Cmd
	m.prompt.input, cmd = m.prompt.input.Update(msg)
	return m, cmd
}

// --- confirmation -----------------------------------------------------------

func (m *Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Confirm):
		if m.pending != nil && m.pending.name == "remove" {
			skill := m.pending.skill
			m.pending = nil
			m.state = m.lastState
			m.status = "loading"
			m.hint = fmt.Sprintf("Removing %s...", skill.Name)
			return m, m.removeSkillCmd(skill)
		}
		m.pending = nil
		m.state = m.lastState
		return m, nil
	case key.Matches(msg, keys.Cancel):
		m.pending = nil
		m.state = m.lastState
		m.logf("action cancelled")
		return m, nil
	default:
		return m, nil
	}
}

// --- helpers ----------------------------------------------------------------

func (m *Model) setAgentFilter(id string) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" || id == "all" {
		m.agentIDs = []string{"all"}
		return
	}
	if _, ok := domain.AgentByID(id); ok {
		m.agentIDs = []string{id}
		return
	}
	m.logf("unknown agent: %s", id)
}

func (m *Model) searchSkills(query string) []domain.Skill {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return m.skills
	}

	var results []domain.Skill
	for _, skill := range m.skills {
		haystack := strings.ToLower(strings.Join([]string{
			skill.Name,
			skill.Description,
			strings.Join(skill.Tools, " "),
			skill.Path,
		}, " "))
		if strings.Contains(haystack, query) {
			results = append(results, skill)
		}
	}
	return results
}

const helpPreview = `# skill-man

Keybindings:

- ? / F1: show this help
- Enter: inspect skill (open file tree)
- x: toggle disable/enable for selected skill
- b: bind/unbind skill to specific agents
- Delete: remove selected skill (with confirmation)
- Ctrl+L: list skills
- Ctrl+F: find skills (prompt)
- Ctrl+A: cycle agent filter
- Ctrl+D: add/install skill (prompt)
- Ctrl+N: create new skill template (prompt)
- Ctrl+R: reload/rescan skills
- Ctrl+U: update skill (selected or all)
- Ctrl+J / Down, Ctrl+K / Up: navigate list
- PgUp / PgDn: scroll preview
- Esc: home / cancel
- Ctrl+C: quit

Prompts appear at the bottom for commands that need text input.
Press Enter to confirm, Esc to cancel.`

func (m *Model) handleBindSelected() (tea.Model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(listItem)
	if !ok || selected.kind != itemKindSkill {
		m.hint = "Select a skill first to manage agent bindings"
		return m, nil
	}
	m.bindingSkill = selected.skill
	m.lastState = m.state
	m.state = stateBindingAgent
	m.hint = "Space: toggle binding | Enter: apply | Esc: cancel"

	// Prepare agentList items, grouped by SkillsDir
	groups := make(map[string][]domain.Agent)
	var dirs []string
	for _, a := range m.allAgents {
		if _, exists := groups[a.EntityDirs[domain.EntitySkill]]; !exists {
			dirs = append(dirs, a.EntityDirs[domain.EntitySkill])
		}
		groups[a.EntityDirs[domain.EntitySkill]] = append(groups[a.EntityDirs[domain.EntitySkill]], a)
	}

	var items []list.Item
	for _, dir := range dirs {
		group := groups[dir]
		var names []string
		for _, a := range group {
			names = append(names, a.Name)
		}

		// Check if it's bound
		bound := false
		for _, a := range group {
			for _, id := range selected.skill.Agents {
				if id == a.ID {
					bound = true
					break
				}
			}
			if bound {
				break
			}
		}

		title := "[ ] " + strings.Join(names, ", ")
		if bound {
			title = "[✅] " + strings.Join(names, ", ")
		}

		items = append(items, listItem{
			kind:  itemKindMessage,
			title: title,
			desc:  dir,
			meta:  group[0].ID, // Use the first agent's ID as representative
		})
	}
	m.agentList.SetItems(items)
	return m, nil
}

func (m *Model) handleBindingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		// Apply bindings
		for _, item := range m.agentList.Items() {
			li := item.(listItem)
			bound := strings.HasPrefix(li.title, "[✅]")
			agentID := li.meta
			agent, ok := domain.AgentByID(agentID)
			if !ok {
				continue
			}

			// We need a tea.Cmd for each operation to do it in the background, or do it synchronously.
			// Since it's fast (symlink), we can just do it synchronously for now.
			var err error
			if bound {
				err = service.BindAgent(m.bindingSkill, agent, m.cwd, m.home)
			} else {
				err = service.UnbindAgent(m.bindingSkill, agent, m.cwd, m.home)
			}
			if err != nil {
				m.errMsg = err.Error()
			}
		}
		m.state = m.lastState
		if m.errMsg == "" {
			m.hint = fmt.Sprintf("Updated agent bindings for %s", m.bindingSkill.Name)
		} else {
			m.status = "error"
		}
		return m, tea.Sequence(m.scanSkillsCmd(), func() tea.Msg { return reselectSkillMsg{name: m.bindingSkill.Name} })

	case key.Matches(msg, keys.Cancel):
		m.state = m.lastState
		m.hint = "Agent binding cancelled"
		return m, nil

	case key.Matches(msg, keys.Toggle):
		idx := m.agentList.Index()
		items := m.agentList.Items()
		if idx >= 0 && idx < len(items) {
			li := items[idx].(listItem)
			if strings.HasPrefix(li.title, "[✅]") {
				li.title = "[ ]" + strings.TrimPrefix(li.title, "[✅]")
			} else {
				li.title = "[✅]" + strings.TrimPrefix(li.title, "[ ]")
			}
			items[idx] = li
			cmd := m.agentList.SetItems(items)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.agentList, cmd = m.agentList.Update(msg)
	return m, cmd
}

func (m *Model) handleInspectingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Home):
		m.state = m.lastState
		m.hint = "Returned to skill list"
		return m, m.syncSelectionPreview()

	case key.Matches(msg, keys.PgDown, keys.PgUp):
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd
	}

	oldSelected := m.tree.SelectedItem()
	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	newSelected := m.tree.SelectedItem()

	if newSelected.path != "" && newSelected.path != oldSelected.path && !newSelected.isDir {
		// load preview for the file
		m.status = "loading"
		return m, tea.Batch(cmd, m.previewFileCmd(newSelected.path))
	}

	return m, cmd
}

func (m *Model) previewFileCmd(path string) tea.Cmd {
	width := m.preview.Width
	if width == 0 {
		width = max(40, m.width/2)
	}
	m.previewGen++
	gen := m.previewGen
	return func() tea.Msg {
		// Reuse RenderSkillPreview but with a dummy skill that points to this file
		dummy := domain.Skill{
			Name:          filepath.Base(path),
			SkillFilePath: path,
		}
		content, err := service.RenderSkillPreview(dummy, width)
		return previewLoadedMsg{content: content, err: err, gen: gen}
	}
}
