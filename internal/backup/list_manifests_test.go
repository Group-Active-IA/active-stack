package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestListManifests_ReturnsManifestsFromDir verifies that ListManifests
// returns the backups present in the given directory.
func TestListManifests_ReturnsManifestsFromDir(t *testing.T) {
	dir := t.TempDir()
	// Create two backup subdirectories each with a manifest.
	for _, id := range []string{"backup-001", "backup-002"} {
		subDir := filepath.Join(dir, id)
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("mkdir %v: %v", subDir, err)
		}
		m := Manifest{
			ID:        id,
			CreatedAt: time.Now(),
			RootDir:   subDir,
		}
		data, _ := json.MarshalIndent(m, "", "  ")
		if err := os.WriteFile(filepath.Join(subDir, ManifestFilename), data, 0o644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
	}

	manifests, err := ListManifests(dir)
	if err != nil {
		t.Fatalf("ListManifests error: %v", err)
	}
	if len(manifests) != 2 {
		t.Errorf("len(manifests) = %d, want 2", len(manifests))
	}
}

// TestListManifests_EmptyDirReturnsEmpty verifies that ListManifests on an
// empty (or non-existent) directory returns empty slice with no error.
func TestListManifests_EmptyDirReturnsEmpty(t *testing.T) {
	dir := t.TempDir() // empty directory

	manifests, err := ListManifests(dir)
	if err != nil {
		t.Fatalf("ListManifests error on empty dir: %v", err)
	}
	if len(manifests) != 0 {
		t.Errorf("len(manifests) = %d, want 0", len(manifests))
	}
}

// TestListManifests_NonexistentDirReturnsEmpty verifies graceful handling
// of a non-existent directory (listManifests returns nil, nil for IsNotExist).
func TestListManifests_NonexistentDirReturnsEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	manifests, err := ListManifests(dir)
	if err != nil {
		t.Fatalf("ListManifests error on nonexistent dir: %v", err)
	}
	if manifests != nil && len(manifests) != 0 {
		t.Errorf("expected empty result, got %d manifests", len(manifests))
	}
}
