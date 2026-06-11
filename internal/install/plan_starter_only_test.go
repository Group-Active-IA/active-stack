package install_test

// Regression test for the planner-resolves-starter-only fix (C-30 / planner-resolves-starter-only).
//
// Root cause: plan.go built the planner's dependency-resolution index from
// cat.ForMode(model.ModeCustom), which excludes starter-only harnesses since
// C-32. Starter-only harnesses resolved upstream by SelectHarnesses (via ByID)
// were therefore unknown to the planner → "unknown harness" error.
//
// Fix: build the index from cat.AllHarnesses() so the resolution universe is
// the COMPLETE catalog. SelectHarnesses remains the single selection gate.

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestBuildPlan_StarterOnlyHarnessResolvesViaFullCatalog asserts that BuildPlan
// does NOT fail with "unknown harness" when the selected set (built via ModeCustom
// from a starter that includes starter-only harnesses) is handed to the planner.
//
// This is the install-layer regression test for the bug fixed by AllHarnesses().
// It must FAIL on the old tree (planner index = ForMode(ModeCustom), excludes
// starter-only) and PASS after the fix (planner index = AllHarnesses()).
func TestBuildPlan_StarterOnlyHarnessResolvesViaFullCatalog(t *testing.T) {
	// Inject no-op snapshot/restore so BuildPlan does not need the filesystem.
	restoreSnap := install.SetSnapshotCreateWithHints(func(dir string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	// Load the real embedded catalog — same as production.
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	// The "base" starter bundles test-driven-development (a starter-only harness).
	// Mirroring starter_add.go: resolve → collect IDs → ModeCustom intent.
	harnesses, err := cat.ResolveStarter("base")
	if err != nil {
		t.Fatalf("ResolveStarter(\"base\") error = %v", err)
	}
	harnessIDs := make([]string, 0, len(harnesses))
	for _, h := range harnesses {
		harnessIDs = append(harnessIDs, h.ID)
	}
	if len(harnessIDs) == 0 {
		t.Fatal("starter \"base\" resolved to 0 harnesses — catalog issue")
	}

	// Confirm test-driven-development is in the resolved set.
	found := false
	for _, id := range harnessIDs {
		if id == "test-driven-development" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("starter \"base\" did not resolve test-driven-development; ids = %v", harnessIDs)
	}

	intent := install.Intent{
		Mode:   model.ModeCustom,
		Custom: harnessIDs,
		Agents: []model.Agent{model.AgentClaude},
	}

	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	// (a) BuildPlan must NOT return an "unknown harness" error.
	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		if strings.Contains(err.Error(), "unknown harness") {
			t.Fatalf("regression: BuildPlan returned \"unknown harness\" error: %v\n"+
				"Root cause: planner index built from ForMode(ModeCustom) excludes starter-only harnesses.\n"+
				"Fix: use cat.AllHarnesses() to build the planner's resolution index.", err)
		}
		t.Fatalf("BuildPlan() unexpected error = %v", err)
	}

	// (b) test-driven-development must appear in the plan's Apply steps.
	ids := applyStepIDs(plan.StagePlan)
	if !containsID(ids, "test-driven-development") {
		t.Errorf("test-driven-development should be in the plan Apply steps, got %v", ids)
	}
}

// TestBuildPlan_FullMode_NoStarterOnlyHarness asserts that a install --mode full
// plan contains no harness with Scope == ScopeStarterOnly, even after the planner
// index is widened to AllHarnesses(). This proves scope enforcement stays at the
// selection layer (SelectHarnesses/ForMode), not at the resolution layer.
func TestBuildPlan_FullMode_NoStarterOnlyHarness(t *testing.T) {
	// Inject no-op snapshot/restore.
	restoreSnap := install.SetSnapshotCreateWithHints(func(dir string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap", RootDir: dir}, nil
	})
	defer restoreSnap()

	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	intent := install.Intent{
		Mode:   model.ModeFull,
		Agents: []model.Agent{model.AgentClaude},
	}

	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	plan, err := install.BuildPlan(cat, intent, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// No starter-only harness should appear in the plan.
	ids := applyStepIDs(plan.StagePlan)
	for _, id := range ids {
		h, ok := cat.ByID(id)
		if !ok {
			continue // IDs like "verify-hook" may not be in catalog
		}
		if h.IsStarterOnly() {
			t.Errorf("starter-only harness %q leaked into --mode full plan (scope filter broken)", id)
		}
	}
}
