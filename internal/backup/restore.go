package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

type RestoreService struct{}

func (s RestoreService) Restore(manifest Manifest) error {
	if manifest.Compressed {
		return s.restoreCompressed(manifest)
	}
	return s.restorePlain(manifest)
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
