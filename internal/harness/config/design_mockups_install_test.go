package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

const (
	dmOpen  = "<!-- active-stack:design-mockups -->"
	dmClose = "<!-- /active-stack:design-mockups -->"
	dmHead  = "## Design Stage — UI Mockups"
)

func designMockupsHarness() model.Harness {
	return model.Harness{
		ID:        "design-mockups",
		Type:      model.HarnessConfig,
		Agents:    []model.Agent{model.AgentClaude},
		DependsOn: []string{"sdd-orchestrator"},
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// TestInstall_DesignMockups_ClaudeOnly verifies the section is injected only
// for claude even when opencode (primary-agent) and codex are in the same call.
func TestInstall_DesignMockups_ClaudeOnly(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	ocInstr := filepath.Join(home, "oc", "AGENTS.md")
	ocSettings := filepath.Join(home, "oc", "opencode.json")
	codexMD := filepath.Join(home, "codex", "AGENTS.md")

	ocBefore := `{"agent":{"keep":{"mode":"primary"}}}`
	if err := os.MkdirAll(filepath.Dir(ocSettings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ocSettings, []byte(ocBefore), 0o644); err != nil {
		t.Fatal(err)
	}

	adapters := []config.AgentAdapter{
		claudeAdapter{path: claudeMD},
		opencodeAdapter{instr: ocInstr, settings: ocSettings},
		fakeAdapter{agent: model.AgentCodex, path: codexMD, variant: "codex"},
	}

	res, err := config.Install(designMockupsHarness(), adapters, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0] != claudeMD {
		t.Errorf("Files = %v, want only %s", res.Files, claudeMD)
	}

	got := readFile(t, claudeMD)
	if c := strings.Count(got, dmOpen); c != 1 {
		t.Errorf("design-mockups open marker count = %d, want 1", c)
	}
	if !strings.Contains(got, dmHead) {
		t.Error("claude file must contain the design stage heading")
	}
	if strings.Contains(got, "active-stack:sdd-orchestrator") {
		t.Error("design-mockups install must not compose the orchestrator block")
	}

	if after := readFile(t, ocSettings); after != ocBefore {
		t.Errorf("opencode settings changed:\n%s", after)
	}
	if _, err := os.Stat(ocInstr); err == nil {
		t.Error("opencode instructions file must not be created")
	}
	if _, err := os.Stat(codexMD); err == nil {
		t.Error("codex instructions file must not be created")
	}
}

// TestInstall_DesignMockups_Idempotent verifies re-installing neither
// duplicates the section nor rewrites the file.
func TestInstall_DesignMockups_Idempotent(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	adapters := []config.AgentAdapter{claudeAdapter{path: claudeMD}}

	if _, err := config.Install(designMockupsHarness(), adapters, home); err != nil {
		t.Fatalf("first Install: %v", err)
	}
	second, err := config.Install(designMockupsHarness(), adapters, home)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !second.AllAlready {
		t.Error("second install should report AllAlready")
	}
	if c := strings.Count(readFile(t, claudeMD), dmOpen); c != 1 {
		t.Errorf("section count after reinstall = %d, want 1", c)
	}
}

// TestInstall_DesignMockups_PreservesOrchestratorAndUserContent verifies the
// install only adds its own section.
func TestInstall_DesignMockups_PreservesOrchestratorAndUserContent(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	orch := "<!-- active-stack:sdd-orchestrator -->\norchestrator body\n<!-- /active-stack:sdd-orchestrator -->"
	initial := "# My notes\n\nmine before\n\n" + orch + "\n\nmine after\n"
	if err := os.WriteFile(claudeMD, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Install(designMockupsHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err != nil {
		t.Fatalf("Install: %v", err)
	}
	got := readFile(t, claudeMD)
	for _, keep := range []string{"# My notes", "mine before", "mine after", orch} {
		if !strings.Contains(got, keep) {
			t.Errorf("content %q lost after install", keep)
		}
	}
	if !strings.Contains(got, dmOpen) || !strings.Contains(got, dmClose) {
		t.Error("design-mockups section missing")
	}
}

// TestInstall_DesignMockups_BacksUpFirst verifies a snapshot is taken before
// the file is written, and that a failed snapshot leaves the file untouched.
func TestInstall_DesignMockups_BacksUpFirst(t *testing.T) {
	t.Run("snapshot runs before write", func(t *testing.T) {
		home := t.TempDir()
		claudeMD := filepath.Join(home, "CLAUDE.md")
		original := "# original\n"
		if err := os.WriteFile(claudeMD, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
		orig := config.SnapshotterCreate
		defer func() { config.SnapshotterCreate = orig }()
		var snapped []string
		var contentAtSnapshot string
		config.SnapshotterCreate = func(_ string, paths []string) error {
			snapped = append(snapped, paths...)
			contentAtSnapshot = readFile(t, claudeMD)
			return nil
		}

		if _, err := config.Install(designMockupsHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err != nil {
			t.Fatalf("Install: %v", err)
		}
		if len(snapped) != 1 || snapped[0] != claudeMD {
			t.Errorf("snapshot paths = %v, want [%s]", snapped, claudeMD)
		}
		if contentAtSnapshot != original {
			t.Error("snapshot must run before the file is modified")
		}
	})

	t.Run("failed snapshot aborts without touching the file", func(t *testing.T) {
		home := t.TempDir()
		claudeMD := filepath.Join(home, "CLAUDE.md")
		original := "# original\n"
		if err := os.WriteFile(claudeMD, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
		orig := config.SnapshotterCreate
		defer func() { config.SnapshotterCreate = orig }()
		config.SnapshotterCreate = func(string, []string) error { return errors.New("disk full") }

		if _, err := config.Install(designMockupsHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err == nil {
			t.Fatal("Install must fail when the backup fails")
		}
		if got := readFile(t, claudeMD); got != original {
			t.Errorf("file modified despite backup failure:\n%s", got)
		}
	})
}

// TestInstall_Orchestrator_KeepsDesignMockupsSection verifies that installing
// the orchestrator over a file that already has the design section keeps it.
func TestInstall_Orchestrator_KeepsDesignMockupsSection(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	adapters := []config.AgentAdapter{claudeAdapter{path: claudeMD}}

	if _, err := config.Install(designMockupsHarness(), adapters, home); err != nil {
		t.Fatalf("design install: %v", err)
	}
	orch := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation"}}
	if _, err := config.Install(orch, adapters, home); err != nil {
		t.Fatalf("orchestrator install: %v", err)
	}
	got := readFile(t, claudeMD)
	if c := strings.Count(got, dmOpen); c != 1 {
		t.Errorf("design-mockups section count = %d, want 1 after orchestrator install", c)
	}
	if !strings.Contains(got, "<!-- active-stack:sdd-orchestrator -->") {
		t.Error("orchestrator section must be present")
	}
}

// TestDesignMockups_UpgradeReplacesOldSectionInPlace verifies that installing
// over a file holding the previous section content replaces it in place: one
// design-mockups section with the new visual verify step, and the orchestrator
// section plus user content untouched byte for byte (visual-verify).
func TestDesignMockups_UpgradeReplacesOldSectionInPlace(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	orch := "<!-- active-stack:sdd-orchestrator -->\norchestrator body\n<!-- /active-stack:sdd-orchestrator -->"
	oldSection := dmOpen + "\n## Design Stage — UI Mockups\n\nold content without the visual step\n" + dmClose
	before := "# My notes\n\nmine before\n\n" + orch + "\n\n" + oldSection + "\n\nmine after\n"
	if err := os.WriteFile(claudeMD, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := config.Install(designMockupsHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.AllAlready {
		t.Error("upgrade must report a change, not AllAlready")
	}

	got := readFile(t, claudeMD)
	if c := strings.Count(got, dmOpen); c != 1 {
		t.Errorf("design-mockups open marker count = %d, want 1", c)
	}
	if c := strings.Count(got, dmClose); c != 1 {
		t.Errorf("design-mockups close marker count = %d, want 1", c)
	}
	if strings.Contains(got, "old content without the visual step") {
		t.Error("old section content must be replaced")
	}
	if !strings.Contains(got, "Visual verify") {
		t.Error("new section must contain the visual verify step")
	}
	for _, keep := range []string{"# My notes\n\nmine before\n\n" + orch + "\n\n", "\n\nmine after\n"} {
		if !strings.Contains(got, keep) {
			t.Errorf("surrounding content %q not preserved byte for byte", keep)
		}
	}
	// No marker other than the two known sections was introduced.
	if c := strings.Count(got, "<!-- active-stack:"); c != 2 {
		t.Errorf("open marker count = %d, want 2 (orchestrator + design-mockups)", c)
	}
	// Re-running on the upgraded file is a no-op.
	again, err := config.Install(designMockupsHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !again.AllAlready {
		t.Error("second install after upgrade should report AllAlready")
	}
}
