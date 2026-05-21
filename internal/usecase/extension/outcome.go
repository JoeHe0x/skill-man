package extension

// Kind identifies which extension tab should receive post-mutation reselection.
type Kind int

const (
	KindSkill Kind = iota
	KindMCP
)

// Outcome is the result of an extension mutation use case (no UI dependencies).
type Outcome struct {
	Kind         Kind
	AffectedName string
	Message      string
	Err          error
}
