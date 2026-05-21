package panel

import (
	"context"

	"github.com/JoeHe0x/skill-man/internal/domain/agent"
)

// Core drives list content, scanning, and preview for one extension tab (no Bubble Tea).
type Core interface {
	Tab() Tab
	Count() int
	CountLabel() string
	Capabilities() Capabilities

	Scan(ctx context.Context, cwd, home string, agents []agent.Agent) ScannedMsg
	ApplyScan(msg ScannedMsg) bool

	ListItems(agentFilter []string) []Item
	SearchItems(query string, agentFilter []string) []Item

	PanelTitle(state ViewState) string
	ReloadHint() string
	StaticPreview() string
	PreviewMarkdown(selected Item, width int) (string, error)

	SelectedSkill(item Item) bool
	SelectedMCP(item Item) bool
}

// Panel is the extension tab contract used by the app (framework-agnostic core).
type Panel = Core
