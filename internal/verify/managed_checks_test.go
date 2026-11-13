package verify_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

var orchestrator = model.Harness{
	ID: "sdd-orchestrator", Type: model.HarnessConfig,
	Toggles: []string{"delegation", "model-routing", "engram", "tdd", "governance"},
}

type managedFixture struct {
	home, md, skills string
	adapter          fakeAdapter
}

func compat(state claude.VersionState) func() claude.VersionCheck {
	return func() claude.VersionCheck {
		return claude.VersionCheck{State: state, Version: "2.1.318", Min: claude.MinVersion}
	}
}

// installManaged installs the layered delivery into a temp home (as the installer does).
func installManaged(t *testing.T) managedFixture {
	t.Helper()
	home := t.TempDir()
	f := managedFixture{home: home, md: filepath.Join(home, ".claude", "CLAUDE.md"), skills: filepath.Join(home, ".claude", "skills")}
	f.adapter = fakeAdapter{agent: model.AgentClaude, skillsDir: f.skills, instructionsPath: f.md, delivery: model.ConfigDeliveryManagedSkill}
	if err := os.MkdirAll(filepath.Dir(f.md), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.md, []byte("# mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := managedCfgAdapter{f}
	if _, err := config.InstallWith(orchestrator, []config.AgentAdapter{a}, home, config.InstallOptions{Home: home, DetectVersion: compat(claude.VersionCompatible)}); err != nil {
		t.Fatal(err)
	}
	return f
}

type managedCfgAdapter struct{ f managedFixture }

func (a managedCfgAdapter) Agent() model.Agent             { return model.AgentClaude }
func (a managedCfgAdapter) InstructionsPath(string) string { return a.f.md }
func (a managedCfgAdapter) SkillsDir(string) string        { return a.f.skills }
func (a managedCfgAdapter) VariantKey() string             { return "claude" }
func (a managedCfgAdapter) SettingsPath(string) string     { return "" }
func (a managedCfgAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryManagedSkill
}

func (f managedFixture) verify(t *testing.T, state claude.VersionState) []verify.CheckResult {
	t.Helper()
	old := verify.DetectClaudeVersion
	verify.DetectClaudeVersion = compat(state)
	defer func() { verify.DetectClaudeVersion = old }()
	return verify.RunChecks(context.Background(), verify.ChecksForHarness(orchestrator, []verify.Adapter{f.adapter}, f.home))
}

func statuses(rs []verify.CheckResult) (failed, warned []string) {
	for _, r := range rs {
		switch r.Status {
		case verify.CheckStatusFailed:
			failed = append(failed, r.ID+": "+r.Error)
		case verify.CheckStatusWarning:
			warned = append(warned, r.ID+": "+r.Error)
		}
	}
	return
}

func (f managedFixture) skillFile(rel string) string {
	return filepath.Join(f.skills, "sdd-orchestrator", filepath.FromSlash(rel))
}

// 7.1: a correct layered install passes and the output names the skill route.
func TestManagedVerify_PassesOnCorrectInstall(t *testing.T) {
	f := installManaged(t)
	rs := f.verify(t, claude.VersionCompatible)
	failed, warned := statuses(rs)
	if len(failed) != 0 || len(warned) != 0 {
		t.Errorf("correct install must pass cleanly: failed=%v warned=%v", failed, warned)
	}
	found := false
	for _, r := range rs {
		found = found || strings.Contains(r.Description, "skill")
	}
	if !found {
		t.Errorf("output must name the active route (skill): %+v", rs)
	}
}

// 7.1: each defect of a compatible install fails.
func TestManagedVerify_FailsOnEachDefect(t *testing.T) {
	cases := map[string]func(t *testing.T, f managedFixture){
		"missing SKILL.md": func(t *testing.T, f managedFixture) { mustRemove(t, f.skillFile("SKILL.md")) },
		"missing module of an active toggle": func(t *testing.T, f managedFixture) {
			mustRemove(t, f.skillFile("modules/tdd.md"))
		},
		"SKILL.md without disable-model-invocation": func(t *testing.T, f managedFixture) {
			b, _ := os.ReadFile(f.skillFile("SKILL.md"))
			mustWrite(t, f.skillFile("SKILL.md"), strings.ReplaceAll(string(b), "disable-model-invocation: true", "disable-model-invocation: false"))
		},
		"universal marker twice": func(t *testing.T, f managedFixture) {
			b, _ := os.ReadFile(f.md)
			mustWrite(t, f.md, string(b)+"\n<!-- active-stack:"+config.UniversalSectionID+" -->\ndup\n<!-- /active-stack:"+config.UniversalSectionID+" -->\n")
		},
		"universal marker absent": func(t *testing.T, f managedFixture) {
			b, _ := os.ReadFile(f.md)
			mustWrite(t, f.md, filemerge.InjectMarkdownSection(string(b), config.UniversalSectionID, ""))
		},
		"legacy orchestrator section left": func(t *testing.T, f managedFixture) {
			b, _ := os.ReadFile(f.md)
			mustWrite(t, f.md, filemerge.InjectMarkdownSection(string(b), "sdd-orchestrator", "legacy"))
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := installManaged(t)
			mutate(t, f)
			if failed, _ := statuses(f.verify(t, claude.VersionCompatible)); len(failed) == 0 {
				t.Errorf("%s must fail verification", name)
			}
		})
	}
}

// 7.1: legacy mode by version and an unchecked version are warnings, never failures.
func TestManagedVerify_LegacyAndUncheckedAreWarnings(t *testing.T) {
	home := t.TempDir()
	f := managedFixture{home: home, md: filepath.Join(home, "CLAUDE.md"), skills: filepath.Join(home, "skills")}
	f.adapter = fakeAdapter{agent: model.AgentClaude, skillsDir: f.skills, instructionsPath: f.md, delivery: model.ConfigDeliveryManagedSkill}
	composed, _ := config.Compose(orchestrator.Toggles, "claude")
	mustWrite(t, f.md, filemerge.InjectMarkdownSection("# mine\n", "sdd-orchestrator", composed))
	for _, state := range []claude.VersionState{claude.VersionInsufficient, claude.VersionUnknown} {
		rs := f.verify(t, state)
		failed, warned := statuses(rs)
		if len(failed) != 0 || len(warned) == 0 || !strings.Contains(strings.Join(warned, "\n"), "legacy mode by version") {
			t.Errorf("state %v: want a 'legacy mode by version' warning and no failure, failed=%v warned=%v", state, failed, warned)
		}
	}
	// Absent binary with a correct layered install: pass plus a "version not verified" warning.
	g := installManaged(t)
	failed, warned := statuses(g.verify(t, claude.VersionAbsent))
	if len(failed) != 0 || !strings.Contains(strings.Join(warned, "\n"), "version not verified") {
		t.Errorf("absent: want warning 'version not verified', failed=%v warned=%v", failed, warned)
	}
}

// 7.4: the check uses the skills dir the adapter resolves; a skill that exists
// only under <home> does not satisfy a Project adapter.
func TestManagedVerify_UsesTheAdapterSkillsDir(t *testing.T) {
	f := installManaged(t) // skill installed under the "home" layout
	project := t.TempDir()
	proj := fakeAdapter{agent: model.AgentClaude, skillsDir: filepath.Join(project, ".claude", "skills"), instructionsPath: f.md, delivery: model.ConfigDeliveryManagedSkill}
	old := verify.DetectClaudeVersion
	verify.DetectClaudeVersion = compat(claude.VersionCompatible)
	defer func() { verify.DetectClaudeVersion = old }()
	rs := verify.RunChecks(context.Background(), verify.ChecksForHarness(orchestrator, []verify.Adapter{proj}, f.home))
	failed, _ := statuses(rs)
	if len(failed) == 0 {
		t.Error("a skill present only in <home> must not make the Project check pass")
	}
	if !strings.Contains(strings.Join(failed, "\n"), filepath.Join(project, ".claude", "skills", "sdd-orchestrator")) {
		t.Errorf("failure must name the project skill path: %v", failed)
	}
}

func mustRemove(t *testing.T, p string) {
	t.Helper()
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, c string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
		t.Fatal(err)
	}
}
