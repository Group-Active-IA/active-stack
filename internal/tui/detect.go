package tui

import (
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// DetectInstalledAgents returns the model.Agent values whose known config
// directory exists under homeDir.
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
		{model.AgentCodex, filepath.Join(homeDir, ".codex")},
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
