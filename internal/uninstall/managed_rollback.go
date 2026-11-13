package uninstall

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

// managedRollback narrows the uninstall snapshot to paths this step changed.
// Parent identities prevent a later replacement from redirecting restoration.
type managedRollback struct {
	owned   map[string]bool
	changed map[string]bool
	parents map[string]os.FileInfo
}

func newManagedRollback() *managedRollback {
	return &managedRollback{owned: map[string]bool{}, changed: map[string]bool{}, parents: map[string]os.FileInfo{}}
}

func (r *managedRollback) capture(paths []string, home string) error {
	for _, path := range paths {
		r.owned[filepath.Clean(path)] = true
		for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
			if _, exists := r.parents[dir]; !exists {
				info, err := os.Lstat(dir)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("inspect rollback parent %q: %w", dir, err)
				}
				r.parents[dir] = info
			}
			if dir == filepath.Clean(home) || dir == filepath.Dir(dir) {
				break
			}
		}
	}
	return nil
}

func (r *managedRollback) record(manifest *backup.Manifest, changed []string, instructions, original, id string) {
	for _, path := range changed {
		r.changed[filepath.Clean(path)] = true
		if manifest == nil {
			continue
		}
		if path != instructions {
			manifest.GuardRollbackAbsent(path)
			continue
		}
		original = filemerge.RemoveMarkdownSectionsPreservingContent(original, id)
		manifest.GuardRollback(path, sha256.Sum256([]byte(original)))
	}
}

func (r *managedRollback) restore(manifest backup.Manifest) error {
	var filtered []backup.ManifestEntry
	var conflicts error
	for _, entry := range manifest.Entries {
		path := filepath.Clean(entry.OriginalPath)
		if !r.owned[path] {
			filtered = append(filtered, entry)
			continue
		}
		if !r.changed[path] {
			continue
		}
		valid := true
		for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
			before, recorded := r.parents[dir]
			if !recorded {
				break
			}
			now, err := os.Lstat(dir)
			if err != nil && !os.IsNotExist(err) || err == nil && (before == nil || !os.SameFile(before, now)) {
				conflicts = errors.Join(conflicts, fmt.Errorf("preserved %q because rollback parent %q changed; backup at %q", path, dir, manifest.RootDir))
				valid = false
				break
			}
			if dir == filepath.Dir(dir) {
				break
			}
		}
		if valid {
			filtered = append(filtered, entry)
		}
	}
	manifest.Entries = filtered
	return errors.Join(conflicts, restoreFn(manifest))
}
