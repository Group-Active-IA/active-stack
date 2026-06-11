package planner

import "github.com/Group-Active-IA/active-stack/internal/model"

// Graph is an adjacency-list representation of harness dependencies.
// Keys are harness IDs; values are the IDs of their direct dependencies.
type Graph struct {
	dependencies map[string][]string
}

// NewGraph builds a Graph from a slice of harnesses. Each harness's DependsOn
// field populates its adjacency list.
func NewGraph(harnesses []model.Harness) Graph {
	deps := make(map[string][]string, len(harnesses))
	for _, h := range harnesses {
		copied := make([]string, len(h.DependsOn))
		copy(copied, h.DependsOn)
		deps[h.ID] = copied
	}
	return Graph{dependencies: deps}
}

// Has reports whether id is a known node in the graph.
func (g Graph) Has(id string) bool {
	_, ok := g.dependencies[id]
	return ok
}

// DependenciesOf returns a copy of the direct dependency IDs for id, or nil
// if id is not in the graph.
func (g Graph) DependenciesOf(id string) []string {
	deps, ok := g.dependencies[id]
	if !ok {
		return nil
	}
	copied := make([]string, len(deps))
	copy(copied, deps)
	return copied
}
