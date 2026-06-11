package install_test

// skill_runner_test.go — regression tests for the nil-runner panic in skillStep.
//
// Bug: buildHarnessStep constructed a skillStep without setting the runner field.
// For harnesses with Source.Method "clone", the real skillInstallFn
// forwards that nil runner to skill.NewInstaller → cloneInstaller
// → runner.Run() → nil pointer dereference panic.
//
// Existing tests in steps_test.go and integration_test.go all replace
// skillInstallFn via SetSkillInstallFn, so the nil runner was never reached.
//
// These tests call the REAL skillInstallFn by NOT replacing it, verifying that
// after the fix the stub runner is invoked (not nil-dereferenced).

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	skillinstaller "github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// Stub runner
// ─────────────────────────────────────────────────────────────────────────────

// installStubRunner satisfies both install.CmdRunner and skill.Runner.
// It records that it was called and returns a predictable error so that
// clone logic halts without real network or disk access.
// The panic-before-fix occurs BEFORE this error is reached — the nil
// dereference happens on the first runner.Run call.
type installStubRunner struct {
	called bool
}

func (r *installStubRunner) Run(_ context.Context, _ []string) error {
	r.called = true
	return errors.New("stub runner: intentional early exit")
}

// Compile-time interface checks.
var _ install.CmdRunner = (*installStubRunner)(nil)
var _ skillinstaller.Runner = (*installStubRunner)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSkillStep_CloneMethod_RunnerIsNotNil(t *testing.T) {
	// BEFORE fix: panics with nil pointer dereference at clone.go:37
	// AFTER  fix: stub runner is called; returns a non-nil error without panicking
	stub := &installStubRunner{}

	h := model.Harness{
		ID:           "agent-instruction",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/agent-instruction", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(noopSnapshotFn)
	defer restoreSnap()

	// Inject the stub runner via Options so BuildPlan wires it into the skillStep.
	opts := install.WithCmdRunner(buildOptions(homeDir, reg, nil), stub)

	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeFull,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	runPlanStep(t, plan, "skill:agent-instruction", func(runErr error) {
		// A non-nil error is expected: stub runner stops the clone.
		// A nil error would mean git actually ran — that must not happen.
		if runErr == nil {
			t.Fatal("expected stub runner error, got nil (real git ran?)")
		}
		if !stub.called {
			t.Error("stub runner must be invoked for clone method")
		}
	})
}

func TestSkillStep_EmbedMethod_NoRunnerNeeded(t *testing.T) {
	// Regression guard: the fix must not break the embed path.
	// Embed skills do not use the runner, so nil is acceptable.
	h := model.Harness{
		ID:           "embed-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Method: "embed"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	embeddedFS := fstest.MapFS{
		"skills/embed-skill/SKILL.md": &fstest.MapFile{Data: []byte("# embed skill")},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(noopSnapshotFn)
	defer restoreSnap()

	// No CmdRunner injected — embed does not need one.
	opts := install.WithEmbeddedSkillsFS(buildOptions(homeDir, reg, nil), embeddedFS)

	plan, err := install.BuildPlan(cat, install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeFull,
	}, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	runPlanStep(t, plan, "skill:embed-skill", func(runErr error) {
		if runErr != nil {
			t.Errorf("embed skill step must succeed, got: %v", runErr)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// runPlanStep finds the step with stepID in plan.Apply, runs it inside a
// deferred recover (converting a panic into a test failure), then invokes
// check with the error returned by Run().
func runPlanStep(t *testing.T, plan install.Plan, stepID string, check func(error)) {
	t.Helper()
	for _, step := range plan.Apply {
		if step.ID() != stepID {
			continue
		}
		var runErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC in step %q (nil runner bug not fixed): %v", stepID, r)
				}
			}()
			runErr = step.Run()
		}()
		check(runErr)
		return
	}
	t.Fatalf("step %q not found in plan.Apply", stepID)
}

// noopSnapshotFn is a SetSnapshotCreate-compatible no-op.
func noopSnapshotFn(dir string, paths []string) (backup.Manifest, error) {
	return backup.Manifest{}, nil
}

// Compile-time guard.
var _ fs.FS = fstest.MapFS{}
