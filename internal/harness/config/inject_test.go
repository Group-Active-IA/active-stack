package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/config"
)

// TestInject_FirstInjection verifies that Inject appends the sdd-orchestrator
// section when the target file does not yet contain it.
func TestInject_FirstInjection(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")

	existing := "# My Config\n\nSome existing content.\n"
	if err := os.WriteFile(target, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	composed := "# SDD Orchestrator\n\nOrchestrator block.\n"
	snapshotDir := filepath.Join(dir, "backups")

	wr, err := config.Inject(target, composed, snapshotDir)
	if err != nil {
		t.Fatalf("Inject error: %v", err)
	}
	if !wr.Changed {
		t.Error("first injection should report Changed=true")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "<!-- jr-stack:sdd-orchestrator -->") {
		t.Error("injected file should contain opening marker")
	}
	if !strings.Contains(string(content), "<!-- /jr-stack:sdd-orchestrator -->") {
		t.Error("injected file should contain closing marker")
	}
	if !strings.Contains(string(content), "Orchestrator block.") {
		t.Error("injected file should contain composed content")
	}
	if !strings.Contains(string(content), "Some existing content.") {
		t.Error("content outside markers must be preserved")
	}
}

// TestInject_Idempotent verifies that re-injecting the same content is a no-op.
func TestInject_Idempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")

	existing := "# My Config\n\nSome existing content.\n"
	if err := os.WriteFile(target, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	composed := "# SDD Orchestrator\n\nOrchestrator block.\n"
	snapshotDir := filepath.Join(dir, "backups")

	// First injection.
	if _, err := config.Inject(target, composed, snapshotDir); err != nil {
		t.Fatalf("first inject: %v", err)
	}

	// Read after first injection.
	after1, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	// Second injection with same content.
	wr, err := config.Inject(target, composed, snapshotDir)
	if err != nil {
		t.Fatalf("second inject: %v", err)
	}
	if wr.Changed {
		t.Error("re-injection of byte-identical content must NOT report Changed")
	}

	after2, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(after1) != string(after2) {
		t.Error("idempotent re-injection must leave file byte-identical")
	}
}

// TestInject_ReplacesExistingSection verifies that re-injecting with different
// content replaces the section in-place without duplication.
func TestInject_ReplacesExistingSection(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")

	existing := "# My Config\n\nSome existing content.\n"
	if err := os.WriteFile(target, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	snapshotDir := filepath.Join(dir, "backups")

	// First injection.
	if _, err := config.Inject(target, "# Old Block\n", snapshotDir); err != nil {
		t.Fatalf("first inject: %v", err)
	}

	// Update with new content.
	wr, err := config.Inject(target, "# New Block\n", snapshotDir)
	if err != nil {
		t.Fatalf("second inject: %v", err)
	}
	if !wr.Changed {
		t.Error("update injection should report Changed=true")
	}

	content, _ := os.ReadFile(target)
	cs := string(content)

	// Only one opening marker — no duplication.
	count := strings.Count(cs, "<!-- jr-stack:sdd-orchestrator -->")
	if count != 1 {
		t.Errorf("marker count = %d, want 1 (no duplication)", count)
	}
	if strings.Contains(cs, "# Old Block") {
		t.Error("old block should be replaced")
	}
	if !strings.Contains(cs, "# New Block") {
		t.Error("new block should be present")
	}
}

// TestInject_BackupCreated verifies that a backup is created before writing
// when the target file already exists.
func TestInject_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	snapshotDir := filepath.Join(dir, "backups")

	if err := os.WriteFile(target, []byte("# Existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Inject(target, "# New content\n", snapshotDir); err != nil {
		t.Fatalf("inject: %v", err)
	}

	// Backup directory should have been created.
	info, err := os.Stat(snapshotDir)
	if err != nil {
		t.Fatalf("backup dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("snapshot dir should be a directory")
	}
}

// TestInject_BackupFailureAborts verifies that if backup fails, Inject returns
// an error without touching the target file.
func TestInject_BackupFailureAborts(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	original := "# Original content\n"

	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	// Swap in a snapshotter that always fails.
	prev := config.SnapshotterCreate
	config.SnapshotterCreate = func(snapshotDir string, paths []string) error {
		return errors.New("simulated backup failure")
	}
	defer func() { config.SnapshotterCreate = prev }()

	_, err := config.Inject(target, "# Modified content\n", filepath.Join(dir, "backups"))
	if err == nil {
		t.Fatal("Inject should return error when backup fails")
	}

	// File must be untouched.
	content, _ := os.ReadFile(target)
	if string(content) != original {
		t.Error("target file must not be modified when backup fails")
	}
}

// TestInject_EmptyPath verifies that Inject with an empty path is a no-op.
func TestInject_EmptyPath(t *testing.T) {
	wr, err := config.Inject("", "# content\n", "/tmp/backups")
	if err != nil {
		t.Fatalf("Inject with empty path should not error: %v", err)
	}
	if wr.Changed {
		t.Error("Inject with empty path should not report Changed")
	}
}

// TestInject_ContentOutsideMarkersPreserved verifies that content before and
// after the sdd-orchestrator section is never touched.
func TestInject_ContentOutsideMarkersPreserved(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "CLAUDE.md")
	snapshotDir := filepath.Join(dir, "backups")

	original := "# Header\n\nBefore block.\n"
	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Inject(target, "# Block\n", snapshotDir); err != nil {
		t.Fatalf("inject: %v", err)
	}

	content, _ := os.ReadFile(target)
	cs := string(content)
	if !strings.Contains(cs, "Before block.") {
		t.Error("content before the injected block must be preserved")
	}
	if !strings.Contains(cs, "# Header") {
		t.Error("header before the injected block must be preserved")
	}
}
