package uninstall_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

// layeredAdapter satisfies both config.AgentAdapter and uninstall.AgentAdapter.
type layeredAdapter struct{ home string }

func (a layeredAdapter) Agent() model.Agent { return model.AgentClaude }
func (a layeredAdapter) InstructionsPath(string) string {
	return filepath.Join(a.home, ".claude", "CLAUDE.md")
}
func (a layeredAdapter) SkillsDir(string) string { return filepath.Join(a.home, ".claude", "skills") }
func (a layeredAdapter) SettingsPath(string) string {
	return filepath.Join(a.home, ".claude", "settings.json")
}
func (a layeredAdapter) CommandsDir(string) string { return "" }
func (a layeredAdapter) VariantKey() string        { return "claude" }
func (a layeredAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryManagedSkill
}

var orch = model.Harness{
	ID: "sdd-orchestrator", Type: model.HarnessConfig,
	Toggles:      []string{"tdd", "delegation"},
	InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
}

type migrated struct {
	home, md, skill, settings, original string
	record                              config.MigrationRecord
	adapter                             layeredAdapter
}

// migrate builds a legacy CLAUDE.md, migrates it with the real installer and
// returns the pieces the uninstall tests need.
func migrate(t *testing.T) migrated {
	t.Helper()
	home := t.TempDir()
	m := migrated{home: home, adapter: layeredAdapter{home}}
	m.md = m.adapter.InstructionsPath(home)
	m.skill = filepath.Join(m.adapter.SkillsDir(home), "sdd-orchestrator")
	m.settings = m.adapter.SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(m.md), 0o755); err != nil {
		t.Fatal(err)
	}
	composed, _ := config.Compose(orch.Toggles, "claude")
	m.original = "# mine\n\nbefore\n\n" + filemerge.InjectMarkdownSection("", "sdd-orchestrator", composed) + "\nafter\n"
	write(t, m.md, m.original)
	write(t, m.settings, "{\"agent\": \"mine\"}\n")
	_, err := config.InstallWith(orch, []config.AgentAdapter{m.adapter}, home, config.InstallOptions{
		Home: home,
		DetectVersion: func() claude.VersionCheck {
			return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec, ok, err := config.LoadMigrationRecord(home, m.md)
	if err != nil || !ok {
		t.Fatalf("migration record: ok=%v err=%v", ok, err)
	}
	m.record = rec
	return m
}

func write(t *testing.T, p, c string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func plan(t *testing.T, m migrated, intent uninstall.Intent, opts uninstall.Options) uninstall.Plan {
	t.Helper()
	opts.HomeDir = m.home
	opts.Registry = &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: m.adapter}}
	intent.Agents = []model.Agent{model.AgentClaude}
	intent.Mode = model.ModeLite
	p, err := uninstall.BuildPlan(&fakeCatalog{harnesses: []model.Harness{orch}}, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return p
}

func runApply(p uninstall.Plan) error {
	for _, s := range p.Prepare {
		if err := s.Run(); err != nil {
			return err
		}
	}
	for _, s := range p.Apply {
		if err := s.Run(); err != nil {
			return err
		}
	}
	return nil
}

// 7.2: targeted uninstall removes only the universal block and the owned skill dir.
func TestManagedUninstall_TargetedRemovesOnlyOwnedPieces(t *testing.T) {
	m := migrate(t)
	other := filepath.Join(m.adapter.SkillsDir(m.home), "other-skill", "SKILL.md")
	write(t, other, "mine")
	if err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyTargeted}, uninstall.Options{})); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	got := read(t, m.md)
	if strings.Contains(got, "active-stack:") || !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Errorf("only the universal block must go; user text stays:\n%s", got)
	}
	if exists(m.skill) {
		t.Error("the owned skill dir must be removed")
	}
	if read(t, other) != "mine" || read(t, m.settings) != "{\"agent\": \"mine\"}\n" {
		t.Error("other skills and settings.json must stay untouched")
	}
}

