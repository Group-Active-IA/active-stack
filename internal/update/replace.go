package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func copyBinaryFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(dst)
		return closeErr
	}
	return nil
}

func binaryDigest(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}
	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum, nil
}

// BinaryMatches compares two executables by SHA256 without modifying either.
func BinaryMatches(a, b string) (bool, error) {
	first, err := binaryDigest(a)
	if err != nil {
		return false, err
	}
	second, err := binaryDigest(b)
	if err != nil {
		return false, err
	}
	return first == second, nil
}

// ReplaceBinary snapshots the previous executable, stages the verified payload
// alongside the target, then activates it with a rename. It restores the old
// executable if activation fails. An unchanged binary is left untouched.
func ReplaceBinary(verifiedPath, target, backupDir string) (bool, error) {
	verifiedHash, err := binaryDigest(verifiedPath)
	if err != nil {
		return false, fmt.Errorf("read verified binary: %w", err)
	}
	info, err := os.Lstat(target)
	exists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if exists && !info.Mode().IsRegular() {
		return false, fmt.Errorf("binary target %q is not a regular file", target)
	}
	if exists {
		currentHash, err := binaryDigest(target)
		if err != nil {
			return false, err
		}
		if currentHash == verifiedHash {
			return false, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return false, err
	}
	stageFile, err := os.CreateTemp(filepath.Dir(target), ".active-stack-bin-*")
	if err != nil {
		return false, err
	}
	stage := stageFile.Name()
	defer os.Remove(stage)
	source, err := os.Open(verifiedPath)
	if err != nil {
		stageFile.Close()
		return false, err
	}
	_, copyErr := io.Copy(stageFile, source)
	sourceErr := source.Close()
	if copyErr != nil {
		stageFile.Close()
		return false, fmt.Errorf("stage binary: %w", copyErr)
	}
	if sourceErr != nil {
		stageFile.Close()
		return false, sourceErr
	}
	if err := stageFile.Chmod(0o755); err != nil {
		stageFile.Close()
		return false, err
	}
	if err := stageFile.Sync(); err != nil {
		stageFile.Close()
		return false, err
	}
	if err := stageFile.Close(); err != nil {
		return false, err
	}
	if exists {
		if err := os.MkdirAll(backupDir, 0o755); err != nil {
			return false, err
		}
		backup := filepath.Join(backupDir, fmt.Sprintf("%s-%s", time.Now().UTC().Format("20060102-150405.000000000"), filepath.Base(target)))
		if err := copyBinaryFile(target, backup); err != nil {
			return false, fmt.Errorf("back up binary: %w", err)
		}
	}
	old := stage + ".old"
	if exists {
		if err := os.Rename(target, old); err != nil {
			return false, fmt.Errorf("move old binary: %w", err)
		}
	}
	if err := os.Rename(stage, target); err != nil {
		if exists {
			if restoreErr := os.Rename(old, target); restoreErr != nil {
				return false, fmt.Errorf("activate binary: %w; restore old binary: %v", err, restoreErr)
			}
		}
		return false, fmt.Errorf("activate binary: %w", err)
	}
	if exists {
		_ = os.Remove(old)
	}
	return true, nil
}
