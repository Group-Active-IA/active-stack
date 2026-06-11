// Package main — C-32 entry-point wiring test.
//
// RED: This file fails to compile on current HEAD because wireEmbeddedFS()
// does not exist. Once it exists but does NOT call WithEmbeddedCommandsFS,
// the runtime assertion fails. GREEN: passes after wireEmbeddedFS() calls
// install.WithEmbeddedCommandsFS(assets.CommandsFS).
//
// Design: design.md D1 (Option A), D2 (real-wiring test strategy).
// The canonical seam is the GetEmbeddedCommandsFS / ResetEmbeddedCommandsFS
// pair in internal/install/testseams.go.
package main

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/install"
)

// TestWireEmbeddedFS_SetsCommandsFS is the real entry-point wiring test.
//
// It calls wireEmbeddedFS() — the same function main() calls before any
// dispatch — and asserts that embeddedCommandsFS is non-nil afterwards.
//
// RED: fails to compile because wireEmbeddedFS() does not exist yet.
// After wireEmbeddedFS() is added without the WithEmbeddedCommandsFS call,
// the nil assertion below fails at runtime.
// GREEN: passes after wireEmbeddedFS() calls
// install.WithEmbeddedCommandsFS(assets.CommandsFS).
func TestWireEmbeddedFS_SetsCommandsFS(t *testing.T) {
	// Start from the nil / cold-start state.
	install.ResetEmbeddedCommandsFS()
	t.Cleanup(install.ResetEmbeddedCommandsFS)

	// Pre-condition: embeddedCommandsFS must be nil before wiring.
	if got := install.GetEmbeddedCommandsFS(); got != nil {
		t.Fatalf("pre-condition: embeddedCommandsFS must be nil before wireEmbeddedFS(); got %T", got)
	}

	// Act: call the binary entry-point wiring function.
	wireEmbeddedFS()

	// Assert: embeddedCommandsFS must now be non-nil.
	if got := install.GetEmbeddedCommandsFS(); got == nil {
		t.Fatal("embeddedCommandsFS is nil after wireEmbeddedFS() — " +
			"the function must call install.WithEmbeddedCommandsFS(assets.CommandsFS)")
	}
}
