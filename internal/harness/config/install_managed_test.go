package config_test

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
)

// managedAdapter is a Claude-like adapter with the layered (managed skill) delivery.
type managedAdapter struct{ md, skills, settings string }

func (a managedAdapter) Agent() model.Agent             { return model.AgentClaude }
func (a managedAdapter) InstructionsPath(string) string { return a.md }
func (a managedAdapter) SkillsDir(string) string        { return a.skills }
func (a managedAdapter) VariantKey() string             { return "claude" }
func (a managedAdapter) SettingsPath(string) string     { return a.settings }
func (a managedAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryManagedSkill
}

type managedEnv struct {
	root, home, md, skills, settings string
	adapter                          managedAdapter
}

func newManagedEnv(t *testing.T) managedEnv {
	t.Helper()
	root := t.TempDir()
	e := managedEnv{root: root, home: filepath.Join(root, "home")}
	e.md = filepath.Join(e.home, ".claude", "CLAUDE.md")
	e.skills = filepath.Join(e.home, ".claude", "skills")
	e.settings = filepath.Join(e.home, ".claude", "settings.json")
	e.adapter = managedAdapter{md: e.md, skills: e.skills, settings: e.settings}
	if err := os.MkdirAll(filepath.Dir(e.md), 0o755); err != nil {
		t.Fatal(err)
	}
	return e
}

var orchHarness = model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: allToggles}

func version(state claude.VersionState, v string) func() claude.VersionCheck {
	return func() claude.VersionCheck {
		return claude.VersionCheck{State: state, Version: v, Min: claude.MinVersion}
	}
}

func (e managedEnv) opts(detect func() claude.VersionCheck) config.InstallOptions {
	return config.InstallOptions{Home: e.home, DetectVersion: detect}
}

func (e managedEnv) install(t *testing.T, o config.InstallOptions) config.Result {
	t.Helper()
	res, err := config.InstallWith(orchHarness, []config.AgentAdapter{e.adapter}, e.home, o)
	if err != nil {
		t.Fatalf("InstallWith: %v", err)
	}
	return res
}

func (e managedEnv) write(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const userBefore = "# My rules\n\nbefore text\n"
const userAfter = "\nafter text\n"

func legacyBlock(t *testing.T) string {
	t.Helper()
	composed, err := config.Compose(allToggles, "claude")
	if err != nil {
		t.Fatal(err)
	}
	return filemerge.InjectMarkdownSection("", "sdd-orchestrator", composed)
}

// 6.1 (M): migration of a legacy block keeps the user text byte for byte.
func TestInstallManaged_MigratesLegacyBlock(t *testing.T) {
	e := newManagedEnv(t)
	original := userBefore + "\n" + legacyBlock(t) + userAfter
	e.write(t, e.md, original)

	res := e.install(t, e.opts(version(claude.VersionCompatible, "2.1.318")))

	got := readText(t, e.md)
	if strings.Contains(got, "active-stack:sdd-orchestrator") || strings.Contains(got, "active-stack:sdd-delegation") {
		t.Errorf("coordinator sections must be gone:\n%s", got)
	}
	if strings.Count(got, "<!-- active-stack:"+config.UniversalSectionID+" -->") != 1 {
		t.Errorf("universal block must appear once:\n%s", got)
	}
	// The expected file is the original with the legacy block removed and the
	// universal block injected: everything else is identical, byte for byte.
	layers, err := config.ComposeLayers(allToggles)
	if err != nil {
		t.Fatal(err)
	}
	want := filemerge.InjectMarkdownSection(filemerge.InjectMarkdownSection(original, "sdd-orchestrator", ""), config.UniversalSectionID, layers.Universal)
	if got != want {
		t.Errorf("CLAUDE.md is not byte for byte the expected migration.\nwant %q\ngot  %q", want, got)
	}
	if !strings.HasPrefix(got, userBefore) || !strings.Contains(got, userAfter) {
		t.Errorf("user text before/after changed:\n%q", got)
	}
	md := readText(t, filepath.Join(e.skills, "sdd-orchestrator", "SKILL.md"))
	if !strings.Contains(md, "disable-model-invocation: true") {
		t.Error("SKILL.md missing")
	}
	for _, m := range []string{"artifacts", "tdd", "delegation", "model-assignments", "engram-opsx"} {
		readText(t, filepath.Join(e.skills, "sdd-orchestrator", "modules", m+".md"))
	}
	if res.Modes[model.AgentClaude] != "skill" {
		t.Errorf("mode = %q, want skill", res.Modes[model.AgentClaude])
	}
	// Snapshot with the exact previous content.
	if prev := snapshotContent(t, e.home, e.md); prev != original {
		t.Errorf("snapshot does not hold the previous CLAUDE.md:\n%q", prev)
	}
}

func readText(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Errorf("expected %s: %v", p, err)
	}
	return string(b)
}

