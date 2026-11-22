package proxyreg

import (
	"errors"
	"fmt"
	"os"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// AgentStatus reports, for one agent, whether the Engram MCP entry runs
// through the proxy.
type AgentStatus struct {
	Agent model.Agent
	File  string
	// HasEntry is false when the file does not exist or has no Engram entry:
	// there is nothing to proxy.
	HasEntry bool
	// Registered is true when the entry runs through the proxy.
	Registered bool
}

// Status reports, without writing anything, whether each Claude and Codex
// Engram entry runs through the proxy. Other agents are skipped.
func Status(home string, adapters []Adapter) ([]AgentStatus, error) {
	var out []AgentStatus
	for _, a := range adapters {
		path := a.MCPConfigPath(home, ServerName)
		if path == "" || (a.Agent() != model.AgentClaude && a.Agent() != model.AgentCodex) {
			continue
		}
		st := AgentStatus{Agent: a.Agent(), File: path}
		raw, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			out = append(out, st)
			continue
		}
		if err != nil {
			return out, fmt.Errorf("proxyreg: %s (agent %q): %w", path, a.Agent(), err)
		}
		if a.Agent() == model.AgentClaude {
			e, lerr := locateClaude(raw, ServerName)
			switch {
			case errors.Is(lerr, ErrEntryNotFound):
			case lerr != nil:
				return out, fmt.Errorf("proxyreg: %s (agent %q): %w", path, a.Agent(), lerr)
			default:
				st.HasEntry, st.Registered = true, isProxied(raw, e)
			}
		} else {
			lines := splitLines(string(raw))
			_, _, found, ferr := findBlock(lines)
			if ferr != nil {
				return out, fmt.Errorf("proxyreg: %s (agent %q): %w", path, a.Agent(), ferr)
			}
			if found {
				st.HasEntry, st.Registered = true, true
			} else {
				for _, l := range lines {
					if isServerHeader(l.text, ServerName) {
						st.HasEntry = true
						break
					}
				}
			}
		}
		out = append(out, st)
	}
	return out, nil
}
