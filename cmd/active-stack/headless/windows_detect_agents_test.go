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
)

// TestDetectAgents_MatchesRunWindowsDetect asserts that the extracted
// detectAgents(homeDir) helper returns the same agent set that
// RunWindowsDetect emits for a fabricated --home.
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
