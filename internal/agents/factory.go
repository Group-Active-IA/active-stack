package agents

import (
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/agents/opencode"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// NewAdapter returns a concrete Adapter for the given agent.
// Supported scope: claude, codex, and opencode.
// Returns AgentNotSupportedError for any other agent.
func NewAdapter(agent model.Agent) (Adapter, error) {
	switch agent {
	case model.AgentClaude:
		return claude.NewAdapter(), nil
	case model.AgentOpenCode:
		return opencode.NewAdapter(), nil
	case model.AgentCodex:
		return codex.NewAdapter(), nil
	default:
		return nil, AgentNotSupportedError{Agent: agent}
	}
}

// NewDefaultRegistry returns a Registry pre-populated with exactly the P0
// agents: claude, codex, and opencode.
// Remaining agents (gemini, cursor, vscode, windsurf, antigravity) are
// NOT included — add them by creating a sub-package + one entry here.
func NewDefaultRegistry() (*Registry, error) {
	r := NewRegistry()
	for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex, model.AgentOpenCode} {
		adapter, err := NewAdapter(agent)
		if err != nil {
			return nil, fmt.Errorf("create %s adapter: %w", agent, err)
		}
		if err := r.Register(adapter); err != nil {
			return nil, fmt.Errorf("register %s adapter: %w", agent, err)
		}
	}
	return r, nil
}
