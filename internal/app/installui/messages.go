package installui

import (
	domaininstall "github.com/JoeHe0x/skill-man/internal/domain/install"
)

// SearchDoneMsg is sent when a registry search finishes.
type SearchDoneMsg struct {
	Results []domaininstall.Candidate
	Err     error
}

// InstallDoneMsg is sent when an install command finishes.
type InstallDoneMsg struct {
	Name string
	Err  error
}

// ProgressTickMsg drives the indeterminate progress bar while installing.
type ProgressTickMsg struct{}

// ClosedMsg means the user dismissed the wizard.
type ClosedMsg struct {
	Hint string
}

// HintMsg asks the host to show a transient footer hint without closing.
type HintMsg struct {
	Text string
}

// CancelInstallMsg asks the host to abort an in-flight install (second Esc).
type CancelInstallMsg struct{}

// RequestInstallMsg asks the host to run Install with a cancellable context.
type RequestInstallMsg struct {
	AgentIDs []string
}
