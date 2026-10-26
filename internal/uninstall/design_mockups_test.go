package uninstall_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

const (
	dmOrchSection = "<!-- active-stack:sdd-orchestrator -->\norchestrator block\n<!-- /active-stack:sdd-orchestrator -->\n"
	dmSection     = "<!-- active-stack:design-mockups -->\ndesign block\n<!-- /active-stack:design-mockups -->\n"
)

// uninstallConfigHarness writes content to the claude instructions file, runs
// the targeted uninstall of the config harness id and returns the file content.
func uninstallConfigHarness(t *testing.T, id, content string) string {
	t.Helper()
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}
	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instrPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	defer uninstall.SetSnapshotCreate(func(string, []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})()

	h := model.Harness{ID: id, Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeFull}}
	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{h}},
		uninstall.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeFull, Strategy: uninstall.StrategyTargeted},
		uninstall.Options{
			HomeDir:  homeDir,
			Registry: &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: adapter}},
		},
	)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatalf("step.Run: %v", err)
		}
	}
	raw, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestMarkerRemoval_DesignMockups_Clean: uninstalling design-mockups removes
// only its own section (characterization: the removal logic is generic by ID).
func TestMarkerRemoval_DesignMockups_Clean(t *testing.T) {
	got := uninstallConfigHarness(t, "design-mockups", "# Mine\n\n"+dmOrchSection+"\n"+dmSection+"\nfooter\n")
	if strings.Contains(got, "design-mockups") || strings.Contains(got, "design block") {
		t.Errorf("design-mockups section not removed:\n%s", got)
	}
	for _, keep := range []string{"# Mine", "footer", "<!-- active-stack:sdd-orchestrator -->", "orchestrator block"} {
		if !strings.Contains(got, keep) {
			t.Errorf("content %q lost:\n%s", keep, got)
		}
	}
}

// TestMarkerRemoval_DesignMockups_AbsentIsNoop: nothing to remove, file intact.
func TestMarkerRemoval_DesignMockups_AbsentIsNoop(t *testing.T) {
	original := "# Mine\n\n" + dmOrchSection + "\nfooter\n"
	if got := uninstallConfigHarness(t, "design-mockups", original); got != original {
		t.Errorf("file modified:\n%s", got)
	}
}

// TestMarkerRemoval_Orchestrator_KeepsDesignMockups: uninstalling only the
// orchestrator must not purge the design-mockups section (owned section).
func TestMarkerRemoval_Orchestrator_KeepsDesignMockups(t *testing.T) {
	got := uninstallConfigHarness(t, "sdd-orchestrator", "# Mine\n\n"+dmOrchSection+"\n"+dmSection)
	if strings.Contains(got, "orchestrator block") {
		t.Errorf("orchestrator section should be removed:\n%s", got)
	}
	if !strings.Contains(got, "<!-- active-stack:design-mockups -->") || !strings.Contains(got, "design block") {
		t.Errorf("design-mockups section must survive:\n%s", got)
	}
}
