package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func checkEngramHooks(ctx context.Context, binary string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for _, hook := range []string{"codex-user-prompt-submit", "claude-pre-tool-use"} {
		out, err := runner(ctx, binary, "hook", hook)
		if err != nil {
			return fmt.Errorf("%s: %w", hook, err)
		}
		if !json.Valid(out) {
			return fmt.Errorf("%s returned invalid JSON", hook)
		}
	}
	return nil
}

func publishEngram(candidate, installed string, existing bool) error {
	if !existing {
		return os.Rename(candidate, installed)
	}

	backupDir, err := os.MkdirTemp(filepath.Dir(installed), ".engram-backup-")
	if err != nil {
		return err
	}
	backup := filepath.Join(backupDir, filepath.Base(installed))
	if err := os.Rename(installed, backup); err != nil {
		_ = os.Remove(backupDir)
		return fmt.Errorf("back up Engram before upgrading; close clients using it and retry: %w", err)
	}
	if err := os.Rename(candidate, installed); err != nil {
		if restoreErr := os.Rename(backup, installed); restoreErr != nil {
			return errors.Join(err, fmt.Errorf("restore Engram from %s: %w", backup, restoreErr))
		}
		_ = os.Remove(backupDir)
		return fmt.Errorf("replace Engram: %w", err)
	}
	return nil
}
