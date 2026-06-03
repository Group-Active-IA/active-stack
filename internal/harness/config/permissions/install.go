package permissions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// snapshotterCreate is the function used to create a backup snapshot.
// It is a package-level variable so tests can inject a failing implementation
// to verify that backup failures abort the write (task 4.8).
var snapshotterCreate = func(snapshotDir string, paths []string) error {
	_, err := backup.NewSnapshotter().Create(snapshotDir, paths)
	return err
}

// SetSnapshotterCreate replaces the snapshotter function for testing and
// returns a restore function that resets it to the original.
// This is the only test-seam exposed by this package.
func SetSnapshotterCreate(fn func(snapshotDir string, paths []string) error) (restore func()) {
	old := snapshotterCreate
	snapshotterCreate = fn
	return func() { snapshotterCreate = old }
}

// osReadFile reads a file and returns its contents, or nil if the file does
// not exist. It is a package-level variable so tests can inject fakes.
var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}
	return content, nil
}

// Install applies the permission overlay for each adapter in the list.
//
// For each adapter:
//  1. Resolve the settings file path via adapter.SettingsPath(homeDir).
//  2. Determine the overlay for the agent and tier. If nil → skip without error (no-op).
//  3. If the settings file exists → create a backup snapshot FIRST.
//     If the backup fails → abort; the settings file is NOT touched.
//  4. Merge the overlay into the existing settings (idempotent deep-merge).
//  5. Write atomically; skip if byte-identical (WriteFileAtomic).
//
// The tier controls which permission level is composed. A zero-value tier
// normalizes defensively to TierBalanceado — NEVER TierBypass.
//
// Governance: ALTO. The backup is mandatory and cannot be disabled.
func Install(homeDir string, adapters []PermissionsAdapter, tier model.PermissionTier) (Result, error) {
	// Defensive normalization: never let a zero-value produce bypass behavior.
	tier = tier.Normalize()

	var result Result

	for _, adapter := range adapters {
		changed, path, err := installOne(homeDir, adapter, tier)
		if err != nil {
			return Result{}, err
		}
		if changed {
			result.Changed = true
			result.Files = append(result.Files, path)
		}
	}

	return result, nil
}

func installOne(homeDir string, adapter PermissionsAdapter, tier model.PermissionTier) (changed bool, path string, err error) {
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		// Explicit no-op: agent has no injectable settings path.
		return false, "", nil
	}

	overlay := agentOverlay(adapter.Agent(), tier)
	if overlay == nil {
		// Explicit no-op: agent does not support settings.json permission injection.
		return false, "", nil
	}

	// Backup BEFORE touching the file — only if the file already exists.
	if err := backupIfExists(homeDir, adapter.Agent(), settingsPath); err != nil {
		return false, "", err
	}

	writeResult, err := mergeJSONFile(settingsPath, overlay)
	if err != nil {
		return false, "", fmt.Errorf("permissions: merge settings for agent %q: %w", adapter.Agent(), err)
	}

	return writeResult.Changed, settingsPath, nil
}

// backupIfExists creates a snapshot of settingsPath if the file exists.
// If the backup fails, it returns an error — the caller must NOT write.
func backupIfExists(homeDir string, agent model.Agent, settingsPath string) error {
	if _, err := os.Stat(settingsPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet — nothing to back up.
			return nil
		}
		return fmt.Errorf("permissions: stat settings file %q: %w", settingsPath, err)
	}

	snapshotDir := filepath.Join(homeDir, ".jr-stack", "backups", fmt.Sprintf("permissions-%s", string(agent)))
	if err := snapshotterCreate(snapshotDir, []string{settingsPath}); err != nil {
		return fmt.Errorf("permissions: backup settings for agent %q: %w", agent, err)
	}

	return nil
}

// mergeJSONFile reads the existing settings file (or treats it as empty if it
// doesn't exist), deep-merges the overlay, and writes atomically.
func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjects(baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, fmt.Errorf("merge json objects: %w", err)
	}

	return filemerge.WriteFileAtomic(path, merged, 0o644)
}
