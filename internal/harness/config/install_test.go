package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// claudeAdapter is a fake adapter for the Claude agent pointing to a temp file.
type claudeAdapter struct {
	path string
}

func (c claudeAdapter) Agent() model.Agent              { return model.AgentClaude }
func (c claudeAdapter) InstructionsPath(_ string) string { return c.path }
func (c claudeAdapter) VariantKey() string               { return "claude" }
func (c claudeAdapter) SettingsPath(_ string) string     { return "" }
func (c claudeAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}

// TestInstall_EndToEnd verifies the full flow: compose + inject + idempotency.
func TestInstall_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "CLAUDE.md")

	// Pre-existing file.
	if err := os.WriteFile(targetPath, []byte("# My Project Config\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := model.Harness{
		ID:      "sdd-orchestrator",
		Type:    model.HarnessConfig,
		Toggles: []string{"model-routing", "delegation"},
	}
	adapters := []config.AgentAdapter{claudeAdapter{path: targetPath}}

	// First install.
	result, err := config.Install(h, adapters, dir)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if result.AllAlready {
		t.Error("first install should NOT be AllAlready")
	}
	if len(result.Files) == 0 {
		t.Error("first install should list the written file")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	cs := string(content)

	// Section must be present and not duplicated.
	if !strings.Contains(cs, "<!-- jr-stack:sdd-orchestrator -->") {
		t.Error("installed file must contain sdd-orchestrator section")
	}
	count := strings.Count(cs, "<!-- jr-stack:sdd-orchestrator -->")
	if count != 1 {
		t.Errorf("section marker count = %d, want 1", count)
	}

	// Backup must exist.
	entries, err := os.ReadDir(filepath.Join(dir, ".jr-stack", "backups"))
	if err != nil {
		t.Fatalf("backup dir missing: %v", err)
	}
	if len(entries) == 0 {
		t.Error("backup should have been created on first install")
	}

	// Second install (idempotency).
	result2, err := config.Install(h, adapters, dir)
	if err != nil {
		t.Fatalf("second Install error: %v", err)
	}
	if !result2.AllAlready {
		t.Error("second install with same content should be AllAlready")
	}
}

// TestInstall_WrongType verifies that Install returns an error for non-config harnesses.
func TestInstall_WrongType(t *testing.T) {
	h := model.Harness{
		ID:   "some-skill",
		Type: model.HarnessSkill,
	}
	_, err := config.Install(h, nil, t.TempDir())
	if err == nil {
		t.Error("Install should error when harness type is not config")
	}
}

// TestInstall_SkipEmptyPath verifies adapters that return empty InstructionsPath
// are silently skipped (no error, no file written).
func TestInstall_SkipEmptyPath(t *testing.T) {
	dir := t.TempDir()
	h := model.Harness{
		ID:   "sdd-orchestrator",
		Type: model.HarnessConfig,
	}

	skipAdapter := fakeAdapter{agent: model.AgentCursor, path: "", variant: "cursor"}
	result, err := config.Install(h, []config.AgentAdapter{skipAdapter}, dir)
	if err != nil {
		t.Fatalf("Install with skip adapter error: %v", err)
	}
	if len(result.Files) != 0 {
		t.Error("skip adapter should produce no written files")
	}
}
