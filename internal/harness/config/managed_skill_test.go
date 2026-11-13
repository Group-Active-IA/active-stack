package config

import (
	"bytes"
	"errors"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type managedTestAdapter struct{}

func (managedTestAdapter) Agent() model.Agent { return model.AgentCodex }
func (managedTestAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, "instructions", "AGENTS.md")
}
func (managedTestAdapter) SkillsDir(home string) string { return filepath.Join(home, "skills") }

func TestManagedSkillInstallAndReinstall(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	original := "user rules\n<!-- jr-stack:sdd-orchestrator -->\nold coordinator\n<!-- /jr-stack:sdd-orchestrator -->\n<!-- active-stack:design-mockups -->\nkeep mockups\n<!-- /active-stack:design-mockups -->\n"
	os.WriteFile(path, []byte(original), 0600)
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"tdd"}}
	result, err := InstallManagedSkill(h, a, home)
	if err != nil {
		t.Fatal(err)
	}
	if result.AllAlready {
		t.Fatal("first install reported unchanged")
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "old coordinator") || !strings.Contains(string(raw), "keep mockups") || !strings.Contains(string(raw), "user rules") {
		t.Fatalf("bad migration: %s", raw)
	}
	if _, err = os.Stat(filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	result, err = InstallManagedSkill(h, a, home)
	if err != nil || !result.AllAlready {
		t.Fatalf("reinstall result=%+v err=%v", result, err)
	}
}

func TestManagedSkillRejectsForeignSkillBeforeMigration(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	os.WriteFile(path, []byte("user instructions"), 0600)
	skill := filepath.Join(a.SkillsDir(home), "sdd-orchestrator")
	os.MkdirAll(skill, 0700)
	os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("user skill"), 0600)
	_, err := InstallManagedSkill(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, a, home)
	if err == nil {
		t.Fatal("foreign skill accepted")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "user instructions" {
		t.Fatal("instructions changed before preflight passed")
	}
}

func TestManagedSkillRefreshRetiresOnlyOwnedModules(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"tdd"}}
	if _, e := InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	note := filepath.Join(a.SkillsDir(home), h.ID, "modules", "user-note.md")
	os.WriteFile(note, []byte("keep"), 0600)
	h.Toggles = nil
	if _, e := InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(a.SkillsDir(home), h.ID, "modules", "tdd.md")); !os.IsNotExist(e) {
		t.Fatal("disabled module retained")
	}
	raw, _ := os.ReadFile(note)
	if string(raw) != "keep" {
		t.Fatal("unowned note changed")
	}
}

func TestManagedSkillRejectsEditedPackage(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	InstallManagedSkill(h, a, home)
	skill := filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")
	os.WriteFile(skill, []byte("my edited skill"), 0600)
	if _, e := InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("edited package overwritten")
	}
	raw, _ := os.ReadFile(skill)
	if string(raw) != "my edited skill" {
		t.Fatal("edited file lost")
	}
}

func TestManagedSkillPreservesFirstSnapshot(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	os.WriteFile(path, []byte("original user rules"), 0600)
	InstallManagedSkill(h, a, home)
	before, e := os.ReadFile(filepath.Join(migrationRoot(a, home, h.ID), "snapshot.tar.gz"))
	if e != nil {
		t.Fatal(e)
	}
	h.Toggles = []string{"tdd"}
	if _, e = InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	h.Toggles = []string{"engram"}
	if _, e = InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	after, _ := os.ReadFile(filepath.Join(migrationRoot(a, home, h.ID), "snapshot.tar.gz"))
	if !bytes.Equal(before, after) {
		t.Fatal("first snapshot replaced")
	}
	manifest, e := backup.ReadManifest(filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename))
	if e != nil {
		t.Fatal(e)
	}
	if !manifest.Pinned {
		t.Fatal("migration snapshot not pinned")
	}
	listed, e := backup.ListManifests(filepath.Join(home, ".active-stack", "backups"))
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, item := range listed {
		if item.RootDir == manifest.RootDir {
			found = true
			if !item.Pinned || item.Source != backup.BackupSourceInstall {
				t.Fatal("migration backup metadata missing from listing")
			}
		}
	}
	if !found {
		t.Fatal("migration backup unavailable in existing list")
	}

	if e = (backup.RestoreService{}).Restore(manifest); e != nil {
		t.Fatal(e)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "original user rules" {
		t.Fatalf("restore lost original: %s", raw)
	}
}

func TestManagedParentValidationOutsideHome(t *testing.T) {
	if e := checkRegularParents(filepath.Join(t.TempDir(), "outside"), t.TempDir()); e == nil {
		t.Fatal("outside home accepted")
	}
}

func TestMigrationManifestIncludesManagedParents(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, e := InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	m, e := backup.ReadManifest(filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename))
	if e != nil {
		t.Fatal(e)
	}
	root := filepath.Join(a.SkillsDir(home), h.ID)
	for _, entry := range m.Entries {
		if entry.OriginalPath == root && entry.IsDir {
			return
		}
	}
	t.Fatal("managed skill parent not recorded in migration snapshot")
}

