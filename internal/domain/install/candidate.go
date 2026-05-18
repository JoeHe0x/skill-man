package install

// Kind identifies which extension type is being installed.
type Kind string

const (
	KindSkill Kind = "skill"
	KindMCP   Kind = "mcp"
)

// Candidate is a searchable install target from a registry or local path.
type Candidate struct {
	Source   string // installer source id, e.g. owner/repo@skill or local path
	Name     string // display name
	Installs string // human-readable install count (registry only)
	URL      string
	Local    bool
}
