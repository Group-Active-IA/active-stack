package uninstall_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// legacyClaudeMD mirrors a real bloated install: the current owned section
// (sdd-orchestrator, with its nested owned children) plus standalone legacy
// sections from older layouts (persona, engram-protocol, strict-tdd-mode).
const legacyClaudeMD = "# My Config\n\nUser-written intro.\n\n" +
	"<!-- jr-stack:persona -->\nlegacy persona\n<!-- /jr-stack:persona -->\n\n" +
	"<!-- jr-stack:engram-protocol -->\nlegacy engram\n<!-- /jr-stack:engram-protocol -->\n\n" +
	"<!-- jr-stack:sdd-orchestrator -->\norchestrator\n" +
	"<!-- jr-stack:sdd-delegation -->\ndeleg\n<!-- /jr-stack:sdd-delegation -->\n" +
	"<!-- /jr-stack:sdd-orchestrator -->\n\n" +
	"<!-- jr-stack:strict-tdd-mode -->\nStrict TDD Mode: enabled\n<!-- /jr-stack:strict-tdd-mode -->\n\n" +
	"# User footer\n"

// TestMarkerRemovalStepPurgesStaleSections verifies that uninstalling the config
// harness removes its OWN section AND purges every legacy jr-stack section left
// by older layouts, leaving only the user's own content behind.
func TestMarkerRemovalStepPurgesStaleSections(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	instrPath := adapter.InstructionsPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(instrPath), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(instrPath, []byte(legacyClaudeMD), 0o644); err != nil {
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
			t.Fatalf("step.Run() error = %v", err)
		}
	}

	raw, err := os.ReadFile(instrPath)
	if err != nil {
		t.Fatalf("read instructions: %v", err)
	}
	got := string(raw)

	// EVERY jr-stack section (owned + legacy + nested) must be gone.
	for _, id := range []string{
		"persona", "engram-protocol", "strict-tdd-mode",
		"sdd-orchestrator", "sdd-delegation",
	} {
		if contains(got, "<!-- jr-stack:"+id+" -->") {
			t.Errorf("section %q opening marker still present after uninstall:\n%s", id, got)
		}
		if contains(got, "<!-- /jr-stack:"+id+" -->") {
			t.Errorf("section %q closing marker still present after uninstall:\n%s", id, got)
		}
	}

	// User content outside jr-stack markers must survive.
	if !contains(got, "# My Config") || !contains(got, "User-written intro.") || !contains(got, "# User footer") {
		t.Errorf("user content was lost during uninstall:\n%s", got)
	}
}