// snapshotContent finds the migration snapshot under <home>/.active-stack/backups
// and returns the stored content of path.
func snapshotContent(t *testing.T, home, path string) string {
	t.Helper()
	root := filepath.Join(home, ".active-stack", "backups")
	dirs, _ := os.ReadDir(root)
	for _, d := range dirs {
		if !strings.HasSuffix(d.Name(), "-layered") {
			continue
		}
		tmp := t.TempDir()
		if _, err := backup.ExtractArchive(filepath.Join(root, d.Name(), backup.ArchiveFilename), tmp); err != nil {
			continue
		}
		// Same stripping as backup/snapshot.go: drop the volume ("C:") and the leading separator.
		rel := strings.TrimPrefix(path, filepath.VolumeName(path))
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if b, err := os.ReadFile(filepath.Join(tmp, "files", rel)); err == nil {
			return string(b)
		}
	}
	t.Errorf("no layered snapshot with %s under %s", path, root)
	return ""
}

// 6.2: design-mockups survives the migration untouched.
func TestInstallManaged_KeepsDesignMockupsSection(t *testing.T) {
	e := newManagedEnv(t)
	mockups := filemerge.InjectMarkdownSection("", "design-mockups", "MOCKUPS-BODY\n")
	e.write(t, e.md, userBefore+"\n"+mockups+"\n"+legacyBlock(t))
	e.install(t, e.opts(version(claude.VersionCompatible, "2.1.318")))
	got := readText(t, e.md)
	if !strings.Contains(got, mockups) {
		t.Errorf("design-mockups section changed or vanished:\n%s", got)
	}
}

// 6.3: install never writes the agent key of settings.json and keeps a previous value.
func TestInstallManaged_DoesNotTouchSettingsAgentKey(t *testing.T) {
	e := newManagedEnv(t)
	e.install(t, e.opts(version(claude.VersionCompatible, "2.1.318")))
	if _, err := os.Stat(e.settings); err == nil {
		t.Error("settings.json must not be created by the layered install")
	}
	const prev = "{\n  \"agent\": \"my-own-agent\"\n}\n"
	e.write(t, e.settings, prev)
	e.install(t, e.opts(version(claude.VersionCompatible, "2.1.318")))
	if got := readText(t, e.settings); got != prev {
		t.Errorf("settings.json changed: %q", got)
	}
}

// 6.4a: insufficient version with a previous legacy block: no skill, legacy block kept, warning.
func TestInstallManaged_InsufficientVersionKeepsLegacy(t *testing.T) {
	e := newManagedEnv(t)
	original := userBefore + "\n" + legacyBlock(t)
	e.write(t, e.md, original)
	res := e.install(t, e.opts(version(claude.VersionInsufficient, "2.0.0")))
	if _, err := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); err == nil {
		t.Error("skill must not be written in legacy mode")
	}
	if got := readText(t, e.md); got != original {
		t.Errorf("legacy block must be left as is:\n%q", got)
	}
	if res.Modes[model.AgentClaude] != "legacy" || !hasWarning(res, claude.MinVersion) {
		t.Errorf("want legacy mode and a warning with the minimum version, got %q %v", res.Modes, res.Warnings)
	}
}

func hasWarning(res config.Result, parts ...string) bool {
	for _, w := range res.Warnings {
		ok := true
		for _, p := range parts {
			ok = ok && strings.Contains(w, p)
		}
		if ok {
			return true
		}
	}
	return false
}

// 6.4b: insufficient or unknown version without a previous delivery: the legacy block is written.
func TestInstallManaged_InsufficientOrUnknownWritesLegacyBlock(t *testing.T) {
	for name, state := range map[string]claude.VersionState{"insufficient": claude.VersionInsufficient, "unknown": claude.VersionUnknown} {
		t.Run(name, func(t *testing.T) {
			e := newManagedEnv(t)
			e.write(t, e.md, userBefore)
			res := e.install(t, e.opts(version(state, "")))
			composed, _ := config.Compose(allToggles, "claude")
			want := filemerge.InjectMarkdownSection(userBefore, "sdd-orchestrator", composed)
			if got := readText(t, e.md); got != want {
				t.Errorf("want the legacy composition injected:\n%.300q", got)
			}
			if _, err := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); err == nil {
				t.Error("no skill in legacy mode")
			}
			if res.Modes[model.AgentClaude] != "legacy" || !hasWarning(res, "CLAUDE.md") {
				t.Errorf("want legacy mode and a compatibility warning, got %q %v", res.Modes, res.Warnings)
			}
		})
	}
}

