package tui

import (
	"os"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// DetectInstalledAgents returns the model.Agent values whose known config
// directory exists under homeDir. It checks only P0 agents (claude, opencode).
//
// The mapping from filesystem paths to model.Agent is local to the TUI so
// the system package stays import-cycle-free. When new agents are added to
// the P0 registry, add their entries here.
func DetectInstalledAgents(homeDir string) []model.Agent {
	type entry struct {
		agent model.Agent
		path  string
	}
	candidates := []entry{
		{model.AgentClaude, filepath.Join(homeDir, ".claude")},
		{model.AgentOpenCode, filepath.Join(homeDir, ".config", "opencode")},
	}

	var found []model.Agent
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err == nil {
			found = append(found, c.agent)
		}
	}
	return found
}
