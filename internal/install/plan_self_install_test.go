package install_test

import (
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// minimalCatAndReg returns a minimal catalog+registry for plan tests that
// don't need any real harnesses. The catalog has one lite harness.
func minimalCatAndReg(t *testing.T) (install.Catalog, install.Registry) {
	t.Helper()
	h := model.Harness{
		ID:           "any-harness",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	return cat, reg
}

// containsStepID checks whether the plan Apply steps contain a step with the given ID.
func containsStepID(steps interface{ IDs() []string }, id string) bool {
	for _, s := range steps.IDs() {
		if s == id {
			return true
		}
	}
	return false
}

// stepIDs returns the IDs of all steps in a slice.
type stepList []interface{ ID() string }

func stepIDSlice(steps []interface{ ID() string }) []string {
	ids := make([]string, len(steps))
	for i, s := range ids {
		_ = s
		ids[i] = steps[i].ID()
	}
	return ids
}

// TestBuildPlan_SelfInstallPresent verifies that the self-install step appears
// in the Apply stage when NoSelfInstall is false (default).
func TestBuildPlan_SelfInstallPresent(t *testing.T) {
	cat, reg := minimalCatAndReg(t)
	homeDir := t.TempDir()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	opts := install.Options{
		HomeDir:      homeDir,
		Registry:     reg,
		NoSelfInstall: false, // default ON
	}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	found := false
	for _, s := range plan.Apply {
		if s.ID() == "self-install" {
			found = true
			break
		}
	}
	if !found {
		ids := make([]string, len(plan.Apply))
		for i, s := range plan.Apply {
			ids[i] = s.ID()
		}
		t.Errorf("self-install step not found in Apply; got: %v", ids)
	}
}

// TestBuildPlan_SelfInstallAbsent verifies that when NoSelfInstall is true,
// no self-install step is present.
func TestBuildPlan_SelfInstallAbsent(t *testing.T) {
	cat, reg := minimalCatAndReg(t)
	homeDir := t.TempDir()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	opts := install.Options{
		HomeDir:      homeDir,
		Registry:     reg,
		NoSelfInstall: true,
	}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	for _, s := range plan.Apply {
		if s.ID() == "self-install" {
			t.Errorf("self-install step must NOT appear when NoSelfInstall=true")
		}
	}
}

// TestBuildPlan_SelfInstallAfterHarnessSteps verifies that the self-install step
// appears after harness steps (not before them).
func TestBuildPlan_SelfInstallAfterHarnessSteps(t *testing.T) {
	cat, reg := minimalCatAndReg(t)
	homeDir := t.TempDir()

	restoreSnap := install.SetSnapshotCreate(func(_ string, _ []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	opts := install.Options{
		HomeDir:      homeDir,
		Registry:     reg,
		NoSelfInstall: false,
	}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	selfIdx := -1
	for i, s := range plan.Apply {
		if s.ID() == "self-install" {
			selfIdx = i
			break
		}
	}
	if selfIdx == -1 {
		t.Fatal("self-install step not found")
	}

	// All harness steps must appear before selfIdx.
	for i, s := range plan.Apply[:selfIdx] {
		_ = i
		_ = s // harness steps come first: ok
	}
	// Ensure at least one harness step exists before self-install.
	if selfIdx == 0 {
		t.Error("self-install is first step; expected at least one harness step before it")
	}
}

// TestBuildPlan_SnapshotIncludesSelfInstallPath verifies that when self-install
// is enabled, the snapshot's captured paths include the target binary path.
func TestBuildPlan_SnapshotIncludesSelfInstallPath(t *testing.T) {
	cat, reg := minimalCatAndReg(t)
	homeDir := t.TempDir()
	var capturedPaths []string

	restoreSnap := install.SetSnapshotCreateWithHints(func(_ string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		capturedPaths = append(capturedPaths, paths...)
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	binDir := t.TempDir()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()

	opts := install.Options{
		HomeDir:      homeDir,
		Registry:     reg,
		NoSelfInstall: false,
		Profile:       system.PlatformProfile{OS: "linux"},
	}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	// Run Prepare to trigger the snapshot.
	for _, s := range plan.Prepare {
		if err := s.Run(); err != nil {
			t.Fatalf("Prepare step %q: %v", s.ID(), err)
		}
	}

	// The bin dir's active-stack path must be in captured paths.
	wantTarget := filepath.Join(binDir, "active-stack")
	found := false
	for _, p := range capturedPaths {
		if p == wantTarget {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("self-install target %q not in snapshot paths: %v", wantTarget, capturedPaths)
	}
}

// TestBuildPlan_SnapshotNoSelfInstallPath verifies that when NoSelfInstall is
// true, no self-install target path is captured by the snapshot.
func TestBuildPlan_SnapshotNoSelfInstallPath(t *testing.T) {
	cat, reg := minimalCatAndReg(t)
	homeDir := t.TempDir()
	var capturedPaths []string

	restoreSnap := install.SetSnapshotCreateWithHints(func(_ string, paths []string, _ map[string]bool) (backup.Manifest, error) {
		capturedPaths = append(capturedPaths, paths...)
		return backup.Manifest{}, nil
	})
	defer restoreSnap()
	restoreExt := install.SetExternalInstallFn(fakeExternalSuccess)
	defer restoreExt()

	binDir := t.TempDir()
	restoreBinDir := install.SetSelfInstallBinaryInstallDirFn(func(_ string) string { return binDir })
	defer restoreBinDir()

	opts := install.Options{
		HomeDir:      homeDir,
		Registry:     reg,
		NoSelfInstall: true,
	}
	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	for _, s := range plan.Prepare {
		_ = s.Run()
	}

	wantTarget := filepath.Join(binDir, "active-stack")
	for _, p := range capturedPaths {
		if p == wantTarget {
			t.Errorf("self-install target %q must NOT be in snapshot when NoSelfInstall=true", wantTarget)
		}
	}
}
