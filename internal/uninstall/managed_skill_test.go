package uninstall_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

func TestManagedSkillUninstallPreservesUserNotes(t *testing.T) {
	home := t.TempDir()
	a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := config.Install(h, []config.AgentAdapter{a}, home); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(a.SkillsDir(home), h.ID)
	note := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(note, []byte("my note"), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := uninstall.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, uninstall.Intent{
		Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeCustom, Custom: []string{h.ID}, Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(home, &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentCodex: a}}))
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("managed skill still exists: %v", err)
	}
	if raw, err := os.ReadFile(note); err != nil || string(raw) != "my note" {
		t.Fatalf("user note = %q, error %v", raw, err)
	}
	raw, err := os.ReadFile(a.InstructionsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "<!-- active-stack:sdd-orchestrator -->") {
		t.Fatal("managed instruction marker remains")
	}
}

func TestManagedSkillRollbackPreservesRecreatedFile(t *testing.T) {
	home := t.TempDir()
	a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	plan, err := uninstall.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, uninstall.Intent{
		Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeCustom, Custom: []string{h.ID}, Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(home, &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentCodex: a}}))
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("new user content"), 0o600); err != nil {
		t.Fatal(err)
	}
	rollback := plan.Apply[0].(interface{ Rollback() error })
	if err := rollback.Rollback(); err == nil {
		t.Error("expected rollback conflict")
	}
	if raw, err := os.ReadFile(path); err != nil || string(raw) != "new user content" {
		t.Fatalf("recreated file = %q, error %v", raw, err)
	}
}

func TestManagedSkillRollbackRestoresRemovedPackage(t *testing.T) {
	home := t.TempDir()
	a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"tdd"}}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	plan := managedUninstallPlan(t, h, a, home)
	runManagedUninstall(t, plan)
	if err := plan.Apply[0].(interface{ Rollback() error }).Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := config.VerifyManagedSkill(h, a, home); err != nil {
		t.Fatalf("restored package: %v", err)
	}
}

func TestManagedSkillRollbackRejectsReplacedParent(t *testing.T) {
	home := t.TempDir()
	a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	plan := managedUninstallPlan(t, h, a, home)
	runManagedUninstall(t, plan)
	foreign := t.TempDir()
	dir := filepath.Join(a.SkillsDir(home), h.ID)
	if err := os.Symlink(foreign, dir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := plan.Apply[0].(interface{ Rollback() error }).Rollback(); err == nil {
		t.Fatal("expected replaced-parent conflict")
	}
	entries, err := os.ReadDir(foreign)
	if err != nil || len(entries) != 0 {
		t.Fatalf("foreign directory changed: %v, error %v", entries, err)
	}
}

func TestManagedSkillUninstallPreservesEditedManagedFile(t *testing.T) {
	home := t.TempDir()
	a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := config.InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	instructions, err := os.ReadFile(a.InstructionsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")
	if err := os.WriteFile(path, []byte("user changed managed skill"), 0o600); err != nil {
		t.Fatal(err)
	}
	plan := managedUninstallPlan(t, h, a, home)
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	if err := plan.Apply[0].Run(); err == nil {
		t.Fatal("expected edited managed file conflict")
	}
	if raw, err := os.ReadFile(path); err != nil || string(raw) != "user changed managed skill" {
		t.Fatalf("managed user edit = %q, error %v", raw, err)
	}
	if raw, err := os.ReadFile(a.InstructionsPath(home)); err != nil || string(raw) != string(instructions) {
		t.Fatalf("instructions changed: %q, error %v", raw, err)
	}
}

func managedUninstallPlan(t *testing.T, h model.Harness, a fakeAdapter, home string) uninstall.Plan {
	t.Helper()
	plan, err := uninstall.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, uninstall.Intent{
		Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeCustom, Custom: []string{h.ID}, Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(home, &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentCodex: a}}))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func runManagedUninstall(t *testing.T, plan uninstall.Plan) {
	t.Helper()
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestManagedSkillRollbackPreservesInstructionBytes(t *testing.T) {
	for _, userNotes := range []string{"\r\n\r\n# User notes\r\nkeep trailing spaces  \r\n\r\n", "\n\n# Notes\n\n\n"} {
		t.Run(strings.ReplaceAll(userNotes, "\r", "CR"), func(t *testing.T) {
			home := t.TempDir()
			a := fakeAdapter{agent: model.AgentCodex, homeDir: home, delivery: model.ConfigDeliveryManagedSkill, variantKey: "codex"}
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
			if err := os.WriteFile(a.InstructionsPath(home), []byte(userNotes), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.InstallManagedSkill(h, a, home); err != nil {
				t.Fatal(err)
			}
			installed, err := os.ReadFile(a.InstructionsPath(home))
			if err != nil {
				t.Fatal(err)
			}
			plan := managedUninstallPlan(t, h, a, home)
			runManagedUninstall(t, plan)
			if err := plan.Apply[0].(interface{ Rollback() error }).Rollback(); err != nil {
				t.Fatal(err)
			}
			if restored, err := os.ReadFile(a.InstructionsPath(home)); err != nil || string(restored) != string(installed) {
				t.Fatalf("instruction bytes changed during rollback: %q, error %v", restored, err)
			}
		})
	}
}
