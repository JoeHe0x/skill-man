package panel

// Tab identifies an extension category shown in the left panel.
type Tab int

const (
	TabSkills Tab = iota
	TabMCP
)

// String returns the human-readable tab label.
func (t Tab) String() string {
	switch t {
	case TabMCP:
		return "MCP"
	default:
		return "Skills"
	}
}

// Next returns the next tab in cycle order.
func (t Tab) Next() Tab {
	if t == TabSkills {
		return TabMCP
	}
	return TabSkills
}

// Prev returns the previous tab in cycle order.
func (t Tab) Prev() Tab {
	if t == TabMCP {
		return TabSkills
	}
	return TabMCP
}
