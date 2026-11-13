package install_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	cfginstaller "github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// managedProjectAdapter is projectAdapter with the layered (managed skill) delivery.
type managedProjectAdapter struct{ projectAdapter }

func (managedProjectAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryManagedSkill
}

var orchestratorHarness = model.Harness{
	ID: "sdd-orchestrator", Name: "SDD", Type: model.HarnessConfig,
	Toggles:      []string{"tdd"},
	InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	Agents:       []model.Agent{model.AgentClaude},
}

func managedPlan(t *testing.T, opts install.Options) install.Plan {
	t.Helper()
	cat := &fakeCatalog{harnesses: []model.Harness{orchestratorHarness}}
	opts.Registry = &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: managedProjectAdapter{projectAdapter{agent: model.AgentClaude}},
	}}
	opts.NoSelfInstall = true
	plan, err := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeLite}, opts)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return plan
}

// 6.6: the snapshot plan covers the skill directory (and its files when present)
// of the target, never a path outside it.
func TestManagedWiring_WritePathsIncludeSkillDir(t *testing.T) {
	root := t.TempDir()
	skill := filepath.Join(root, ".claude", "skills", "sdd-orchestrator")
	if err := os.MkdirAll(filepath.Join(skill, "modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "modules", "tdd.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var captured []string
	defer install.SetSnapshotCreate(func(_ string, paths []string) (backup.Manifest, error) {
		captured = paths
		return backup.Manifest{}, nil
	})()
	plan := managedPlan(t, install.Options{HomeDir: t.TempDir(), ProjectRoot: root, Target: model.Project})
	for _, s := range plan.Prepare {
		_ = s.Run()
	}
	has := func(p string) bool {
		for _, c := range captured {
			if filepath.Clean(c) == filepath.Clean(p) {
				return true
			}
		}
		return false
	}
	if !has(skill) || !has(filepath.Join(skill, "modules", "tdd.md")) || !has(filepath.Join(skill, "modules")) {
		t.Errorf("snapshot paths must include the skill dir and its files, got %v", captured)
	}
}

// 6.6 / C8: the installer receives the real home and the target, and the Machine
// skills dir as the shadow directory.
func TestManagedWiring_PassesHomeTargetAndShadowDir(t *testing.T) {
	home, root := t.TempDir(), t.TempDir()
	var got cfginstaller.InstallOptions
	defer install.SetConfigInstallWithFn(func(_ model.Harness, _ string, o cfginstaller.InstallOptions) (cfginstaller.Result, error) {
		got = o
		return cfginstaller.Result{}, nil
	})()
	plan := managedPlan(t, install.Options{HomeDir: home, ProjectRoot: root, Target: model.Project})
	for _, s := range plan.Apply {
		_ = s.Run()
	}
	if got.Home != home || got.Target != model.Project {
		t.Errorf("options = %+v, want Home=%q Target=Project", got, home)
	}
	if want := filepath.Join(home, ".claude", "skills"); got.ShadowSkillsDir[model.AgentClaude] != want {
		t.Errorf("shadow dir = %q, want %q", got.ShadowSkillsDir[model.AgentClaude], want)
	}
}

func runApplyEvents(t *testing.T, opts install.Options, res cfginstaller.Result) []pipeline.ProgressEvent {
	t.Helper()
	defer install.SetConfigInstallWithFn(func(model.Harness, string, cfginstaller.InstallOptions) (cfginstaller.Result, error) {
		return res, nil
	})()
	var events []pipeline.ProgressEvent
	opts.OnProgress = func(e pipeline.ProgressEvent) { events = append(events, e) }
	plan := managedPlan(t, opts)
	for _, s := range plan.Apply {
		if err := s.Run(); err != nil {
			t.Fatal(err)
		}
	}
	return events
}

func joinedErrs(evs []pipeline.ProgressEvent) string {
	var b strings.Builder
	for _, e := range evs {
		if e.Err != nil {
			b.WriteString(e.Err.Error() + "\n")
		}
	}
	return b.String()
}

// 4.6 / 5.3 / 6.7: warnings and the BREAKING notice reach the installer output,
// in both languages.
func TestManagedWiring_WarningsAndBreakingNoticeReachOutput(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.LangEN, i18n.LangES} {
		t.Run(string(lang), func(t *testing.T) {
			res := cfginstaller.Result{
				Warnings: []string{"WARN-UNCLOSED-MARKER", "WARN-PRECEDENCE"},
				Modes:    map[model.Agent]string{model.AgentClaude: "skill"},
			}
			evs := runApplyEvents(t, install.Options{HomeDir: t.TempDir(), Lang: lang}, res)
			out := joinedErrs(evs)
			for _, want := range []string{"WARN-UNCLOSED-MARKER", "WARN-PRECEDENCE", "/sdd-orchestrator", "BREAKING"} {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q:\n%s", want, out)
				}
			}
			if lang == i18n.LangES && !strings.Contains(out, "ya no vive en CLAUDE.md") {
				t.Errorf("Spanish notice expected:\n%s", out)
			}
		})
	}
}

// Legacy mode: warnings yes, no BREAKING notice (the coordinator did not move).
func TestManagedWiring_LegacyModeHasNoBreakingNotice(t *testing.T) {
	res := cfginstaller.Result{Warnings: []string{"WARN-LEGACY"}, Modes: map[model.Agent]string{model.AgentClaude: "legacy"}}
	out := joinedErrs(runApplyEvents(t, install.Options{HomeDir: t.TempDir()}, res))
	if !strings.Contains(out, "WARN-LEGACY") || strings.Contains(out, "BREAKING") {
		t.Errorf("legacy output wrong:\n%s", out)
	}
}

// 7.5: a Project install names the manual cleanup (uninstall is Machine-only).
func TestManagedWiring_ProjectInstallNamesManualCleanup(t *testing.T) {
	root := t.TempDir()
	res := cfginstaller.Result{Modes: map[model.Agent]string{model.AgentClaude: "skill"}}
	out := joinedErrs(runApplyEvents(t, install.Options{HomeDir: t.TempDir(), ProjectRoot: root, Target: model.Project}, res))
	skill := filepath.Join(root, ".claude", "skills", "sdd-orchestrator")
	if !strings.Contains(out, skill) || !strings.Contains(out, "universal-rules") {
		t.Errorf("Project output must name %q and the universal section as manual cleanup:\n%s", skill, out)
	}
	machine := joinedErrs(runApplyEvents(t, install.Options{HomeDir: t.TempDir()}, res))
	if strings.Contains(machine, "universal-rules") {
		t.Error("Machine install must not print the manual cleanup")
	}
}

// C5 (PR #8 review): the install does not depend on the claude binary of the
// machine. A claude on PATH that reports an old version must not change the
// result when the detector is injected.
func TestManagedWiring_InstallIgnoresClaudeOnPathWhenDetectorInjected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the PATH stub is a POSIX shell script")
	}
	bin := t.TempDir()
	stub := "#!/bin/sh\necho '1.0.0 (Claude Code)'\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	defer install.SetClaudeDetectFn(func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
	})()
	home := t.TempDir()
	plan := managedPlan(t, install.Options{HomeDir: home})
	for _, s := range append(append([]pipeline.Step{}, plan.Prepare...), plan.Apply...) {
		if err := s.Run(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "sdd-orchestrator", "SKILL.md")); err != nil {
		t.Errorf("the injected detector must decide the layered delivery: %v", err)
	}
}