func (managedTestAdapter) VariantKey() string              { return "codex" }
func (managedTestAdapter) SettingsPath(home string) string { return filepath.Join(home, "config.toml") }
func (managedTestAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryManagedSkill
}

func TestManagedDeliveryThroughInstall(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, e := Install(h, []AgentAdapter{a}, home); e != nil {
		t.Fatal(e)
	}
	if e := VerifyManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
}

func TestManagedSkillOperationLock(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	id := "sdd-orchestrator"
	lock := filepath.Join(home, ".active-stack", "locks", id+"-"+string(a.Agent()))
	os.MkdirAll(lock, 0700)
	if _, e := InstallManagedSkill(model.Harness{ID: id, Type: model.HarnessConfig}, a, home); e == nil {
		t.Fatal("concurrent install accepted")
	}
	if _, e := os.Stat(a.InstructionsPath(home)); !os.IsNotExist(e) {
		t.Fatal("busy install wrote instructions")
	}
}

func TestManagedSkillRejectsMissingMigrationSnapshot(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	InstallManagedSkill(h, a, home)
	os.Remove(filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename))
	if _, e := InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("lost original migration backup replaced with installed state")
	}
}

func TestManagedSkillWriteFailureRollsBackAndRetries(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	os.WriteFile(path, []byte("before"), 0600)
	saved := managedSkillWrite
	t.Cleanup(func() { managedSkillWrite = saved })
	managedSkillWrite = func(path string, b []byte, mode os.FileMode) (filemerge.WriteResult, error) {
		if strings.HasSuffix(path, "artifacts.md") {
			return filemerge.WriteResult{}, errors.New("injected skill write failure")
		}
		return saved(path, b, mode)
	}
	if _, e := InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("failure ignored")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "before" {
		t.Fatalf("rollback did not restore %s", raw)
	}
	managedSkillWrite = saved
	if _, e := InstallManagedSkill(h, a, home); e != nil {
		t.Fatalf("retry blocked: %v", e)
	}
}

func TestManagedSkillFinalManifestFailurePreservesBackup(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	InstallManagedSkill(h, a, home)
	firstPath := filepath.Join(migrationRoot(a, home, h.ID), backup.ManifestFilename)
	prior, _ := os.ReadFile(firstPath)
	saved := managedSkillWrite
	t.Cleanup(func() { managedSkillWrite = saved })
	managedSkillWrite = func(path string, b []byte, mode os.FileMode) (filemerge.WriteResult, error) {
		if path == firstPath {
			return filemerge.WriteResult{}, errors.New("manifest failure")
		}
		return saved(path, b, mode)
	}
	h.Toggles = []string{"tdd"}
	if _, e := InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("metadata failure ignored")
	}
	after, _ := os.ReadFile(firstPath)
	if !bytes.Equal(prior, after) {
		t.Fatal("original migration manifest corrupted")
	}
	managedSkillWrite = saved
	if e := VerifyManagedSkill(model.Harness{ID: h.ID, Type: h.Type}, a, home); e != nil {
		t.Fatalf("prior package not recovered: %v", e)
	}
}

func TestManagedSkillRejectsMalformedOwnMarkers(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	path := a.InstructionsPath(home)
	os.MkdirAll(filepath.Dir(path), 0700)
	original := []byte("own instructions\n<!-- active-stack:sdd-orchestrator -->\nunterminated")
	os.WriteFile(path, original, 0600)
	if _, e := InstallManagedSkill(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, a, home); e == nil {
		t.Fatal("invalid markers accepted")
	}
	raw, _ := os.ReadFile(path)
	if !bytes.Equal(raw, original) {
		t.Fatal("invalid instructions changed")
	}
}

func TestManagedOperationSnapshotSymlinkIsRejected(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	parent := filepath.Join(home, ".active-stack", "backups")
	os.MkdirAll(parent, 0700)
	target := filepath.Join(home, "foreign")
	os.Mkdir(target, 0700)
	path := filepath.Join(parent, "sdd-orchestrator-codex-native-operation")
	if e := os.Symlink(target, path); e != nil {
		t.Fatal(e)
	}
	if _, e := InstallManagedSkill(h, a, home); e == nil {
		t.Fatal("operation snapshot followed user symlink")
	}
	if _, e := os.Stat(filepath.Join(target, backup.ManifestFilename)); !os.IsNotExist(e) {
		t.Fatal("foreign directory was written")
	}
}

func TestManagedInstallRemovePreserveParentPermissions(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	parent := filepath.Dir(a.InstructionsPath(home))
	os.Mkdir(parent, 0700)
	os.WriteFile(a.InstructionsPath(home), []byte("user rules"), 0600)
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, e := InstallManagedSkill(h, a, home); e != nil {
		t.Fatal(e)
	}
	info, _ := os.Stat(parent)
	if info.Mode().Perm() != 0700 {
		t.Fatalf("install widened parent to %o", info.Mode().Perm())
	}
	if _, e := RemoveManagedSkill(a, home, h.ID); e != nil {
		t.Fatal(e)
	}
	info, _ = os.Stat(parent)
	if info.Mode().Perm() != 0700 {
		t.Fatalf("remove widened parent to %o", info.Mode().Perm())
	}
}
