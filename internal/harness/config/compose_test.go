package config_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
)

// -update flag regenerates golden files.
var update = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(goldenPath(name))
	if err != nil {
		t.Fatalf("read golden %q: %v", name, err)
	}
	return string(data)
}

func writeGolden(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatalf("mkdir testdata: %v", err)
	}
	if err := os.WriteFile(goldenPath(name), []byte(content), 0o644); err != nil {
		t.Fatalf("write golden %q: %v", name, err)
	}
}

// TestCompose_AllToggles verifies that with all toggles active the composed
// output is deterministic and matches the golden file.
func TestCompose_AllToggles(t *testing.T) {
	toggles := []string{"delegation", "model-routing", "engram", "tdd", "governance"}
	got, err := config.Compose(toggles, "claude")
	if err != nil {
		t.Fatalf("Compose error: %v", err)
	}
	if got == "" {
		t.Fatal("Compose returned empty string")
	}

	name := "compose_all_toggles_claude"
	if *update {
		writeGolden(t, name, got)
		t.Logf("updated golden %s", name)
		return
	}
	want := loadGolden(t, name)
	if got != want {
		t.Errorf("Compose output does not match golden.\ngot len=%d want len=%d", len(got), len(want))
	}
}

// TestCompose_Deterministic verifies that toggle order does NOT change output.
func TestCompose_Deterministic(t *testing.T) {
	toggles1 := []string{"tdd", "governance", "engram", "delegation", "model-routing"}
	toggles2 := []string{"model-routing", "delegation", "tdd", "engram", "governance"}

	out1, err := config.Compose(toggles1, "claude")
	if err != nil {
		t.Fatalf("Compose(order1) error: %v", err)
	}
	out2, err := config.Compose(toggles2, "claude")
	if err != nil {
		t.Fatalf("Compose(order2) error: %v", err)
	}
	if out1 != out2 {
		t.Errorf("Compose is not deterministic: output differs by toggle order")
	}
}

// TestCompose_AdditiveToggleOff verifies that omitting an additive toggle
// (engram, tdd, governance) removes its fragment from the output.
func TestCompose_AdditiveToggleOff(t *testing.T) {
	// All toggles.
	withEngram, err := config.Compose([]string{"delegation", "model-routing", "engram"}, "claude")
	if err != nil {
		t.Fatalf("Compose with engram: %v", err)
	}
	// Without engram.
	withoutEngram, err := config.Compose([]string{"delegation", "model-routing"}, "claude")
	if err != nil {
		t.Fatalf("Compose without engram: %v", err)
	}

	if !strings.Contains(withEngram, "Engram Persistent Memory") {
		t.Error("output WITH engram toggle should contain Engram section")
	}
	if strings.Contains(withoutEngram, "Engram Persistent Memory") {
		t.Error("output WITHOUT engram toggle must NOT contain Engram section")
	}
}

// TestCompose_TDDToggleOff verifies that omitting tdd removes the TDD flag line.
func TestCompose_TDDToggleOff(t *testing.T) {
	withTDD, err := config.Compose([]string{"tdd"}, "claude")
	if err != nil {
		t.Fatalf("Compose with tdd: %v", err)
	}
	withoutTDD, err := config.Compose([]string{}, "claude")
	if err != nil {
		t.Fatalf("Compose without tdd: %v", err)
	}

	if !strings.Contains(withTDD, "Strict TDD Mode: enabled") {
		t.Error("output WITH tdd toggle should contain 'Strict TDD Mode: enabled'")
	}
	if strings.Contains(withoutTDD, "Strict TDD Mode: enabled") {
		t.Error("output WITHOUT tdd toggle must NOT contain 'Strict TDD Mode: enabled'")
	}
}

// TestCompose_ModelRoutingToggleOff verifies subtractive toggle removes the
// model-routing section from the output.
func TestCompose_ModelRoutingToggleOff(t *testing.T) {
	// With model-routing on, section should be present.
	with, err := config.Compose([]string{"model-routing"}, "claude")
	if err != nil {
		t.Fatalf("Compose with model-routing: %v", err)
	}
	// Without model-routing, section should be absent.
	without, err := config.Compose([]string{}, "claude")
	if err != nil {
		t.Fatalf("Compose without model-routing: %v", err)
	}

	if !strings.Contains(with, "Model Assignments") {
		t.Error("output WITH model-routing should contain 'Model Assignments'")
	}
	if strings.Contains(without, "Model Assignments") {
		t.Error("output WITHOUT model-routing must NOT contain 'Model Assignments'")
	}
}

// TestCompose_DelegationToggleOff verifies subtractive toggle removes the
// delegation section.
func TestCompose_DelegationToggleOff(t *testing.T) {
	with, err := config.Compose([]string{"delegation"}, "claude")
	if err != nil {
		t.Fatalf("Compose with delegation: %v", err)
	}
	without, err := config.Compose([]string{}, "claude")
	if err != nil {
		t.Fatalf("Compose without delegation: %v", err)
	}

	if !strings.Contains(with, "Delegation Rules") {
		t.Error("output WITH delegation should contain 'Delegation Rules'")
	}
	if strings.Contains(without, "Delegation Rules") {
		t.Error("output WITHOUT delegation must NOT contain 'Delegation Rules'")
	}
}

// TestCompose_UnknownToggle verifies that an unrecognized toggle returns an error.
func TestCompose_UnknownToggle(t *testing.T) {
	_, err := config.Compose([]string{"nonexistent-toggle"}, "claude")
	if err == nil {
		t.Error("Compose with unknown toggle should return an error")
	}
}

// TestCompose_GenericFallback verifies that an unknown variant falls back to generic.
func TestCompose_GenericFallback(t *testing.T) {
	// "unknown-agent" has no asset directory; must fall back to generic.
	out, err := config.Compose([]string{}, "unknown-agent")
	if err != nil {
		t.Fatalf("Compose with unknown variant should not error (falls back to generic): %v", err)
	}
	if out == "" {
		t.Error("Compose with generic fallback returned empty string")
	}
	// Should contain text from generic sdd-orchestrator.
	if !strings.Contains(out, "OPSX Orchestrator Instructions") {
		t.Error("generic fallback output should contain orchestrator heading")
	}
}

// TestCompose_VariantCodex verifies a non-claude variant loads its own base.
func TestCompose_VariantCodex(t *testing.T) {
	out, err := config.Compose([]string{}, "codex")
	if err != nil {
		t.Fatalf("Compose codex: %v", err)
	}
	if !strings.Contains(out, "Codex") {
		t.Error("codex variant output should contain 'Codex'")
	}
}

// TestCompose_BaseAlwaysPresent verifies base content is never stripped even
// when all toggles are off.
func TestCompose_BaseAlwaysPresent(t *testing.T) {
	out, err := config.Compose([]string{}, "claude")
	if err != nil {
		t.Fatalf("Compose empty toggles: %v", err)
	}
	if !strings.Contains(out, "OPSX Orchestrator Instructions") {
		t.Error("base is always present even with no toggles")
	}
}
