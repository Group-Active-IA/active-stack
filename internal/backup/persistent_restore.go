package backup

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// restoreProtected is opt-in for manifests recording expected installed bytes.
// Legacy manifests continue using the explicit restore behavior of Restore.
func (s RestoreService) restoreProtected(manifest Manifest) error {
	if manifest.RestoreGuardVersion != 1 {
		return fmt.Errorf("unsupported restore guard version %d", manifest.RestoreGuardVersion)
	}
	files := make(map[string]ManifestEntry)
	var directories []string
	for _, entry := range manifest.Entries {
		path := filepath.Clean(entry.OriginalPath)
		if entry.IsDir {
			info, err := os.Lstat(path)
			if err != nil && !(os.IsNotExist(err) && !entry.Existed) {
				return fmt.Errorf("preserved directory \"%s\" because its state changed; backup at \"%s\"", path, manifest.RootDir)
			}
			if err == nil && !info.IsDir() {
				return fmt.Errorf("preserved replaced directory \"%s\"; backup at \"%s\"", path, manifest.RootDir)
			}
			if !entry.Existed {
				directories = append(directories, path)
			}
			continue
		}
		if _, ok := manifest.RestoreGuards[path]; !ok {
			return fmt.Errorf("missing persistent restore guard for \"%s\"; backup at \"%s\"", path, manifest.RootDir)
		}
		files[path] = entry
	}
	for path := range manifest.RestoreGuards {
		if _, ok := files[path]; !ok {
			return fmt.Errorf("persistent restore guard \"%s\" has no file entry", path)
		}
	}
	if manifest.Compressed {
		tempDir, err := os.MkdirTemp("", "active-stack-protected-restore-*")
		if err != nil {
			return fmt.Errorf("create restore staging: %w", err)
		}
		defer os.RemoveAll(tempDir)
		if _, err := ExtractArchive(filepath.Join(manifest.RootDir, ArchiveFilename), tempDir); err != nil {
			return fmt.Errorf("extract protected restore: %w", err)
		}
		for path, entry := range files {
			if !entry.Existed {
				continue
			}
			relative := filepath.FromSlash(entry.SnapshotPath)
			if filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				return fmt.Errorf("invalid snapshot path \"%s\"", entry.SnapshotPath)
			}
			entry.SnapshotPath = filepath.Join(tempDir, relative)
			files[path] = entry
		}
	}
	var result error
	for _, original := range manifest.Entries {
		if original.IsDir {
			continue
		}
		path := filepath.Clean(original.OriginalPath)
		entry := files[path]
		info, err := os.Lstat(path)
		if os.IsNotExist(err) && !entry.Existed {
			continue
		}
		if err != nil || !info.Mode().IsRegular() {
			result = errors.Join(result, fmt.Errorf("preserved \"%s\" because its state changed; backup at \"%s\"", path, manifest.RootDir))
			continue
		}
		current, err := os.ReadFile(path)
		if err != nil {
			result = errors.Join(result, fmt.Errorf("read current file \"%s\": %w", path, err))
			continue
		}
		if entry.Existed {
			prior, readErr := os.ReadFile(entry.SnapshotPath)
			if readErr != nil {
				return errors.Join(result, fmt.Errorf("read snapshot \"%s\": %w", entry.SnapshotPath, readErr))
			}
			if sha256.Sum256(current) == sha256.Sum256(prior) {
				continue
			}
		}
		if sha256.Sum256(current) != manifest.RestoreGuards[path] {
			result = errors.Join(result, fmt.Errorf("preserved \"%s\" because it changed after installation; backup at \"%s\"", path, manifest.RootDir))
			continue
		}
		if entry.Existed {
			err = restoreEntry(entry, true)
		} else {
			err = os.Remove(path)
		}
		if err != nil {
			return errors.Join(result, err)
		}
	}
	// Remove created directories only after their managed files, deepest first.
	// Remove, unlike RemoveAll, preserves new files the user placed in the tree.
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) > len(directories[j]) })
	for _, path := range directories {
		info, statErr := os.Lstat(path)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil || !info.IsDir() {
			result = errors.Join(result, fmt.Errorf("preserved replaced directory \"%s\"; backup at \"%s\"", path, manifest.RootDir))
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			result = errors.Join(result, fmt.Errorf("preserved directory \"%s\": %w; backup at \"%s\"", path, err, manifest.RootDir))
		}
	}
	return result
}
