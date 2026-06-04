package pipeline

// C-32: distinct degraded status — task 1.1 RED → 1.2 GREEN

import "testing"

// TestStepStatusDegraded_Exists asserts that the distinct degraded constant exists
// and is separate from both StepStatusFailed and StepStatusSucceeded.
func TestStepStatusDegraded_Exists(t *testing.T) {
	if StepStatusDegraded == "" {
		t.Fatal("StepStatusDegraded must be a non-empty constant")
	}
	if StepStatusDegraded == StepStatusFailed {
		t.Errorf("StepStatusDegraded must differ from StepStatusFailed (got %q == %q)", StepStatusDegraded, StepStatusFailed)
	}
	if StepStatusDegraded == StepStatusSucceeded {
		t.Errorf("StepStatusDegraded must differ from StepStatusSucceeded (got %q == %q)", StepStatusDegraded, StepStatusSucceeded)
	}
}

// TestStepStatusDegraded_StringValue asserts the wire value is "degraded"
// (used in progress event filtering in headless/TUI — must be a stable, human-readable string).
func TestStepStatusDegraded_StringValue(t *testing.T) {
	if StepStatusDegraded != "degraded" {
		t.Errorf("StepStatusDegraded = %q, want %q", StepStatusDegraded, "degraded")
	}
}
