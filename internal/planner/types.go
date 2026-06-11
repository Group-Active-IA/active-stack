// Package planner resolves harness dependency graphs, produces a deterministic
// topological install order, and builds the review payload shown to the user
// before installation begins.
package planner

import "github.com/Group-Active-IA/active-stack/internal/model"

// Resolver resolves a harness selection into an ordered install plan.
type Resolver interface {
	// Resolve takes the set of explicitly selected harnesses and the full
	// catalog index (id → Harness) for dependency look-ups. It expands
	// transitive dependencies, validates that all declared deps exist, and
	// returns a plan with a deterministic topological order.
	Resolve(selected []model.Harness, all map[string]model.Harness) (ResolvedPlan, error)
}

// ResolvedPlan is the output of a successful Resolve call.
type ResolvedPlan struct {
	// OrderedIDs is the full set of harness IDs to install, in topological
	// order (dependencies first).
	OrderedIDs []string

	// AddedIDs contains the IDs of harnesses that were pulled in automatically
	// as transitive dependencies but were not in the original selection.
	AddedIDs []string
}

// ReviewPayload is the data shown to the user on the confirmation screen
// before installation starts.
type ReviewPayload struct {
	// HarnessActions lists every harness to be installed with its action label.
	HarnessActions []HarnessAction

	// AddedIDs mirrors ResolvedPlan.AddedIDs for convenience.
	AddedIDs []string

	// Agents is the set of agents the installation targets.
	Agents []model.Agent

	// Mode is the install mode chosen by the user.
	Mode model.InstallMode
}

// HarnessAction pairs a harness ID with a human-readable action label.
type HarnessAction struct {
	// ID is the harness identifier.
	ID string

	// Action is either "selected" (user explicitly chose it) or
	// "auto-dependency" (pulled in to satisfy a DependsOn constraint).
	Action string
}