// 6.4c (G11): without the claude binary the layered delivery is written with a warning.
func TestInstallManaged_AbsentBinaryDeliversLayersWithWarning(t *testing.T) {
	e := newManagedEnv(t)
	res := e.install(t, e.opts(version(claude.VersionAbsent, "")))
	readText(t, filepath.Join(e.skills, "sdd-orchestrator", "SKILL.md"))
	if res.Modes[model.AgentClaude] != "skill" || !hasWarning(res, "not verified") {
		t.Errorf("want skill mode and a 'not verified' warning, got %q %v", res.Modes, res.Warnings)
	}
}

// 6.4d: update in legacy mode refreshes the legacy block and is idempotent.
func TestInstallManaged_LegacyUpdateRefreshesAndIsIdempotent(t *testing.T) {
	e := newManagedEnv(t)
	e.write(t, e.md, userBefore)
	e.install(t, e.opts(version(claude.VersionInsufficient, "2.0.0")))
	h2 := orchHarness
	h2.Toggles = []string{"governance"} // "new assets": different composition
	run := func() config.Result {
		res, err := config.InstallWith(h2, []config.AgentAdapter{e.adapter}, e.home, e.opts(version(claude.VersionInsufficient, "2.0.0")))
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	first := run()
	composed, _ := config.Compose(h2.Toggles, "claude")
	if got, want := readText(t, e.md), filemerge.InjectMarkdownSection(userBefore, "sdd-orchestrator", composed); got != want {
		t.Errorf("legacy block not refreshed:\n%.200q", got)
	}
	if first.AllAlready {
		t.Error("first refresh must change the file")
	}
	before := readText(t, e.md)
	if second := run(); !second.AllAlready || readText(t, e.md) != before {
		t.Error("second legacy update must change nothing")
	}
	if _, err := os.Stat(filepath.Join(e.skills, "sdd-orchestrator")); err == nil {
		t.Error("no skill in legacy mode")
	}
}

func witnessLayers(tag string, withTDD bool) func([]string) (config.Layers, error) {
	return func([]string) (config.Layers, error) {
		mods := map[string]string{"artifacts": tag + "-ARTIFACTS\n"}
		if withTDD {
			mods["tdd"] = tag + "-TDD\n"
		}
		return config.Layers{
			Universal: tag + "-UNIVERSAL\n",
			Skill:     config.SkillSpec{ID: "sdd-orchestrator", Description: "d", Body: tag + "-SKILL\n", Modules: mods},
			Modules:   mods,
		}, nil
	}
}

// 6.5: witness T1 -> T2 leaves only T2 everywhere and the user content identical;
// a module whose toggle went away disappears.
func TestInstallManaged_UpdateReplacesWitnessesAndDropsModules(t *testing.T) {
	e := newManagedEnv(t)
	e.write(t, e.md, userBefore+"\n"+legacyBlock(t)+userAfter)
	o := e.opts(version(claude.VersionCompatible, "2.1.318"))
	o.Layers = witnessLayers("T1", true)
	e.install(t, o)
	o.Layers = witnessLayers("T2", false)
	e.install(t, o)

	md := readText(t, e.md)
	if strings.Contains(md, "T1-") || !strings.Contains(md, "T2-UNIVERSAL") {
		t.Errorf("CLAUDE.md must hold only T2:\n%s", md)
	}
	if !strings.HasPrefix(md, userBefore) || !strings.Contains(md, userAfter) {
		t.Errorf("user text changed:\n%q", md)
	}
	sk := filepath.Join(e.skills, "sdd-orchestrator")
	if s := readText(t, filepath.Join(sk, "SKILL.md")); strings.Contains(s, "T1-") || !strings.Contains(s, "T2-SKILL") {
		t.Errorf("SKILL.md must hold only T2:\n%s", s)
	}
	if m := readText(t, filepath.Join(sk, "modules", "artifacts.md")); m != "T2-ARTIFACTS\n" {
		t.Errorf("module not updated: %q", m)
	}
	if _, err := os.Stat(filepath.Join(sk, "modules", "tdd.md")); err == nil {
		t.Error("modules/tdd.md must be removed when the toggle is off")
	}
}

// 6.8: the first migration registers the manifest path and hashes, pins the
// manifest, and a later reinstall does not replace that record.
func TestInstallManaged_RecordsMigrationOnce(t *testing.T) {
	e := newManagedEnv(t)
	e.write(t, e.md, userBefore+"\n"+legacyBlock(t))
	o := e.opts(version(claude.VersionCompatible, "2.1.318"))
	o.Layers = witnessLayers("T1", true)
	res := e.install(t, o)

	rec, ok, err := config.LoadMigrationRecord(e.home, e.md)
	if err != nil || !ok {
		t.Fatalf("migration record missing: ok=%v err=%v", ok, err)
	}
	if rec.ManifestDir == "" || !strings.HasPrefix(rec.ManifestDir, filepath.Join(e.home, ".active-stack", "backups")) {
		t.Errorf("manifest dir = %q", rec.ManifestDir)
	}
	m, err := backup.ReadManifest(filepath.Join(rec.ManifestDir, backup.ManifestFilename))
	if err != nil || !m.Pinned {
		t.Errorf("manifest must be pinned: %+v %v", m.Pinned, err)
	}
	if rec.InstructionsSHA256 != config.SHA256Hex([]byte(readText(t, e.md))) {
		t.Error("recorded CLAUDE.md hash does not match the migrated file")
	}
	skillMD := filepath.Join(e.skills, "sdd-orchestrator", "SKILL.md")
	if rec.SkillFiles[skillMD] != config.SHA256Hex([]byte(readText(t, skillMD))) || len(rec.SkillFiles) != 3 {
		t.Errorf("skill file hashes not recorded: %v", rec.SkillFiles)
	}
	_ = res
	// The legacy block comes back (e.g. restored by hand): migrating again must
	// not replace the first record.
	e.write(t, e.md, readText(t, e.md)+"\n"+legacyBlock(t))
	o.Layers = witnessLayers("T2", false)
	e.install(t, o)
	rec2, _, _ := config.LoadMigrationRecord(e.home, e.md)
	if rec2.ManifestDir != rec.ManifestDir || rec2.InstructionsSHA256 != rec.InstructionsSHA256 {
		t.Errorf("a reinstall replaced the migration record: %+v -> %+v", rec, rec2)
	}
}

// 6.10: an edited skill file is replaced; the edited version stays in a snapshot named in the output.
func TestInstallManaged_EditedSkillFileIsSnapshotted(t *testing.T) {
	e := newManagedEnv(t)
	o := e.opts(version(claude.VersionCompatible, "2.1.318"))
	o.Layers = witnessLayers("T1", true)
	e.install(t, o)
	tdd := filepath.Join(e.skills, "sdd-orchestrator", "modules", "tdd.md")
	e.write(t, tdd, "MY EDITED TDD\n")
	o.Layers = witnessLayers("T2", true)
	res := e.install(t, o)

	if got := readText(t, tdd); got != "T2-TDD\n" {
		t.Errorf("edited file must be replaced, got %q", got)
	}
	if prev := snapshotContent(t, e.home, tdd); prev != "MY EDITED TDD\n" {
		t.Errorf("snapshot must keep the edited version, got %q", prev)
	}
	if !hasWarning(res, filepath.Join(e.home, ".active-stack", "backups")) {
		t.Errorf("output must name the snapshot path: %v", res.Warnings)
	}
}

// 4.6 and 5.3 reach the installer output (Result.Warnings).
func TestInstallManaged_WarningsReachTheResult(t *testing.T) {
	e := newManagedEnv(t)
	e.write(t, e.md, "# Mine\n<!-- active-stack:half-written -->\nno close\n")
	machine := filepath.Join(e.root, "machine", ".claude", "skills")
	e.write(t, filepath.Join(machine, "sdd-orchestrator", "SKILL.md"), "personal")
	o := e.opts(version(claude.VersionCompatible, "2.1.318"))
	o.ShadowSkillsDir = map[model.Agent]string{model.AgentClaude: machine}
	o.Target = model.Project
	res := e.install(t, o)
	if !hasWarning(res, e.md, "half-written") {
		t.Errorf("unclosed marker warning missing: %v", res.Warnings)
	}
	if !hasWarning(res, "precedence", machine, e.skills) {
		t.Errorf("precedence warning missing: %v", res.Warnings)
	}
}

// 6.6: in a Project target nothing is written in <home> except the manifest/backups.
func TestInstallManaged_ProjectTargetWritesOnlyBackupsInHome(t *testing.T) {
	root := t.TempDir()
	home, project := filepath.Join(root, "home"), filepath.Join(root, "project")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	a := managedAdapter{md: filepath.Join(project, ".claude", "CLAUDE.md"), skills: filepath.Join(project, ".claude", "skills")}
	res, err := config.InstallWith(orchHarness, []config.AgentAdapter{a}, project, config.InstallOptions{
		Home: home, Target: model.Project, DetectVersion: version(claude.VersionCompatible, "2.1.318"),
	})
	if err != nil {
		t.Fatal(err)
	}
	readText(t, filepath.Join(project, ".claude", "skills", "sdd-orchestrator", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(project, ".active-stack")); err == nil {
		t.Error("project root must not receive .active-stack")
	}
	var outside []string
	_ = filepath.WalkDir(home, func(p string, d os.DirEntry, err error) error {
		rel, _ := filepath.Rel(home, p)
		if err == nil && !d.IsDir() && !strings.HasPrefix(rel, filepath.Join(".active-stack")) {
			outside = append(outside, rel)
		}
		return nil
	})
	if len(outside) != 0 {
		t.Errorf("home received files outside .active-stack: %v", outside)
	}
	_ = res
}
