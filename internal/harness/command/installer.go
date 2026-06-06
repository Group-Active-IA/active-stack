package command

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
)

// Installer materializes slash-command files from the embedded CommandsFS into
// each focused agent's command directory. It mirrors internal/harness/skill.Installer.
//
// Governance ALTO: snapshot before write (backup); idempotent content-hash skip.
type Installer struct {
	commandsFS fs.FS
}

// NewInstaller creates a new command Installer.
// commandsFS is the embedded FS (assets.CommandsFS) containing the per-agent
// command .md files under commands/<variantKey>/...
func NewInstaller(commandsFS fs.FS) *Installer {
	return &Installer{commandsFS: commandsFS}
}

// Install materializes the command file for each adapter that provides a
// non-empty command directory. An empty CommandsDir() skips the adapter silently.
//
// Parameters:
//   - adapters: per-agent adapters; adapters with empty CommandsDir are skipped.
//   - homeDir: user's home directory, passed to adapter.CommandsDir.
//   - backupDir: parent directory for pre-overwrite backup snapshots.
//
// The per-agent asset path is derived from adapter.VariantKey():
//
//	commands/<variantKey>/<relPath>
//
// where relPath is determined by the per-agent naming convention
// (see assetAndRelPathForVariant).
func (ins *Installer) Install(
	adapters []AgentAdapter,
	homeDir, backupDir string,
) ([]Result, error) {
	if ins.commandsFS == nil {
		return nil, fmt.Errorf("command installer: commandsFS is nil")
	}

	var results []Result
	for _, adapter := range adapters {
		commandsDir := adapter.CommandsDir(homeDir)
		if commandsDir == "" {
			// This adapter does not support commands — skip silently.
			continue
		}

		result, err := ins.installForAdapter(adapter, commandsDir, backupDir)
		if err != nil {
			return nil, fmt.Errorf("agent %q: %w", adapter.Agent(), err)
		}
		results = append(results, result)
	}
	return results, nil
}

// installForAdapter reads the embedded asset for the adapter's variant and
// writes it idempotently (with backup-before-overwrite) to the commands directory.
func (ins *Installer) installForAdapter(
	adapter AgentAdapter,
	commandsDir, backupDir string,
) (Result, error) {
	assetPath, relPath := assetAndRelPathForVariant(adapter.VariantKey())
	if assetPath == "" {
		// Unknown variant — no asset available; skip silently.
		return Result{}, nil
	}

	content, err := fs.ReadFile(ins.commandsFS, assetPath)
	if err != nil {
		return Result{}, fmt.Errorf("command: embedded asset not found at %q: %w", assetPath, err)
	}

	destFile := filepath.Join(commandsDir, relPath)

	// Idempotence check: skip if content is byte-identical.
	identical, err := checkIdempotent(destFile, content)
	if err != nil {
		return Result{}, err
	}
	if identical {
		return Result{CommandPath: destFile, AlreadyInstalled: true}, nil
	}

	// Backup existing file if it exists (governance ALTO).
	if _, statErr := os.Stat(destFile); statErr == nil {
		if backupErr := snapshotCommandFile(backupDir, destFile, adapter.VariantKey()); backupErr != nil {
			return Result{}, backupErr
		}
	}

	// Write the command file.
	if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
		return Result{}, fmt.Errorf("command: mkdir %q: %w", filepath.Dir(destFile), err)
	}
	if err := os.WriteFile(destFile, content, 0o644); err != nil {
		return Result{}, fmt.Errorf("command: write %q: %w", destFile, err)
	}

	return Result{CommandPath: destFile}, nil
}

// RelPathForVariant returns the relative file path (under commandsDir) for
// the given agent variant key. This is the single authoritative source of the
// per-agent naming convention, used by both the command installer and the
// uninstall engine (DRY — one place to change when naming evolves):
//
//	claude   → "jr/starter-add.md"  (namespaced; invoked as /jr:starter-add)
//	opencode → "jr-starter-add.md"  (flat, hyphenated; invoked as /jr-starter-add)
//
// An unknown variantKey returns "" — the caller must skip silently.
func RelPathForVariant(variantKey string) string {
	switch variantKey {
	case "claude":
		return filepath.Join("jr", "starter-add.md")
	case "opencode":
		return "jr-starter-add.md"
	default:
		return ""
	}
}

// assetAndRelPathForVariant returns the embedded FS asset path and the
// relative path (under commandsDir) for the given agent variant key.
//
// The asset path follows the structure: commands/<variantKey>/<relPath>.
// Delegates to RelPathForVariant for the relative path (single source of truth).
//
// An unknown variantKey returns ("", "") — the caller skips silently.
func assetAndRelPathForVariant(variantKey string) (assetPath, relPath string) {
	relPath = RelPathForVariant(variantKey)
	if relPath == "" {
		return "", ""
	}
	// The asset path mirrors: commands/<variantKey>/<relPath>, but uses
	// forward slashes (embedded FS always uses /).
	switch variantKey {
	case "claude":
		assetPath = "commands/claude/jr/starter-add.md"
	case "opencode":
		assetPath = "commands/opencode/jr-starter-add.md"
	}
	return assetPath, relPath
}

// snapshotCommandFile backs up the existing command file at destFile into a
// timestamped sub-directory inside backupDir, before the file is overwritten.
// Mirrors snapshotSkillDir from internal/harness/skill/idempotent.go.
func snapshotCommandFile(backupDir, destFile, variantKey string) error {
	snapName := fmt.Sprintf("%s-cmd-%s", time.Now().UTC().Format("20060102-150405"), variantKey)
	snapDir := filepath.Join(backupDir, snapName)

	s := backup.NewSnapshotter()
	if _, err := s.Create(snapDir, []string{destFile}); err != nil {
		return fmt.Errorf("command: snapshot %q: %w", destFile, err)
	}
	return nil
}
