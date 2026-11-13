package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// MigrationRecord registers the first migration from the legacy block to the
// layered delivery (claude-native-opsx D5 step 7).
type MigrationRecord struct {
	ManifestDir        string            `json:"manifest_dir"`
	InstructionsPath   string            `json:"instructions_path"`
	InstructionsSHA256 string            `json:"instructions_sha256"`
	SkillFiles         map[string]string `json:"skill_files"`
}

// SHA256Hex returns the hex SHA-256 of b.
func SHA256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type migrationFile struct {
	Version int                        `json:"version"`
	Records map[string]MigrationRecord `json:"records"`
}

// MigrationRecordPath is the sibling file of installation.json that holds the
// migration records (design D5 step 7; location fixed by task 6.8).
func MigrationRecordPath(home string) string {
	return filepath.Join(home, ".active-stack", "claude-migration.json")
}

func loadMigrationFile(home string) (migrationFile, error) {
	data, err := os.ReadFile(MigrationRecordPath(home))
	if os.IsNotExist(err) {
		return migrationFile{Version: 1, Records: map[string]MigrationRecord{}}, nil
	}
	if err != nil {
		return migrationFile{}, err
	}
	var f migrationFile
	if err := json.Unmarshal(data, &f); err != nil {
		return migrationFile{}, fmt.Errorf("parse %s: %w", MigrationRecordPath(home), err)
	}
	if f.Records == nil {
		f.Records = map[string]MigrationRecord{}
	}
	return f, nil
}

// LoadMigrationRecord returns the record for an instructions file.
func LoadMigrationRecord(home, instructionsPath string) (MigrationRecord, bool, error) {
	f, err := loadMigrationFile(home)
	if err != nil {
		return MigrationRecord{}, false, err
	}
	r, ok := f.Records[filepath.Clean(instructionsPath)]
	return r, ok, nil
}

// recordMigrationOnce registers the first migration of instructionsPath: the
// manifest path, the hashes of CLAUDE.md and of every skill file, and pins the
// manifest so Prune never removes it. An existing record is never replaced.
func recordMigrationOnce(home, manifestDir, instructionsPath, skillDir string) error {
	f, err := loadMigrationFile(home)
	if err != nil {
		return err
	}
	key := filepath.Clean(instructionsPath)
	if _, exists := f.Records[key]; exists {
		return nil
	}
	raw, err := os.ReadFile(instructionsPath)
	if err != nil {
		return err
	}
	rec := MigrationRecord{ManifestDir: manifestDir, InstructionsPath: key, InstructionsSHA256: SHA256Hex(raw), SkillFiles: map[string]string{}}
	hashes, err := fileHashes(skillDir)
	if err != nil {
		return err
	}
	for p, h := range hashes {
		rec.SkillFiles[p] = hex.EncodeToString(h[:])
	}
	mpath := filepath.Join(manifestDir, backup.ManifestFilename)
	m, err := backup.ReadManifest(mpath)
	if err != nil {
		return err
	}
	m.Pinned = true
	if err := backup.WriteManifest(mpath, m); err != nil {
		return err
	}
	f.Records[key] = rec
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(MigrationRecordPath(home)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(MigrationRecordPath(home), append(out, '\n'), 0o644)
}
