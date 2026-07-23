// Package headless (internal test) — tests for the shared detectAgents
// helper extracted from RunWindowsDetect (windows-contract-hub-operations,
// Task 2.3 RED). Internal (package headless, not headless_test) because
// detectAgents is unexported.
package headless

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// stubRegistry is a minimal install.Registry for detect-filter tests.
type stubRegistry struct {
	supported map[model.Agent]bool
}

func (r *stubRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	if r.supported[agent] {
		return nil, true // adapter value not used by RunWindowsDetect
	}
	return nil, false
}

// TestDetectAgents_MatchesRunWindowsDetect asserts that the extracted
// detectAgents(homeDir) helper returns the same agent set that
// RunWindowsDetect emits when the registry is nil (no filtering).
func TestDetectAgents_MatchesRunWindowsDetect(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.MkdirAll(homeDir+"/.claude", 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(homeDir+"/.codex", 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}

	got := detectAgents(homeDir)

	var out bytes.Buffer
	if err := RunWindowsDetect(homeDir, nil, &out); err != nil {
		t.Fatalf("RunWindowsDetect() error = %v", err)
	}
	var resp windowsDetectResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal detect json: %v\nbody=%s", err, out.String())
	}

	sort.Strings(got)
	sort.Strings(resp.DetectedAgents)
	if len(got) != len(resp.DetectedAgents) {
		t.Fatalf("detectAgents() = %v, want %v (from RunWindowsDetect)", got, resp.DetectedAgents)
	}
	for i := range got {
		if got[i] != resp.DetectedAgents[i] {
			t.Fatalf("detectAgents() = %v, want %v (from RunWindowsDetect)", got, resp.DetectedAgents)
		}
	}
	if len(got) == 0 {
		t.Fatal("expected at least one detected agent")
	}
}

// TestRunWindowsDetect_FiltersUnsupportedAgents asserts that RunWindowsDetect
// with a non-nil registry only returns agents that are BOTH detected AND
// registered as adapters — mirroring AvailableAgentsList in the TUI and
// preventing "no adapter registered for agent X" failures in BuildPlan.
func TestRunWindowsDetect_FiltersUnsupportedAgents(t *testing.T) {
	homeDir := t.TempDir()
	// Simulate claude + codex installed on the machine.
	if err := os.MkdirAll(homeDir+"/.claude", 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(homeDir+"/.codex", 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}

	// Registry only supports claude — codex has no adapter.
	reg := &stubRegistry{supported: map[model.Agent]bool{model.AgentClaude: true}}

	var out bytes.Buffer
	if err := RunWindowsDetect(homeDir, reg, &out); err != nil {
		t.Fatalf("RunWindowsDetect() error = %v", err)
	}
	var resp windowsDetectResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal detect json: %v\nbody=%s", err, out.String())
	}

	// Only claude should appear; codex is detected but has no registered adapter.
	if len(resp.DetectedAgents) != 1 || resp.DetectedAgents[0] != "claude" {
		t.Fatalf("detected_agents = %v, want [claude]", resp.DetectedAgents)
	}
}

// TestRunWindowsDetect_AllUnsupported asserts that when NO detected agents
// have registered adapters, RunWindowsDetect returns an empty list
// (not an error), leaving agent selection up to the UI.
func TestRunWindowsDetect_AllUnsupported(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.MkdirAll(homeDir+"/.codex", 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}

	// Registry has no adapters at all.
	reg := &stubRegistry{supported: map[model.Agent]bool{}}

	var out bytes.Buffer
	if err := RunWindowsDetect(homeDir, reg, &out); err != nil {
		t.Fatalf("RunWindowsDetect() error = %v", err)
	}
	var resp windowsDetectResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal detect json: %v\nbody=%s", err, out.String())
	}

	if len(resp.DetectedAgents) != 0 {
		t.Fatalf("detected_agents = %v, want []", resp.DetectedAgents)
	}
}
