package verify_test

import (
	"os"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// R-07 "cierre sin huecos": a config harness without a registered section
// fragment (engram-cloud before phase 2) installs as an explicit no-op, so
// verify must not demand a marker the installer never writes.
func TestChecksForHarness_ConfigWithoutFragment_DoesNotDemandMarker(t *testing.T) {
	orch := "<!-- active-stack:sdd-orchestrator -->\nx\n<!-- /active-stack:sdd-orchestrator -->\n"
	// engram-cloud is no longer here: R-13 registered its fragment, so it is a
	// real section harness verified against its own marker (test below).
	for _, id := range []string{"made-up-config"} {
		for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex} {
			t.Run(id+"/"+string(agent), func(t *testing.T) {
				homeDir := t.TempDir()
				fa := newFakeAdapter(t, homeDir, agent)
				if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte(orch), 0o644); err != nil {
					t.Fatal(err)
				}
				h := model.Harness{ID: id, Type: model.HarnessConfig, Agents: []model.Agent{agent}}
				checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
				for _, r := range runChecks(t, checks) {
					if r.Status == verify.CheckStatusFailed {
						t.Errorf("check %s failed for a no-op harness: %s", r.ID, r.Error)
					}
				}
			})
		}
	}
}

// Triangulation: a registered section harness is still verified against its own marker.
func TestChecksForHarness_RegisteredSectionStillDemandsMarker(t *testing.T) {
	homeDir := t.TempDir()
	fa := newFakeAdapter(t, homeDir, model.AgentClaude)
	orch := "<!-- active-stack:sdd-orchestrator -->\nx\n<!-- /active-stack:sdd-orchestrator -->\n"
	if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte(orch), 0o644); err != nil {
		t.Fatal(err)
	}
	h := model.Harness{ID: "design-mockups", Type: model.HarnessConfig, Agents: []model.Agent{model.AgentClaude}}
	checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
	if len(checks) != 1 {
		t.Fatalf("checks = %d, want 1", len(checks))
	}
	if r := runChecks(t, checks)[0]; r.Status != verify.CheckStatusFailed {
		t.Errorf("design-mockups without its marker must fail, got %q", r.Status)
	}
}

// engram-cloud has a registered fragment since R-13: verify demands its own
// marker exactly once.
func TestChecksForHarness_EngramCloud_DemandsOwnMarker(t *testing.T) {
	orch := "<!-- active-stack:sdd-orchestrator -->\nx\n<!-- /active-stack:sdd-orchestrator -->\n"
	ec := "<!-- active-stack:engram-cloud -->\ny\n<!-- /active-stack:engram-cloud -->\n"
	cases := []struct {
		name     string
		content  string
		wantFail bool
	}{
		{"both sections pass", orch + ec, false},
		{"only orchestrator fails", orch, true},
		{"duplicated section fails", orch + ec + ec, true},
	}
	for _, tc := range cases {
		for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex} {
			t.Run(tc.name+"/"+string(agent), func(t *testing.T) {
				homeDir := t.TempDir()
				fa := newFakeAdapter(t, homeDir, agent)
				if err := os.WriteFile(fa.InstructionsPath(homeDir), []byte(tc.content), 0o644); err != nil {
					t.Fatal(err)
				}
				h := model.Harness{ID: "engram-cloud", Type: model.HarnessConfig, Agents: []model.Agent{agent}}
				// The marker check is the first; the R-21 hook/proxy checks follow
				// and are covered in engram_cloud_checks_test.go.
				checks := verify.ChecksForHarness(h, []verify.Adapter{fa}, homeDir)
				if len(checks) == 0 || checks[0].ID != "config:engram-cloud:"+string(agent) {
					t.Fatalf("first check = %v, want the marker check", checks)
				}
				r := runChecks(t, checks[:1])[0]
				if got := r.Status == verify.CheckStatusFailed; got != tc.wantFail {
					t.Errorf("failed = %v (%s), want %v", got, r.Error, tc.wantFail)
				}
			})
		}
	}
}
