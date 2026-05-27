package uninstall_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ─────────────────────────────────────────────────────────────────
// markerRemovalStep tests
// ─────────────────────────────────────────────────────────────────

// TestMarkerRemovalStepRemovesSection verifies that the step removes the
// harness marker section from the agent instructions file, leaving surrounding
// content untouched.
func TestMarkerRemovalStepRemovesSection(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	const sectionID = "sdd-orchestrator"
	original := "# My Config\n\nSome content.\n\n<!-- jr-stack:sdd-orchestrator -->\norchestrator block\n<!-- /jr-stack:sdd-orchestrator -->\n\nMore content.\n"
	if err := os.WriteFile(instrPath, []byte(original), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           sectionID,
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Execute via the plan directly.
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	got, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatalf("read instructions: %v", err)
	}

	if contains(string(got), "<!-- jr-stack:sdd-orchestrator -->") {
		t.Errorf("marker still present after removal:\n%s", got)
	}
	if contains(string(got), "orchestrator block") {
		t.Errorf("section content still present after removal:\n%s", got)
	}
	if !contains(string(got), "# My Config") {
		t.Errorf("surrounding content was lost:\n%s", got)
	}
	if !contains(string(got), "More content.") {
		t.Errorf("trailing content was lost:\n%s", got)
	}
}

// TestMarkerRemovalStepAbsentSectionIsNoop verifies that when the section is
// not present the step completes without error and leaves the file unchanged.
func TestMarkerRemovalStepAbsentSectionIsNoop(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	const original = "# Config\n\nNo marker here.\n"
	if err := os.WriteFile(instrPath, []byte(original), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() on absent section returned error = %v", err)
		}
	}

	got, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatalf("read instructions: %v", err)
	}
	if string(got) != original {
		t.Errorf("file was modified despite absent section.\ngot:\n%s\nwant:\n%s", got, original)
	}
}

// TestMarkerRemovalStepRepeatedRunIsNoop verifies idempotency: running the step
// a second time on a file that already has the section removed does not error.
func TestMarkerRemovalStepRepeatedRunIsNoop(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	const sectionID = "sdd-orchestrator"
	original := "# Config\n\n<!-- jr-stack:sdd-orchestrator -->\nblock\n<!-- /jr-stack:sdd-orchestrator -->\n"
	if err := os.WriteFile(instrPath, []byte(original), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           sectionID,
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}}

	buildAndRunPlan := func() error {
		plan, err := uninstall.BuildPlan(
			cat,
			uninstall.Intent{
				Agents:   []model.Agent{model.AgentClaude},
				Mode:     model.ModeLite,
				Strategy: uninstall.StrategyTargeted,
			},
			uninstall.Options{HomeDir: homeDir, Registry: reg},
		)
		if err != nil {
			return err
		}
		for _, step := range plan.Apply {
			if err := step.Run(); err != nil {
				return err
			}
		}
		return nil
	}

	if err := buildAndRunPlan(); err != nil {
		t.Fatalf("first run error = %v", err)
	}
	if err := buildAndRunPlan(); err != nil {
		t.Errorf("second run (repeated uninstall) error = %v", err)
	}
}

// TestMarkerRemovalStepPathResolvedViaAdapter verifies that the instructions
// path comes from the adapter, not from a hardcoded literal.
func TestMarkerRemovalStepPathResolvedViaAdapter(t *testing.T) {
	homeDir := t.TempDir()
	// Use a custom path returned by the adapter to detect hardcoding.
	customPath := filepath.Join(homeDir, "custom-agent", "CLAUDE.md")
	adapter := fakeAdapterCustomPath{
		agent:            model.AgentClaude,
		instructionsPath: customPath,
	}

	if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := "# Custom\n\n<!-- jr-stack:test-h -->\ncontent\n<!-- /jr-stack:test-h -->\n"
	if err := os.WriteFile(customPath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	h := model.Harness{
		ID:           "test-h",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistryCustom{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	got, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("read custom path: %v", err)
	}
	if contains(string(got), "<!-- jr-stack:test-h -->") {
		t.Errorf("marker still present; adapter path was not used:\n%s", got)
	}
}

// ─────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
