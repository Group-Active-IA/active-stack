package planner

import (
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type dependencyResolver struct{}

// NewResolver returns the default Resolver implementation.
func NewResolver() Resolver {
	return dependencyResolver{}
}

// Resolve expands transitive dependencies for every harness in selected, using
// all as the catalog lookup index. It returns a topologically sorted plan
// with a list of IDs that were auto-added as dependencies.
func (r dependencyResolver) Resolve(selected []model.Harness, all map[string]model.Harness) (ResolvedPlan, error) {
	selectedSet := make(map[string]struct{}, len(selected))
	for _, h := range selected {
		selectedSet[h.ID] = struct{}{}
	}

	// Build the transitive dependency map via DFS.
	deps := make(map[string][]string)
	for _, h := range selected {
		if err := expandDependencies(h.ID, all, deps); err != nil {
			return ResolvedPlan{}, err
		}
	}

	ordered, err := TopologicalSort(deps)
	if err != nil {
		return ResolvedPlan{}, err
	}

	var added []string
	for _, id := range ordered {
		if _, explicit := selectedSet[id]; !explicit {
			added = append(added, id)
		}
	}

	return ResolvedPlan{
		OrderedIDs: ordered,
		AddedIDs:   added,
	}, nil
}

// expandDependencies recursively adds id and all its transitive dependencies
// to the deps map. Returns an error if a referenced dep is not in all.
func expandDependencies(id string, all map[string]model.Harness, deps map[string][]string) error {
	if _, visited := deps[id]; visited {
		return nil
	}

	h, ok := all[id]
	if !ok {
		return fmt.Errorf("planner: unknown harness %q", id)
	}

	// Register with its direct dependencies.
	directDeps := make([]string, len(h.DependsOn))
	copy(directDeps, h.DependsOn)
	deps[id] = directDeps

	for _, depID := range h.DependsOn {
		if _, ok := all[depID]; !ok {
			return fmt.Errorf("planner: harness %q depends on unknown harness %q", id, depID)
		}
		if err := expandDependencies(depID, all, deps); err != nil {
			return err
		}
	}

	return nil
}
