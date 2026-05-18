package app

import (
	"github.com/charmbracelet/bubbles/list"

	"github.com/JoeHe0x/skill-man/internal/app/panel"
	"github.com/JoeHe0x/skill-man/internal/commands"
	mcpdomain "github.com/JoeHe0x/skill-man/internal/domain/mcp"
	skilldomain "github.com/JoeHe0x/skill-man/internal/domain/skill"
)

type itemKind int

const (
	itemKindCommand itemKind = iota
	itemKindSkill
	itemKindMCP
	itemKindMessage
)

type listItem struct {
	kind        itemKind
	title       string
	desc        string
	meta        string
	detailLines []string
	command     commands.Spec
	skill       *skilldomain.Skill
	mcp         *mcpdomain.Server
	bindChecked bool // desired bind state in agent binding UI
	bindInitial bool // bound before entering bind UI
}

func (i listItem) FilterValue() string {
	return listItemToPanel(i).FilterValue()
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }

func commandListItems(specs []commands.Spec) []list.Item {
	return panelToListItems(panel.CommandItems(specs))
}
