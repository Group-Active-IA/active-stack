package uninstall

import (
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// EngramCloudHarnessID is the harness whose removal also removes our memory
// read hooks and ENGRAM_CLOUD_AUTOSYNC.
const EngramCloudHarnessID = "engram-cloud"

// mcpPather is the optional part of an adapter that locates the MCP file.
// Adapters without it (or agents outside Claude and Codex) are skipped.
type mcpPather interface {
	MCPConfigPath(homeDir, serverName string) string
}

type autosyncAdapter struct {
	AgentAdapter
	home string
}

func (a autosyncAdapter) MCPConfigPath(home, server string) string {
	if p, ok := a.AgentAdapter.(mcpPather); ok {
		return p.MCPConfigPath(home, server)
	}
	return ""
}

func cloudAgents(adapters []AgentAdapter) []AgentAdapter {
	var out []AgentAdapter
	for _, a := range adapters {
		if a.Agent() == model.AgentClaude || a.Agent() == model.AgentCodex {
			out = append(out, a)
		}
	}
	return out
}

var hooksRemoveFn = func(home string, adapters []hooks.Adapter) (hooks.Result, error) {
	return hooks.Remove(home, adapters)
}

var autosyncRemoveFn = func(home string, adapters []autosync.Adapter) (autosync.Result, error) {
	return autosync.Remove(home, adapters)
}

// cloudRemovalStep removes only what the engram-cloud install added at user
// level: the memory read hooks and ENGRAM_CLOUD_AUTOSYNC. It never touches the
// orchestrator, design-mockups, other tools' hooks, the user's text, or
// anything under .engram/ (config.json, active-stack.json, memories): those
// belong to the project and the person.
type cloudRemovalStep struct {
	adapters []AgentAdapter
	homeDir  string
	manifest *backup.Manifest
}

func (s *cloudRemovalStep) ID() string                     { return "cloud:" + EngramCloudHarnessID }
func (s *cloudRemovalStep) setManifest(m *backup.Manifest) { s.manifest = m }

func (s *cloudRemovalStep) Run() error {
	ca := cloudAgents(s.adapters)
	ha := make([]hooks.Adapter, len(ca))
	aa := make([]autosync.Adapter, len(ca))
	for i, a := range ca {
		ha[i] = a
		aa[i] = autosyncAdapter{AgentAdapter: a, home: s.homeDir}
	}
	if _, err := hooksRemoveFn(s.homeDir, ha); err != nil {
		return fmt.Errorf("remove memory hooks: %w", err)
	}
	if _, err := autosyncRemoveFn(s.homeDir, aa); err != nil {
		return fmt.Errorf("remove %s: %w", autosync.EnvKey, err)
	}
	return nil
}

func (s *cloudRemovalStep) Rollback() error {
	if s.manifest == nil {
		return nil
	}
	return restoreFn(*s.manifest)
}
