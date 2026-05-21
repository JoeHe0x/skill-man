// Package session defines TUI session states and valid transitions.
package session

// State is the Bubble Tea session mode (listing, installing, binding, …).
type State int

const (
	Home State = iota
	Listing
	Searching
	Installing
	Confirming
	HelpOverlay
	BindingAgent
	FilteringAgent
	Inspecting
	CommandPalette
)

var allowedTransitions = map[State][]State{
	Home:           {Listing, Searching, Installing, BindingAgent, Inspecting, Confirming, HelpOverlay, CommandPalette, FilteringAgent},
	Listing:        {Home, Searching, Installing, BindingAgent, Inspecting, Confirming, HelpOverlay, CommandPalette, FilteringAgent},
	Searching:      {Listing, Home},
	Installing:     {Listing, Home},
	Confirming:     {Listing, Home},
	HelpOverlay:    {Listing, Home},
	BindingAgent:   {Listing, Home},
	FilteringAgent: {Listing, Home},
	Inspecting:     {Listing, Home},
	CommandPalette: {Listing, Home},
}

// CanTransition reports whether moving from → to is allowed.
func CanTransition(from, to State) bool {
	if from == to {
		return true
	}
	allowed, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
