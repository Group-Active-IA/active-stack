package headless_test

// C-32: headless executor honest reporting — tasks 3.1, 3.2, 3.3, 3.4, 3.5

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	extinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	skillinstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// buildDegradedCatReg builds a catalog with a single best-effort skill harness whose
// install always fails, and a registry wired for claude.
func buildDegradedCatReg() (*fakeExecCatalog, *fakeExecRegistry) {
	h := model.Harness{
		ID:           "best-effort-h",
		Type:         model.HarnessSkill,
		BestEffort:   true,
		Source:       &model.Source{Repo: "owner/missing-repo", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}
	return cat, reg
}

// Task 3.1 RED — a StepStatusDegraded event renders a distinct glyph (⚠), NOT the ✗
// used for hard failures.
func TestHeadlessExecutor_DegradedGlyph(t *testing.T) {
	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	// Skill install always fails → best-effort → StepStatusDegraded event.
	restoreSkill := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		h model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		return nil, errors.New("source.path not found")
	})
	defer restoreSkill()

	cat, reg := buildDegradedCatReg()
	var out bytes.Buffer
	exitCode := headless.RunHeadless(headless.ParsedFlags{
		Yes:           true,
		HomeDir:       t.TempDir(),
		NoSelfInstall: true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeFull,
		},
	}, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("degraded (best-effort) run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	// Must contain the distinct degraded glyph (⚠), not the hard-failure glyph (✗).
	if !strings.Contains(output, "⚠") {
		t.Errorf("output must contain ⚠ glyph for degraded step; got:\n%s", output)
	}
}

// Task 3.3 RED — after a successful run with degraded harnesses, the end-of-run
// summary enumerates each degraded harness (id + reason) AND exit code is 0.
func TestHeadlessExecutor_DegradedSummary(t *testing.T) {
	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreSkill := install.SetSkillInstallFn(func(
		_ interface{},
		_ fs.FS,
		_ context.Context,
		h model.Harness,
		_ []skillinstaller.AgentAdapter,
		_, _ string,
	) ([]skillinstaller.Result, error) {
		return nil, errors.New("upstream path missing")
	})
	defer restoreSkill()

	cat, reg := buildDegradedCatReg()
	var out bytes.Buffer
	exitCode := headless.RunHeadless(headless.ParsedFlags{
		Yes:           true,
		HomeDir:       t.TempDir(),
		NoSelfInstall: true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeFull,
		},
	}, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("degraded run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	// Summary must contain the "Degraded" block.
	if !strings.Contains(output, "Degraded") {
		t.Errorf("output must contain 'Degraded' summary block; got:\n%s", output)
	}
	// Summary must name the degraded harness.
	if !strings.Contains(output, "best-effort-h") {
		t.Errorf("output must name the degraded harness 'best-effort-h'; got:\n%s", output)
	}
}

// Task 3.5 TRIANGULATE — clean run (no degradations) → summary reports unqualified
// success; no degraded block.
func TestHeadlessExecutor_CleanRun_NoDegradedBlock(t *testing.T) {
	h := model.Harness{
		ID:           "clean-ext",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFn(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	var out bytes.Buffer
	exitCode := headless.RunHeadless(headless.ParsedFlags{
		Yes:           true,
		HomeDir:       t.TempDir(),
		NoSelfInstall: true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeFull,
		},
	}, cat, reg, &out)

	if exitCode != 0 {
		t.Errorf("clean run must exit 0, got %d; output:\n%s", exitCode, out.String())
	}

	output := out.String()
	if strings.Contains(output, "Degraded") {
		t.Errorf("clean run must NOT contain 'Degraded' block; got:\n%s", output)
	}
}
