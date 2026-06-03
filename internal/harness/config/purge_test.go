package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
)

// legacyClaudeMD reproduces the real-world bloated state reported by the user:
// an older installer layout left standalone persona / engram-protocol /
// strict-tdd-mode sections, plus a previous sdd-orchestrator block whose nested
// engram + tdd content now duplicates what the current installer inlines.
const legacyClaudeMD = "# My Project Config\n\n" +
	"<!-- jr-stack:persona -->\n## Rules\nlegacy persona body\n<!-- /jr-stack:persona -->\n\n" +
	"<!-- jr-stack:engram-protocol -->\n## Engram Protocol\nlegacy engram body\n<!-- /jr-stack:engram-protocol -->\n\n" +
	"<!-- jr-stack:sdd-orchestrator -->\n# Old OPSX\n" +
	"<!-- jr-stack:sdd-delegation -->\nold delegation\n<!-- /jr-stack:sdd-delegation -->\n" +
	"<!-- jr-stack:sdd-model-assignments -->\nold routing\n<!-- /jr-stack:sdd-model-assignments -->\n" +
	"<!-- /jr-stack:sdd-orchestrator -->\n\n" +
	"<!-- jr-stack:strict-tdd-mode -->\nStrict TDD Mode: enabled\n<!-- /jr-stack:strict-tdd-mode -->\n"

// TestInject_PurgesStaleSections verifies that a re-install cleans up every
// jr-stack-marked section the current installer no longer owns BEFORE injecting,
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
		if strings.Contains(got, "<!-- jr-stack:"+id+" -->") {
			t.Errorf("stale section %q opening marker should have been purged", id)
		}
		if strings.Contains(got, "<!-- /jr-stack:"+id+" -->") {
			t.Errorf("stale section %q closing marker should have been purged", id)
		}
	}

	// The owned section must remain, exactly once, with the fresh content.
	if c := strings.Count(got, "<!-- jr-stack:sdd-orchestrator -->"); c != 1 {
		t.Errorf("sdd-orchestrator marker count = %d, want 1", c)
	}
	if !strings.Contains(got, "Fresh orchestrator block.") {
		t.Error("fresh composed content should be present")
	}
	if strings.Contains(got, "# Old OPSX") {
		t.Error("old orchestrator body should have been replaced")
	}

	// User content outside any jr-stack marker must be preserved.
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
		"<!-- jr-stack:sdd-delegation -->\nnew delegation\n<!-- /jr-stack:sdd-delegation -->\n" +
		"<!-- jr-stack:sdd-model-assignments -->\nnew routing\n<!-- /jr-stack:sdd-model-assignments -->\n"
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
		if !strings.Contains(got, "<!-- jr-stack:"+id+" -->") {
			t.Errorf("owned nested child %q must be preserved", id)
		}
	}
}
