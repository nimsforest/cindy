// Package cindy provides types and validation for the Cindy agentic CI/CD protocol.
//
// Cindy is a label-based deployment protocol where Git branches + labels form
// the state machine. This package provides label constants, transition validation,
// manifest parsing, review types, and schema safety enforcement.
package cindy

// Label represents a Cindy pipeline label.
type Label string

const (
	Ready             Label = "cindy:ready"
	Analyzing         Label = "cindy:analyzing"
	Approved          Label = "cindy:approved"
	Blocked           Label = "cindy:blocked"
	Deploying         Label = "cindy:deploying"
	Deployed          Label = "cindy:deployed"
	Rejected          Label = "cindy:rejected"
	Rollback          Label = "cindy:rollback"
	HumanReview       Label = "cindy:human-review"
	RevisionRequested Label = "cindy:revision-requested"
)

// AllLabels returns all valid Cindy labels.
func AllLabels() []Label {
	return []Label{
		Ready, Analyzing, Approved, Blocked, Deploying,
		Deployed, Rejected, Rollback, HumanReview, RevisionRequested,
	}
}

// validTransitions maps each label to its allowed next states.
var validTransitions = map[Label][]Label{
	Ready:             {Analyzing},
	Analyzing:         {Approved, Rejected, HumanReview, Blocked, RevisionRequested},
	Approved:          {Deploying, Blocked},
	Blocked:           {Approved},
	Deploying:         {Deployed, Rollback},
	HumanReview:       {Approved, Rejected, RevisionRequested},
	RevisionRequested: {Ready},
	Deployed:          {Rollback},
}

// CanTransition returns true if transitioning from one label to another is valid
// according to the Cindy protocol state machine.
func CanTransition(from, to Label) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// ValidTransitionsFrom returns the list of valid target labels from a given label.
// Returns nil if the label has no valid transitions (terminal state or unknown).
func ValidTransitionsFrom(from Label) []Label {
	return validTransitions[from]
}

// IsTerminal returns true if a label has no valid outgoing transitions
// (i.e., Rejected or Rollback with no further transitions defined,
// though Deployed can transition to Rollback).
func IsTerminal(l Label) bool {
	targets := validTransitions[l]
	return len(targets) == 0
}
