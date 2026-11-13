package backup

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func TestRestorePreservesChangesAfterInstall(t *testing.T) {
	for _, current := range []string{"installed", "user edit"} {
		t.Run(current, func(t *testing.T) {
			dir := t.TempDir()
			config := filepath.Join(dir, "config.toml")
			snapshot := filepath.Join(dir, "snapshot.toml")
			for path, contents := range map[string]string{config: current, snapshot: "original"} {
				if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			manifest := Manifest{
				RootDir: dir,
				Entries: []ManifestEntry{{OriginalPath: config, SnapshotPath: snapshot, Existed: true}},
			}
			manifest.GuardRollback(config, sha256.Sum256([]byte("installed")))
			want := "original"
			if current == "user edit" {
				want = current
			}
			err := (RestoreService{}).Restore(manifest)
			if (err != nil) != (current == "user edit") {
				t.Fatalf("unexpected rollback result: %v", err)
			}
			raw, err := os.ReadFile(config)
			if err != nil || string(raw) != want {
				t.Fatalf("config = %q, want %q; read error: %v", raw, want, err)
			}
		})
	}
}
