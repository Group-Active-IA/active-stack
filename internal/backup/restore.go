package backup

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

type RestoreService struct{}

type rollbackGuard struct {
	expected [sha256.Size]byte
	restored bool
}

// GuardRollback associates a live install snapshot with the bytes written by
// the installer. A manually loaded snapshot has no live rollback guards.
func (m *Manifest) GuardRollback(path string, expected [sha256.Size]byte) {
	if m.rollbackGuards == nil {
		m.rollbackGuards = make(map[string]*rollbackGuard)
	}
	m.rollbackGuards[filepath.Clean(path)] = &rollbackGuard{expected: expected}
}

func (s RestoreService) Restore(manifest Manifest) error {
	entries := manifest.Entries
	var conflicts error
	if len(manifest.rollbackGuards) != 0 {
		entries = make([]ManifestEntry, 0, len(manifest.Entries))
		excluded := make(map[string]bool)
		for path, guard := range manifest.rollbackGuards {
			raw, err := os.ReadFile(path)
			if !guard.restored && (err != nil || sha256.Sum256(raw) != guard.expected) {
				conflicts = errors.Join(conflicts, fmt.Errorf("preserved %q because it changed after installation; backup at %q", path, manifest.RootDir))
				excluded[path] = true
			} else if guard.restored {
				excluded[path] = true
			}
		}
		for _, entry := range manifest.Entries {
			path := filepath.Clean(entry.OriginalPath)
			skip := excluded[path]
			if entry.IsDir && !entry.Existed {
				for protected := range excluded {
					skip = skip || strings.HasPrefix(protected, path+string(filepath.Separator))
				}
			}
			if !skip {
				entries = append(entries, entry)
			}
		}
	}
	manifest.Entries = entries
	var err error
	if manifest.Compressed {
		err = s.restoreCompressed(manifest)
	} else {
		err = s.restorePlain(manifest)
	}
	if err == nil {
		for _, entry := range entries {
			if guard := manifest.rollbackGuards[filepath.Clean(entry.OriginalPath)]; guard != nil {
				guard.restored = true
			}
		}
	}
	return errors.Join(conflicts, err)
}

// restoreCompressed handles backups where Compressed==true.
// It extracts the tar.gz archive into a temp directory, then restores each
// entry by resolving the relative SnapshotPath inside that temp directory.
func (s RestoreService) restoreCompressed(manifest Manifest) error {
	tempDir, err := os.MkdirTemp("", "active-stack-restore-*")
	if err != nil {
		return fmt.Errorf("create temp restore dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(manifest.RootDir, ArchiveFilename)
	if _, err := ExtractArchive(archivePath, tempDir); err != nil {
		return fmt.Errorf("extract archive %q: %w", archivePath, err)
	}

	for _, entry := range manifest.Entries {
		switch {
		case entry.Existed && entry.IsDir:
			// Preexisting directory → NO-OP.
			// SAFETY: never remove a directory that existed before the install.
			// Its individual files are restored by their own ManifestEntry records.

		case entry.Existed && !entry.IsDir:
			// Preexisting file → restore content from archive.
			// SnapshotPath must be relative inside the archive (e.g. "files/.config/foo.json").
			// An absolute path would cause filepath.Join to ignore tempDir, reading from
			// the live filesystem instead of the extraction directory.
			if filepath.IsAbs(entry.SnapshotPath) {
				return fmt.Errorf("manifest entry %q has absolute SnapshotPath %q, expected relative", entry.OriginalPath, entry.SnapshotPath)
			}
			resolvedEntry := ManifestEntry{
				OriginalPath: entry.OriginalPath,
				SnapshotPath: filepath.Join(tempDir, filepath.FromSlash(entry.SnapshotPath)),
				Existed:      true,
				Mode:         entry.Mode,
			}
			if err := restoreEntry(resolvedEntry); err != nil {
				return err
			}

		case !entry.Existed && entry.IsDir:
			// Directory created by the install → remove the whole tree.
			// os.RemoveAll handles non-empty directories and ignores IsNotExist.
			if err := os.RemoveAll(entry.OriginalPath); err != nil {
				return fmt.Errorf("remove-all path %q: %w", entry.OriginalPath, err)
			}

		default:
			// !entry.Existed && !entry.IsDir: file created by the install → remove it.
			if err := os.Remove(entry.OriginalPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove path %q: %w", entry.OriginalPath, err)
			}
		}
	}

	return nil
}

// restorePlain handles old-style backups where Compressed==false.
// SnapshotPath is an absolute path to a plain file on disk.
func (s RestoreService) restorePlain(manifest Manifest) error {
	for _, entry := range manifest.Entries {
		switch {
		case entry.Existed && entry.IsDir:
			// Preexisting directory → NO-OP (SAFETY: preserve user config).

		case entry.Existed && !entry.IsDir:
			// Preexisting file → restore content.
			if err := restoreEntry(entry); err != nil {
				return err
			}

		case !entry.Existed && entry.IsDir:
			// Directory created by the install → remove the whole tree.
			if err := os.RemoveAll(entry.OriginalPath); err != nil {
				return fmt.Errorf("remove-all path %q: %w", entry.OriginalPath, err)
			}

		default:
			// !entry.Existed && !entry.IsDir: file created by the install → remove it.
			if err := os.Remove(entry.OriginalPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove path %q: %w", entry.OriginalPath, err)
			}
		}
	}

	return nil
}

func restoreEntry(entry ManifestEntry) error {
	content, err := os.ReadFile(entry.SnapshotPath)
	if err != nil {
		return fmt.Errorf("read snapshot file %q: %w", entry.SnapshotPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(entry.OriginalPath), 0o755); err != nil {
		return fmt.Errorf("create restore directory for %q: %w", entry.OriginalPath, err)
	}

	if _, err := filemerge.WriteFileAtomic(entry.OriginalPath, content, os.FileMode(entry.Mode)); err != nil {
		return fmt.Errorf("restore path %q: %w", entry.OriginalPath, err)
	}

	return nil
}
