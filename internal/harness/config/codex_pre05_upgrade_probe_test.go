package config_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

const preContextSource = "91fbb70950babe3b694cfa1ddee141fd959acb40"

// This opt-in compatibility probe uses the actual pre-05 installer to produce
// ownership and backup data. It does not fabricate an old inventory format,
// change production seams, build a product binary or invoke a model.
func TestCodexPre05PackageUpgrade(t *testing.T) {
	if os.Getenv("ACTIVE_STACK_CODEX_PRE05_UPGRADE_PROBE") != "1" {
		t.Skip("set ACTIVE_STACK_CODEX_PRE05_UPGRADE_PROBE=1 to exercise the real pre-05 installer")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	rootBytes, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatal(err)
	}
	repo := strings.TrimSpace(string(rootBytes))
	oldTree := filepath.Join(t.TempDir(), "old source")
	gitAdd := exec.CommandContext(ctx, "git", "worktree", "add", "--detach", oldTree, preContextSource)
	gitAdd.Dir = repo
	if out, err := gitAdd.CombinedOutput(); err != nil {
		t.Fatalf("prepare old source: %s: %v", out, err)
	}
	defer func() {
		cleanup := exec.Command("git", "worktree", "remove", "--force", oldTree)
		cleanup.Dir = repo
		if out, err := cleanup.CombinedOutput(); err != nil {
			t.Errorf("remove old test worktree: %s: %v", out, err)
		}
	}()
	if err := os.WriteFile(filepath.Join(oldTree, "internal", "harness", "config", "pre05_fixture_export_test.go"), []byte(preContextFixtureTest), 0600); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(t.TempDir(), "pre-05 profile with spaces")
	seed := exec.CommandContext(ctx, "go", "test", "./internal/harness/config", "-run", "^TestExportPre05Fixture$", "-count=1", "-v")
	seed.Dir = oldTree
	seed.Env = append(os.Environ(), "ACTIVE_STACK_CODEX_OLD_SEED_HOME="+home)
	if out, err := seed.CombinedOutput(); err != nil {
		t.Fatalf("run old installer at %s: %s: %v", preContextSource, out, err)
	}
	t.Logf("seed produced by actual InstallManagedSkill at %s", preContextSource)

	a := codex.NewAdapter()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation", "engram", "governance", "tdd", "model-routing"}}
	skillRoot := filepath.Join(a.SkillsDir(home), h.ID)
	newModules := []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"}
	for _, id := range newModules {
		if _, err := os.Stat(filepath.Join(skillRoot, "modules", id+".md")); !os.IsNotExist(err) {
			t.Fatalf("seed unexpectedly contains new module %s: %v", id, err)
		}
	}
	migration := filepath.Join(home, ".active-stack", "backups", h.ID+"-codex-native-migration")
	manifestPath := filepath.Join(migration, backup.ManifestFilename)
	before, err := backup.ReadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	archiveBefore, err := os.ReadFile(filepath.Join(migration, "snapshot.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	note := filepath.Join(skillRoot, "user-note.md")
	if err := os.WriteFile(note, []byte("user note from old installation\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	if err := config.VerifyManagedSkill(h, a, home); err != nil {
		t.Fatalf("upgraded package does not verify: %v", err)
	}
	var state struct {
		Owner string            `json:"owner"`
		Files map[string]string `json:"files"`
	}
	raw, err := os.ReadFile(filepath.Join(skillRoot, ".active-stack-managed.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	if state.Owner != "active-stack" {
		t.Fatalf("upgraded ownership lost: %s", state.Owner)
	}
	for _, id := range newModules {
		relative := filepath.Join("modules", id+".md")
		content, err := os.ReadFile(filepath.Join(skillRoot, relative))
		if err != nil || len(bytes.TrimSpace(content)) == 0 {
			t.Fatalf("upgraded module %s: %v", id, err)
		}
		digest := sha256.Sum256(content)
		if state.Files[relative] != hex.EncodeToString(digest[:]) {
			t.Fatalf("module %s not inventoried with its exact bytes", id)
		}
	}
	// A second update changes optional modules. It must preserve both the new
	// always-delivered context modules and the original migration archive.
	h.Toggles = []string{"engram"}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	result, err := config.InstallManagedSkill(h, a, home)
	if err != nil || !result.AllAlready || len(result.Files) != 0 {
		t.Fatalf("upgrade reinstall is not idempotent: %+v %v", result, err)
	}
	after, err := backup.ReadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	archiveAfter, err := os.ReadFile(filepath.Join(migration, "snapshot.tar.gz"))
	if err != nil || !bytes.Equal(archiveBefore, archiveAfter) || before.ID != after.ID || !before.CreatedAt.Equal(after.CreatedAt) || before.RootDir != after.RootDir || !after.Pinned {
		t.Fatalf("first snapshot identity/archive changed: %v", err)
	}
	if _, err := config.RemoveManagedSkill(a, home, h.ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range newModules {
		if _, err := os.Stat(filepath.Join(skillRoot, "modules", id+".md")); !os.IsNotExist(err) {
			t.Fatalf("new owned module %s not removed: %v", id, err)
		}
	}
	if raw, err := os.ReadFile(note); err != nil || string(raw) != "user note from old installation\n" {
		t.Fatalf("old installation user note lost: %s %v", raw, err)
	}
	if raw, err := os.ReadFile(a.InstructionsPath(home)); err != nil || !bytes.Contains(raw, []byte("pre-05 user rules")) || bytes.Contains(raw, []byte("<!-- active-stack:sdd-orchestrator -->")) {
		t.Fatalf("uninstall did not preserve old user rules: %s %v", raw, err)
	}

	// Use another authentic old profile for later restoration, independently of
	// the uninstall case containing a new user note.
	restoreHome := filepath.Join(t.TempDir(), "pre-05 restoration profile")
	restoreSeed := exec.CommandContext(ctx, "go", "test", "./internal/harness/config", "-run", "^TestExportPre05Fixture$", "-count=1")
	restoreSeed.Dir = oldTree
	restoreSeed.Env = append(os.Environ(), "ACTIVE_STACK_CODEX_OLD_SEED_HOME="+restoreHome)
	if out, err := restoreSeed.CombinedOutput(); err != nil {
		t.Fatalf("seed old restoration profile: %s: %v", out, err)
	}
	if _, err := config.InstallManagedSkill(h, a, restoreHome); err != nil {
		t.Fatal(err)
	}
	restoredManifest, err := backup.ReadManifest(filepath.Join(restoreHome, ".active-stack", "backups", h.ID+"-codex-native-migration", backup.ManifestFilename))
	if err != nil {
		t.Fatal(err)
	}
	if err := (backup.RestoreService{}).Restore(restoredManifest); err != nil {
		t.Fatalf("restore reloaded original snapshot after upgrade: %v", err)
	}
	if raw, err := os.ReadFile(a.InstructionsPath(restoreHome)); err != nil || string(raw) != "pre-05 user rules\n" {
		t.Fatalf("original pre-05 instructions not restored: %s %v", raw, err)
	}
	for _, id := range newModules {
		if _, err := os.Stat(filepath.Join(a.SkillsDir(restoreHome), h.ID, "modules", id+".md")); !os.IsNotExist(err) {
			t.Fatalf("later restoration retained new module %s: %v", id, err)
		}
	}
}

const preContextFixtureTest = `package config_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/Group-Active-IA/active-stack/internal/agents/codex"
    "github.com/Group-Active-IA/active-stack/internal/harness/config"
    "github.com/Group-Active-IA/active-stack/internal/model"
)

func TestExportPre05Fixture(t *testing.T) {
    home := os.Getenv("ACTIVE_STACK_CODEX_OLD_SEED_HOME")
    if home == "" { t.Fatal("missing isolated seed home") }
    a := codex.NewAdapter()
    if err := os.MkdirAll(filepath.Dir(a.InstructionsPath(home)), 0700); err != nil { t.Fatal(err) }
    if err := os.WriteFile(a.InstructionsPath(home), []byte("pre-05 user rules\n"), 0600); err != nil { t.Fatal(err) }
    h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation", "engram", "governance", "tdd", "model-routing"}}
    if _, err := config.InstallManagedSkill(h, a, home); err != nil { t.Fatal(err) }
}
`
