package config

import (
	"fmt"
	"os"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
)

// SnapshotterCreate is the function used to create a backup snapshot before
// writing to a user's instructions file. It is a package-level variable so
// tests can swap it out without touching real filesystem backups.
//
// Mirrors the pattern established in internal/harness/external/mcp.go.
var SnapshotterCreate = func(snapshotDir string, paths []string) error {
	s := backup.NewSnapshotter()
	_, err := s.Create(snapshotDir, paths)
	return err
}

// ownedSectionIDs is the registry of markdown section IDs the config installer
// manages. It is the single source of truth for the purge policy: any
// jr-stack-marked section found in a target file whose ID is NOT in this set is
// treated as stale (a leftover from a previous installer layout) and removed
// before the current block is injected.
//
// Why include the nested children — `sdd-delegation` and `sdd-model-assignments`
// live inside the `sdd-orchestrator` block (toggle-controlled). They are ours,
// so they must never be purged as standalone stale sections.
//
// When a new owned section is added in the future, register it here.
var ownedSectionIDs = map[string]bool{
	"sdd-orchestrator":      true,
	"sdd-delegation":        true,
	"sdd-model-assignments": true,
}

// PurgeStaleSections removes every jr-stack-marked section in content whose ID
// is not owned by the current installer. This cleans up legacy/renamed sections
// from older layouts (e.g. persona, engram-protocol, strict-tdd-mode) so neither
// a re-install nor an uninstall leaves orphaned, duplicated blocks behind.
//
// It is the single source of truth for the stale-section policy, shared by the
// install path (pre-injection cleanup) and the uninstall path (footprint
// teardown). Owned sections are never touched here — on uninstall the owned
// section is removed separately by its own marker-removal step.
//
// Only well-formed sections (matching open+close markers) are touched; a
// half-written marker is ignored by filemerge.MarkedSectionIDs and left intact.
func PurgeStaleSections(content string) string {
	for _, id := range filemerge.MarkedSectionIDs(content) {
		if !ownedSectionIDs[id] {
			content = filemerge.InjectMarkdownSection(content, id, "")
		}
	}
	return content
}

// InjectResult describes the outcome of a single file injection.
type InjectResult struct {
	// Changed is true when the file was written (new content differs from existing).
	Changed bool
	// Created is true when the file was created (did not exist before injection).
	Created bool
}

// Inject writes the composed content into the sdd-orchestrator section of the
// file at path, with an atomic backup-first guarantee.
//
// Flow:
//  1. Empty path → skip (no-op, no error).
//  2. File exists → create backup at snapshotDir BEFORE any write.
//  3. Read existing content (or start from empty string if file is new).
//  4. Inject the sdd-orchestrator section via filemerge.InjectMarkdownSection.
//  5. Write atomically via filemerge.WriteFileAtomic (skips if byte-identical).
//
// Backup failure → return error immediately, file is NOT touched.
func Inject(path, composed, snapshotDir string) (InjectResult, error) {
	if path == "" {
		return InjectResult{}, nil
	}

	// Backup existing file if it exists.
	if _, err := os.Stat(path); err == nil {
		if err := SnapshotterCreate(snapshotDir, []string{path}); err != nil {
			return InjectResult{}, fmt.Errorf("backup %q before injection: %w", path, err)
		}
	}

	// Read current content.
	existing := readFileOrEmpty(path)

	// Purge stale sections from older installer layouts BEFORE injecting, so a
	// re-install does not accumulate orphaned, duplicated blocks. The backup
	// above already snapshotted the pre-purge state.
	existing = PurgeStaleSections(existing)

	// Inject (or replace) the sdd-orchestrator section.
	updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", composed)

	// Write atomically (skips if byte-identical).
	wr, err := filemerge.WriteFileAtomic(path, []byte(updated), 0o644)
	if err != nil {
		return InjectResult{}, fmt.Errorf("write %q: %w", path, err)
	}

	return InjectResult{
		Changed: wr.Changed,
		Created: wr.Created,
	}, nil
}

// readFileOrEmpty reads a file's content, returning an empty string if the
// file does not exist.
func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