// 7.2 (M): a skill directory without the ownership mark is never deleted.
func TestManagedUninstall_KeepsSkillWithoutMark(t *testing.T) {
	m := migrate(t)
	write(t, filepath.Join(m.skill, "SKILL.md"), "---\nname: sdd-orchestrator\n---\nMY OWN\n")
	if err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyTargeted}, uninstall.Options{})); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if read(t, filepath.Join(m.skill, "SKILL.md")) != "---\nname: sdd-orchestrator\n---\nMY OWN\n" {
		t.Error("a skill without the ownership mark must be kept")
	}
}

// 7.6: if deleting the skill dir fails after the universal block was removed,
// the block is restored and the error names the path.
func TestManagedUninstall_SkillRemovalFailureRestoresBlock(t *testing.T) {
	m := migrate(t)
	before := read(t, m.md)
	defer uninstall.SetSkillRemovalFn(func(string) error { return os.ErrPermission })()
	err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyTargeted}, uninstall.Options{}))
	if err == nil || !strings.Contains(err.Error(), m.skill) {
		t.Fatalf("want an error naming %q, got %v", m.skill, err)
	}
	if got := read(t, m.md); got != before {
		t.Errorf("CLAUDE.md must be restored byte for byte.\nwant %q\ngot  %q", before, got)
	}
}

func migrationManifest(t *testing.T, m migrated) backup.Manifest {
	t.Helper()
	mf, err := backup.ReadManifest(filepath.Join(m.record.ManifestDir, backup.ManifestFilename))
	if err != nil {
		t.Fatal(err)
	}
	return mf
}

// 7.2: restore without later edits returns CLAUDE.md byte for byte, removes the
// skill dir and does not touch settings.json.
func TestManagedUninstall_RestoreWithoutEdits(t *testing.T) {
	m := migrate(t)
	mf := migrationManifest(t, m)
	if err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &mf})); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := read(t, m.md); got != m.original {
		t.Errorf("CLAUDE.md not restored byte for byte.\nwant %q\ngot  %q", m.original, got)
	}
	if exists(m.skill) {
		t.Error("the skill dir created by Active Stack must be gone")
	}
	if read(t, m.settings) != "{\"agent\": \"mine\"}\n" {
		t.Error("settings.json must not change")
	}
}

func editedSnapshots(t *testing.T, home string) []string {
	t.Helper()
	root := filepath.Join(home, ".active-stack", "backups")
	var out []string
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if strings.Contains(e.Name(), "uninstall-edited") {
			out = append(out, filepath.Join(root, e.Name()))
		}
	}
	return out
}

// 7.2 (M): an edited CLAUDE.md is not overwritten without --overwrite-edited
// (and --yes does not change that: the guard is independent of it); the edited
// content is saved first; with the flag the restore proceeds.
func TestManagedUninstall_RestoreGuardsEditedFiles(t *testing.T) {
	m := migrate(t)
	mf := migrationManifest(t, m)
	edited := read(t, m.md) + "\nmy later edit\n"
	write(t, m.md, edited)

	err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &mf}))
	if err == nil || !strings.Contains(err.Error(), m.md) || !strings.Contains(err.Error(), "--overwrite-edited") {
		t.Fatalf("want an error naming %q and --overwrite-edited, got %v", m.md, err)
	}
	if read(t, m.md) != edited {
		t.Error("CLAUDE.md must not be written when it was edited")
	}
	snaps := editedSnapshots(t, m.home)
	if len(snaps) != 1 {
		t.Fatalf("want one snapshot of the edited content, got %v", snaps)
	}
	if !exists(m.skill) {
		t.Error("nothing may be restored when the guard trips")
	}

	err = runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &mf, OverwriteEdited: true}))
	if err != nil {
		t.Fatalf("restore with --overwrite-edited: %v", err)
	}
	if read(t, m.md) != m.original {
		t.Error("with --overwrite-edited the original must be restored")
	}
	if len(editedSnapshots(t, m.home)) < 2 {
		t.Error("the edited content must be saved again before overwriting")
	}
}

