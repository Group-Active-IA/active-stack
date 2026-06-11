package install_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestBuildPlanMachineTarget_SnapshotDirUnderHomeDir is the explicit
// zero-regression guard for C-27. It asserts that when Target is unspecified
// (zero-value = Machine), the snapshot dir is under homeDir, exactly as before.
func TestBuildPlanMachineTarget_SnapshotDirUnderHomeDir(t *testing.T) {
	homeDir := t.TempDir()
	capturedSnapDir := ""

	restore := install.SetSnapshotCreate(func(dir string, _ []string) (backup.Manifest, error) {
		capturedSnapDir = dir
		return backup.Manifest{}, nil
	})
	defer restore()

	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return nil
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	// No Target set → zero-value = Machine. No ProjectRoot set.
	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Prepare {
		_ = step.Run()
	}

	wantSnapDir := filepath.Join(homeDir, ".active-stack", "backups", "install")
	if capturedSnapDir != wantSnapDir {
		t.Errorf("machine snapshot dir = %q, want %q", capturedSnapDir, wantSnapDir)
	}
}

// TestBuildPlanMachineTarget_PathsUnderHomeDir verifies that without a Target,
// all write paths resolve under homeDir (zero-regression from C-27).
func TestBuildPlanMachineTarget_PathsUnderHomeDir(t *testing.T) {
	homeDir := filepath.FromSlash("/home/machineuser")
	var capturedPaths []string

	restore := install.SetSnapshotCreate(func(_ string, paths []string) (backup.Manifest, error) {
		capturedPaths = paths
		return backup.Manifest{}, nil
	})
	defer restore()

	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return nil
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}

	// No Target → Machine (zero-value). Paths must be under homeDir.
	// NoSelfInstall=true: this test is not about self-install, it checks
	// harness paths only — exclude the user-system bin dir from assertions.
	opts := buildOptions(homeDir, reg, nil)
	opts.NoSelfInstall = true
	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Prepare {
		_ = step.Run()
	}

	homeDirSlash := filepath.ToSlash(homeDir)
	for _, p := range capturedPaths {
		if !strings.HasPrefix(filepath.ToSlash(p), homeDirSlash) {
			t.Errorf("machine write path %q does not resolve under homeDir %q", p, homeDir)
		}
	}
}
