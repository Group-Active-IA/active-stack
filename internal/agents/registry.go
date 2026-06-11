package agents

import (
	"fmt"
	"slices"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Registry maps model.Agent identifiers to their concrete Adapter.
// Use NewRegistry to construct and Register to populate.
type Registry struct {
	adapters map[model.Agent]Adapter
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{adapters: make(map[model.Agent]Adapter)}
}

// Register adds adapter to the registry.
// Returns ErrDuplicateAdapter if the same agent is registered more than once.
func (r *Registry) Register(adapter Adapter) error {
	agent := adapter.Agent()
	if _, exists := r.adapters[agent]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateAdapter, agent)
	}
	r.adapters[agent] = adapter
	return nil
}

// Get returns the adapter for the given agent.
// The second return value is false when no adapter is registered for agent.
func (r *Registry) Get(agent model.Agent) (Adapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// SupportedAgents returns the sorted list of agents that have a registered
// adapter in this registry.
func (r *Registry) SupportedAgents() []model.Agent {
	agents := make([]model.Agent, 0, len(r.adapters))
	for a := range r.adapters {
		agents = append(agents, a)
	}
	slices.Sort(agents)
	return agents
}