// 7.2: a manifest other than the registered migration one keeps the guard and warns.
func TestManagedUninstall_OtherManifestKeepsGuardAndWarns(t *testing.T) {
	m := migrate(t)
	other, err := backup.NewSnapshotter().Create(filepath.Join(m.home, "other-manifest"), []string{m.md})
	if err != nil {
		t.Fatal(err)
	}
	edited := read(t, m.md) + "\nedit\n"
	write(t, m.md, edited)
	var warned []string
	opts := uninstall.Options{RestoreManifest: &other, OnProgress: func(e pipeline.ProgressEvent) {
		if e.Err != nil {
			warned = append(warned, e.Err.Error())
		}
	}}
	err = runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, opts))
	if err == nil {
		t.Fatal("the hash guard must apply to any manifest")
	}
	if read(t, m.md) != edited {
		t.Error("CLAUDE.md must not be overwritten")
	}
	joined := strings.Join(warned, "\n")
	if !strings.Contains(joined, other.RootDir) || !strings.Contains(joined, "migration manifest") {
		t.Errorf("output must warn that it is not the migration manifest and name its path: %q", joined)
	}
}

// B2 (PR #8 review): skipping settings.json only applies when the manifest being
// restored is the registered claude-native-opsx migration one. Any other
// manifest (for instance the permissions harness one) restores it as in main.
func TestManagedUninstall_RestoreOtherManifestRestoresSettings(t *testing.T) {
	m := migrate(t)
	const before = "{\"permissions\": \"BEFORE\"}\n"
	write(t, m.settings, before)
	other, err := backup.NewSnapshotter().Create(filepath.Join(m.home, "settings-manifest"), []string{m.settings})
	if err != nil {
		t.Fatal(err)
	}
	write(t, m.settings, "{\"permissions\": \"AFTER-INSTALL\"}\n")
	if err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &other})); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := read(t, m.settings); got != before {
		t.Errorf("settings.json must be restored from a non-migration manifest.\nwant %q\ngot  %q", before, got)
	}
}

// The migration manifest itself never rewrites settings.json, even when it holds
// a snapshot of it.
func TestManagedUninstall_RestoreMigrationManifestSkipsSettings(t *testing.T) {
	m := migrate(t)
	mf := migrationManifest(t, m)
	snap, err := backup.NewSnapshotter().Create(filepath.Join(m.home, "inject-settings"), []string{m.settings})
	if err != nil {
		t.Fatal(err)
	}
	mf.Entries = append(mf.Entries, snap.Entries...)
	write(t, m.settings, "{\"agent\": \"later\"}\n")
	if err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &mf})); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if read(t, m.settings) != "{\"agent\": \"later\"}\n" {
		t.Error("the migration manifest must not touch settings.json")
	}
}

// C2 (PR #8 review): the guard fails closed. A claude-migration.json that exists
// but cannot be parsed stops the restore before writing anything, and the error
// names the file.
func TestManagedUninstall_RestoreFailsClosedOnUnreadableMigrationRecord(t *testing.T) {
	m := migrate(t)
	mf := migrationManifest(t, m)
	edited := read(t, m.md) + "\nmy later edit\n"
	write(t, m.md, edited)
	recPath := config.MigrationRecordPath(m.home)
	write(t, recPath, "{not json")

	err := runApply(plan(t, m, uninstall.Intent{Strategy: uninstall.StrategyRestore}, uninstall.Options{RestoreManifest: &mf, OverwriteEdited: true}))
	if err == nil || !strings.Contains(err.Error(), recPath) {
		t.Fatalf("want an error naming %q, got %v", recPath, err)
	}
	if read(t, m.md) != edited {
		t.Error("nothing may be written when the migration record cannot be read")
	}
	if !exists(m.skill) {
		t.Error("the skill dir must stay when the restore is refused")
	}
}
