package model

// tierCapableAgents is the single source of truth for agents that can
// meaningfully differentiate the three permission tiers (claude, opencode,
// codex). Both the TUI permission screen and the windows options command
// consume this via TierCapable so the two paths cannot diverge.
var tierCapableAgents = map[Agent]bool{
	AgentClaude:   true,
	AgentOpenCode: true,
	AgentCodex:    true,
}

// TierCapable returns true when at least one agent in the set is tier-capable
// (i.e. can meaningfully differentiate estricto/balanceado/bypass). It is a
// pure function of the agent set — deterministic and testable in both
// directions.
func TierCapable(agents []Agent) bool {
	for _, a := range agents {
		if tierCapableAgents[a] {
			return true
		}
	}
	return false
}
