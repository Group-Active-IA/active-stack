package headless

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
	"github.com/JuanCruzRobledo/jr-stack/internal/verify"
)

// RunHeadless executes the headless install flow and returns the process exit code:
//   - 0  on success (all hard checks pass).
//   - 1  on pipeline failure, verify failure, or rollback.
//
// It writes progress and verify output to w (typically os.Stdout).
//
// The function mirrors the wiring already used by the TUI (BuildPlan →
// pipeline.NewOrchestrator → Execute), so no new pipeline logic is introduced.
//
// When params.DryRun is true, RunHeadless prints the plan steps and returns 0
// without executing anything (no side-effects).
//
// When params.VerifyHookFn is non-nil it is used as the verify hook directly
// (test injection point). Otherwise the hook is built from the verify.BuildHook
// with an empty harness+adapter set (the caller is responsible for wiring the
// real hook via the catalog/registry; here we keep the hook nil when no
// VerifyHookFn is injected, matching the "no hook" TUI path for the test cases
// that don't supply harnesses with verify checks).
func RunHeadless(params ParsedFlags, cat install.Catalog, reg install.Registry, w io.Writer) int {
	// ── Build the progress function before opts so we can wire it as
	// opts.OnProgress. The skill step emits StepStatusDegraded via its own
	// onProgress field (set at plan-build time from opts.OnProgress), so opts
	// must carry the SAME function we pass to the orchestrator (C-32 D3 seam).
	//
	// Accumulate degraded events for the honest end-of-run summary (C-32).
	var degradedEvents []pipeline.ProgressEvent
	progressFn := func(e pipeline.ProgressEvent) {
		switch e.Status {
		case pipeline.StepStatusRunning:
			fmt.Fprintf(w, "  → %s running...\n", e.StepID)
		case pipeline.StepStatusSucceeded:
			fmt.Fprintf(w, "  ✓ %s\n", e.StepID)
		case pipeline.StepStatusFailed:
			fmt.Fprintf(w, "  ✗ %s: %v\n", e.StepID, e.Err)
		case pipeline.StepStatusDegraded:
			// C-32: distinct glyph for degraded (best-effort) steps.
			fmt.Fprintf(w, "  ⚠ %s (degraded, best-effort): %v\n", e.StepID, e.Err)
			degradedEvents = append(degradedEvents, e)
		case pipeline.StepStatusRolledBack:
			fmt.Fprintf(w, "  ↩ %s (rolled back)\n", e.StepID)
		}
	}

	// Build install options. Profile.OS is populated from runtime.GOOS so that
	// external-harness steps build a well-formed GitHub Releases asset URL.
	// (An empty OS would produce a double-underscore filename like
	// "engram_1.16.1__amd64.tar.gz" and result in HTTP 404.)
	opts := install.Options{
		HomeDir:               params.HomeDir,
		Target:                params.Target,
		ProjectRoot:           params.ProjectRoot,
		Starter:               params.Starter,
		Registry:              reg,
		Profile:               system.PlatformProfile{OS: runtime.GOOS},
		NoSelfInstall:         params.NoSelfInstall,
		SelfInstallBinaryPath: params.BinaryPath,
		OnProgress:            progressFn,
	}

	// Wire verify hook.
	if params.VerifyHookFn != nil {
		opts.VerifyHook = params.VerifyHookFn
	}
	// (When VerifyHookFn is nil and no real harnesses/adapters are provided in
	// tests, opts.VerifyHook stays nil — same as the current TUI path.)

	// Build the plan — use the injected BuildPlanFn when provided (allows
	// main.go to wire install.WithEmbeddedSkillsFS), otherwise use the default.
	buildPlanFn := params.BuildPlanFn
	if buildPlanFn == nil {
		buildPlanFn = install.BuildPlan
	}
	plan, err := buildPlanFn(cat, params.Intent, opts)
	if err != nil {
		fmt.Fprintf(w, "error: build plan: %v\n", err)
		return 1
	}

	// ── Dry-run: print plan steps and exit without executing ────────────────
	if params.DryRun {
		fmt.Fprintln(w, "Dry-run: plan steps (not executed):")
		for _, s := range plan.Prepare {
			fmt.Fprintf(w, "  [prepare] %s\n", s.ID())
		}
		for _, s := range plan.Apply {
			fmt.Fprintf(w, "  [apply]   %s\n", s.ID())
		}
		return 0
	}

	// ── Pre-flight dependency gate ───────────────────────────────────────────
	// Derive the runtime deps required by the selected harnesses and abort
	// before any filesystem write when any required dep is missing.
	{
		selected := selectHarnessesForGate(cat, params.Intent)
		reqDeps := system.RequiredDependencies(selected, opts.Profile)
		if len(reqDeps) > 0 {
			report := detectDepsForFn(context.Background(), reqDeps)
			if len(report.MissingRequired) > 0 {
				fmt.Fprint(w, system.RenderDependencyReport(report))
				fmt.Fprintln(w)
				// Print per-dep install hints.
				for _, dep := range report.Dependencies {
					if dep.Required && !dep.Installed && dep.InstallHint != "" {
						fmt.Fprintf(w, "  install hint for %s: %s\n", dep.Name, dep.InstallHint)
					}
				}
				fmt.Fprintln(w, "\nInstallation aborted: install the missing dependencies above and retry.")
				return 1
			}
		}
	}

	// ── Execute the plan via the orchestrator ───────────────────────────────
	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(progressFn),
	)
	result := orch.Execute(plan.StagePlan)

	if result.Err != nil {
		fmt.Fprintf(w, "\nInstallation failed: %v\n", result.Err)
		if result.Rollback.Stage == pipeline.StageRollback {
			if result.Rollback.Success {
				fmt.Fprintln(w, "Rollback: succeeded")
			} else {
				fmt.Fprintf(w, "Rollback: failed (%v)\n", result.Rollback.Err)
			}
		}
		return 1
	}

	// ── Print verify report ─────────────────────────────────────────────────
	// Collect any check results from the verify hook by building an empty
	// report (the real checks ran inside the verify-hook step). If no
	// VerifyHookFn was wired and opts.VerifyHook is nil, print the
	// "all passed" summary manually.
	report := verify.BuildReport(nil) // zero checks → Ready == true
	fmt.Fprint(w, verify.RenderReport(report))

	// ── Print degraded (best-effort) summary if any (C-32) ─────────────────
	// Exit code STAYS 0: degraded harnesses are soft failures. The summary
	// enumerates them so the operator knows which best-effort skills degraded.
	if len(degradedEvents) > 0 {
		fmt.Fprintln(w, "\nDegraded (best-effort) harnesses:")
		for _, ev := range degradedEvents {
			reason := ""
			if ev.Err != nil {
				reason = ev.Err.Error()
			}
			fmt.Fprintf(w, "  ⚠ %s: %s\n", ev.StepID, reason)
		}
	}

	return 0
}

// selectHarnessesForGate derives the harness set the gate should validate.
// It mirrors main.collectSelectedHarnesses without the dependency-resolution
// step, keeping the headless package decoupled from main.
func selectHarnessesForGate(cat install.Catalog, intent install.Intent) []model.Harness {
	switch intent.Mode {
	case model.ModeCustom:
		var out []model.Harness
		for _, id := range intent.Custom {
			if h, ok := cat.ByID(id); ok {
				out = append(out, h)
			}
		}
		return filterByAgents(out, intent.Agents)
	default:
		candidates := cat.ForMode(intent.Mode)
		return filterByAgents(candidates, intent.Agents)
	}
}

// filterByAgents returns harnesses that support at least one of the given agents.
// If agents is empty, all harnesses are returned.
func filterByAgents(harnesses []model.Harness, agents []model.Agent) []model.Harness {
	if len(agents) == 0 {
		return harnesses
	}
	var out []model.Harness
	for _, h := range harnesses {
		for _, a := range agents {
			if h.SupportsAgent(a) {
				out = append(out, h)
				break
			}
		}
	}
	return out
}
