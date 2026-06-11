package filemerge

import (
	"strings"
	"testing"
)

// Adversarial edge cases for InjectMarkdownSection, MergeJSONObjects, WriteFileAtomic,
// UpsertCodexEngramBlock, and UpsertTopLevelTOMLString.
// All tests for the removed legacy strip functions have been dropped (dead code in new repo).

func TestAdversarial_InjectMarkdownSection_EmptyExistingAndContent(t *testing.T) {
	result := InjectMarkdownSection("", "sdd", "")
	if result != "" {
		t.Fatalf("empty existing + empty content: expected empty, got %q", result)
	}
}

func TestAdversarial_InjectMarkdownSection_NReinjectionIdempotent(t *testing.T) {
	// Calling N times with same args must produce the same result as calling once.
	content := "## Orchestrator\nDelegate everything.\n"
	base := InjectMarkdownSection("", "sdd-orchestrator", content)
	for i := 0; i < 5; i++ {
		next := InjectMarkdownSection(base, "sdd-orchestrator", content)
		if next != base {
			t.Fatalf("injection not idempotent at iteration %d:\nbefore: %q\nafter:  %q", i+1, base, next)
		}
		base = next
	}
	// Exactly one open marker must be present.
	if count := strings.Count(base, "<!-- active-stack:sdd-orchestrator -->"); count != 1 {
		t.Fatalf("expected 1 open marker, got %d", count)
	}
}

func TestAdversarial_InjectMarkdownSection_MarkerWithSpecialCharsInContent(t *testing.T) {
	// Content that contains marker-like strings should not confuse the parser.
	content := "Use <!-- active-stack:other --> as example.\n"
	result := InjectMarkdownSection("", "myid", content)

	if !strings.Contains(result, "<!-- active-stack:myid -->") {
		t.Fatalf("open marker missing; got: %q", result)
	}
	if !strings.Contains(result, "<!-- /active-stack:myid -->") {
		t.Fatalf("close marker missing; got: %q", result)
	}
}

func TestAdversarial_MergeJSONObjects_EmptyBaseAndOverlay(t *testing.T) {
	merged, err := MergeJSONObjects([]byte(`{}`), []byte(`{}`))
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}
	if strings.TrimSpace(string(merged)) != "{}" {
		t.Fatalf("expected {}, got %q", merged)
	}
}

func TestAdversarial_MergeJSONObjects_EmptyRawBytes(t *testing.T) {
	// Empty raw bytes (not even {}) — base should be treated as {}.
	merged, err := MergeJSONObjects([]byte{}, []byte(`{"key":"val"}`))
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}
	if !strings.Contains(string(merged), `"key"`) {
		t.Fatalf("overlay key missing; got: %s", merged)
	}
}

func TestAdversarial_MergeJSONObjects_ReplaceSentinelWithNull(t *testing.T) {
	base := []byte(`{"a":{"x":1}}`)
	overlay := []byte(`{"a":{"__replace__":null}}`)

	merged, err := MergeJSONObjects(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONObjects() error = %v", err)
	}
	if !strings.Contains(string(merged), `"a": null`) {
		t.Fatalf("expected a to be null after __replace__; got: %s", merged)
	}
}

func TestAdversarial_UpsertCodexEngramBlock_EmptyCmd(t *testing.T) {
	// Empty engramCmd must fall back to "engram".
	result := UpsertCodexEngramBlock("", "")
	if !strings.Contains(result, `command = "engram"`) {
		t.Fatalf("expected fallback to 'engram'; got:\n%s", result)
	}
}

func TestAdversarial_UpsertTopLevelTOMLString_KeyWithEqualsInValue(t *testing.T) {
	// Value containing = must be quoted correctly without breaking TOML.
	result := UpsertTopLevelTOMLString("", "instructions", "key=value pairs matter")
	if !strings.Contains(result, `instructions = "key=value pairs matter"`) {
		t.Fatalf("value not properly quoted; got:\n%s", result)
	}
}

func TestAdversarial_UpsertTopLevelTOMLString_NReinjectionIdempotent(t *testing.T) {
	base := UpsertTopLevelTOMLString("", "model_instructions_file", "/path/file.md")
	for i := 0; i < 5; i++ {
		next := UpsertTopLevelTOMLString(base, "model_instructions_file", "/path/file.md")
		if next != base {
			t.Fatalf("UpsertTopLevelTOMLString not idempotent at iteration %d:\nbefore: %q\nafter:  %q", i+1, base, next)
		}
		base = next
	}
	if count := strings.Count(base, "model_instructions_file"); count != 1 {
		t.Fatalf("expected 1 occurrence, got %d", count)
	}
}
