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
	for _, id := range []string{"engram-cloud", "made-up-config"} {
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
