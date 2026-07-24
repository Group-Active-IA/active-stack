package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
)

// legacyClaudeMD reproduces the real-world bloated state reported by the user:
// an older installer layout left standalone persona / engram-protocol /
// strict-tdd-mode sections, plus a previous sdd-orchestrator block whose nested
// engram + tdd content now duplicates what the current installer inlines.
const legacyClaudeMD = "# My Project Config\n\n" +
	"<!-- active-stack:persona -->\n## Rules\nlegacy persona body\n<!-- /active-stack:persona -->\n\n" +
	"<!-- active-stack:engram-protocol -->\n## Engram Protocol\nlegacy engram body\n<!-- /active-stack:engram-protocol -->\n\n" +
	"<!-- active-stack:sdd-orchestrator -->\n# Old OPSX\n" +
	"<!-- active-stack:sdd-delegation -->\nold delegation\n<!-- /active-stack:sdd-delegation -->\n" +
	"<!-- active-stack:sdd-model-assignments -->\nold routing\n<!-- /active-stack:sdd-model-assignments -->\n" +
	"<!-- /active-stack:sdd-orchestrator -->\n\n" +
	"<!-- active-stack:strict-tdd-mode -->\nStrict TDD Mode: enabled\n<!-- /active-stack:strict-tdd-mode -->\n"

// TestInject_PurgesStaleSections verifies that a re-install cleans up every
// active-stack-marked section the current installer no longer owns BEFORE injecting,
// so legacy orphans (persona / engram-protocol / strict-tdd-mode) do not pile up.
func TestInject_PurgesStaleSections(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(target, []byte(legacyClaudeMD), 0o644); err != nil {
		t.Fatal(err)
	}

	composed := "# OPSX Orchestrator Instructions\n\nFresh orchestrator block.\n"
	snapshotDir := filepath.Join(dir, "backups")

	if _, err := config.Inject(target, composed, snapshotDir); err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)

	// Stale, non-owned sections must be gone entirely (open and close markers).
	staleIDs := []string{"persona", "engram-protocol", "strict-tdd-mode"}
	for _, id := range staleIDs {
		if strings.Contains(got, "<!-- active-stack:"+id+" -->") {
			t.Errorf("stale section %q opening marker should have been purged", id)
		}
		if strings.Contains(got, "<!-- /active-stack:"+id+" -->") {
			t.Errorf("stale section %q closing marker should have been purged", id)
		}
	}

	// The owned section must remain, exactly once, with the fresh content.
	if c := strings.Count(got, "<!-- active-stack:sdd-orchestrator -->"); c != 1 {
		t.Errorf("sdd-orchestrator marker count = %d, want 1", c)
	}
	if !strings.Contains(got, "Fresh orchestrator block.") {
		t.Error("fresh composed content should be present")
	}
	if strings.Contains(got, "# Old OPSX") {
		t.Error("old orchestrator body should have been replaced")
	}

	// User content outside any active-stack marker must be preserved.
	if !strings.Contains(got, "# My Project Config") {
		t.Error("user content outside markers must be preserved")
	}
}

// legacyJrStackClaudeMD reproduces a CLAUDE.md left behind by the project's
// previous name ("JR Stack", before the active-stack rebrand in 2f8336e),
// when markers were written as "<!-- jr-stack:ID -->" instead of today's
// "<!-- active-stack:ID -->". Mirrors legacyClaudeMD exactly, prefix aside.
const legacyJrStackClaudeMD = "# My Project Config\n\n" +
	"<!-- jr-stack:persona -->\n## Rules\nlegacy persona body\n<!-- /jr-stack:persona -->\n\n" +
	"<!-- jr-stack:engram-protocol -->\n## Engram Protocol\nlegacy engram body\n<!-- /jr-stack:engram-protocol -->\n\n" +
	"<!-- jr-stack:sdd-orchestrator -->\n# Old OPSX\n" +
	"<!-- jr-stack:sdd-delegation -->\nold delegation\n<!-- /jr-stack:sdd-delegation -->\n" +
	"<!-- jr-stack:sdd-model-assignments -->\nold routing\n<!-- /jr-stack:sdd-model-assignments -->\n" +
	"<!-- /jr-stack:sdd-orchestrator -->\n\n" +
	"<!-- jr-stack:strict-tdd-mode -->\nStrict TDD Mode: enabled\n<!-- /jr-stack:strict-tdd-mode -->\n"

// TestInject_MigratesLegacyJrStackInstallation is the regression guard for
// the reported bug: an install (or reinstall) run against a CLAUDE.md left
// by the old "JR Stack" installer name must purge every stale jr-stack:
// section and upgrade the owned sdd-orchestrator block to the current
// active-stack: prefix — not leave the old blocks orphaned forever.
func TestInject_MigratesLegacyJrStackInstallation(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(target, []byte(legacyJrStackClaudeMD), 0o644); err != nil {
		t.Fatal(err)
	}

	composed := "# OPSX Orchestrator Instructions\n\nFresh orchestrator block.\n"
	snapshotDir := filepath.Join(dir, "backups")

	if _, err := config.Inject(target, composed, snapshotDir); err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)

	// No jr-stack: marker of any kind should survive — stale sections
	// purged, and the owned section upgraded to the active-stack: prefix.
	if strings.Contains(got, "jr-stack:") {
		t.Errorf("no jr-stack: marker should survive an install, got:\n%s", got)
	}

	// The owned section must be present exactly once, under the current prefix.
	if c := strings.Count(got, "<!-- active-stack:sdd-orchestrator -->"); c != 1 {
		t.Errorf("sdd-orchestrator marker count = %d, want 1", c)
	}
	if !strings.Contains(got, "Fresh orchestrator block.") {
		t.Error("fresh composed content should be present")
	}
	if strings.Contains(got, "# Old OPSX") {
		t.Error("old orchestrator body should have been replaced")
	}

	// Stale legacy sections must be gone entirely, not just re-prefixed.
	for _, id := range []string{"persona", "engram-protocol", "strict-tdd-mode"} {
		if strings.Contains(got, id) {
			t.Errorf("stale legacy section %q should have been purged, not carried over", id)
		}
	}

	// User content outside any marker must be preserved.
	if !strings.Contains(got, "# My Project Config") {
		t.Error("user content outside markers must be preserved")
	}
}

// TestInject_PreservesOwnedNestedChildren guards against the purge being too
// aggressive: sdd-delegation / sdd-model-assignments are owned (nested children
// of sdd-orchestrator) and must NEVER be purged as standalone stale sections.
func TestInject_PreservesOwnedNestedChildren(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(target, []byte(legacyClaudeMD), 0o644); err != nil {
		t.Fatal(err)
	}

	// Compose a block that itself carries the nested owned children.
	composed := "# OPSX\n" +
		"<!-- active-stack:sdd-delegation -->\nnew delegation\n<!-- /active-stack:sdd-delegation -->\n" +
		"<!-- active-stack:sdd-model-assignments -->\nnew routing\n<!-- /active-stack:sdd-model-assignments -->\n"
	snapshotDir := filepath.Join(dir, "backups")

	if _, err := config.Inject(target, composed, snapshotDir); err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)

	for _, id := range []string{"sdd-delegation", "sdd-model-assignments"} {
		if !strings.Contains(got, "<!-- active-stack:"+id+" -->") {
			t.Errorf("owned nested child %q must be preserved", id)
		}
	}
}
