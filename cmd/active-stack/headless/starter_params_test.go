// Package headless — tests for the C2 pre-refactor helper BuildStarterInstallParams
// (windows-contract-hub-operations, Task 1.1 RED).
package headless_test

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestBuildStarterInstallParams_KnownID asserts that a known starter id
// resolves into ParsedFlags with Target=Project, Intent.Mode=Custom, the
// resolved harness ids in Intent.Custom, the requested focal agents, and an
// effective Starter carrying the aggregated MCP list (design D1, proposal
// "Pre-refactor" bullet).
func TestBuildStarterInstallParams_KnownID(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	projectRoot := t.TempDir()
	agents := []model.Agent{model.AgentClaude, model.AgentOpenCode}
	flags := headless.ParsedStarterAddFlags{
		StarterID:   "active-ia",
		ProjectPath: projectRoot,
		DryRun:      true,
		Yes:         true,
		Agents:      agents,
	}

	wantHarnesses, err := cat.ResolveStarter("active-ia")
	if err != nil {
		t.Fatalf("cat.ResolveStarter(active-ia) error = %v", err)
	}
	wantHarnessIDs := make([]string, 0, len(wantHarnesses))
	for _, h := range wantHarnesses {
		wantHarnessIDs = append(wantHarnessIDs, h.ID)
	}
	wantMCPs, err := cat.ResolveStarterMCPs("active-ia")
	if err != nil {
		t.Fatalf("cat.ResolveStarterMCPs(active-ia) error = %v", err)
	}

	params, err := headless.BuildStarterInstallParams(flags, cat, nil)
	if err != nil {
		t.Fatalf("BuildStarterInstallParams() error = %v", err)
	}

	if params.Target != model.Project {
		t.Errorf("Target = %v, want model.Project", params.Target)
	}
	if params.ProjectRoot != projectRoot {
		t.Errorf("ProjectRoot = %q, want %q", params.ProjectRoot, projectRoot)
	}
	if params.Intent.Mode != model.ModeCustom {
		t.Errorf("Intent.Mode = %v, want model.ModeCustom", params.Intent.Mode)
	}
	if len(params.Intent.Custom) != len(wantHarnessIDs) {
		t.Fatalf("Intent.Custom = %v, want %v", params.Intent.Custom, wantHarnessIDs)
	}
	for i, id := range wantHarnessIDs {
		if params.Intent.Custom[i] != id {
			t.Errorf("Intent.Custom[%d] = %q, want %q", i, params.Intent.Custom[i], id)
		}
	}
	if len(params.Intent.Agents) != len(agents) {
		t.Fatalf("Intent.Agents = %v, want %v", params.Intent.Agents, agents)
	}
	if params.Starter == nil {
		t.Fatal("Starter is nil, want non-nil effective starter")
	}
	if params.Starter.ID != "active-ia" {
		t.Errorf("Starter.ID = %q, want active-ia", params.Starter.ID)
	}
	if len(params.Starter.MCPs) != len(wantMCPs) {
		t.Fatalf("Starter.MCPs = %v, want %v", params.Starter.MCPs, wantMCPs)
	}
}

// TestBuildStarterInstallParams_UnknownID asserts that an unknown starter id
// returns an error identifying it, without producing a runnable ParsedFlags.
func TestBuildStarterInstallParams_UnknownID(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	flags := headless.ParsedStarterAddFlags{
		StarterID:   "does-not-exist",
		ProjectPath: t.TempDir(),
		Agents:      []model.Agent{model.AgentClaude},
	}

	_, err = headless.BuildStarterInstallParams(flags, cat, nil)
	if err == nil {
		t.Fatal("expected error for unknown starter id, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "does-not-exist") {
		t.Errorf("error = %q, want it to mention the unknown id", got)
	}
}
