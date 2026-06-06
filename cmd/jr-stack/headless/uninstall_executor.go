package headless

import (
	"fmt"
	"io"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// RunHeadlessUninstall executes the headless uninstall flow and returns the
// process exit code:
//   - 0 on success (all steps complete).
//   - 1 on pipeline failure, BuildPlan error, or rollback.
//
// It writes progress and error output to w (typically os.Stderr).
//
// This is a slimmer sibling of RunHeadless (D2): it carries no dependency gate,
// no verify hook, and no injectable BuildPlanFn — the uninstall engine needs none
// of these. The function calls uninstall.BuildPlan directly.
//
// When params.DryRun is true, RunHeadlessUninstall prints the plan steps and
// returns 0 without executing anything (no side-effects).
//
// The function is EXPORTED so the future tui-menu-hub change can reuse the exact
// execution path from its ScreenUninstall screen.
func RunHeadlessUninstall(params ParsedUninstallFlags, cat uninstall.Catalog, reg uninstall.Registry, w io.Writer) int {
	// ── Build the progress function ─────────────────────────────────────────
	progressFn := func(e pipeline.ProgressEvent) {
		switch e.Status {
		case pipeline.StepStatusRunning:
			fmt.Fprintf(w, "  → %s running...\n", e.StepID)
		case pipeline.StepStatusSucceeded:
			fmt.Fprintf(w, "  ✓ %s\n", e.StepID)
		case pipeline.StepStatusFailed:
			fmt.Fprintf(w, "  ✗ %s: %v\n", e.StepID, e.Err)
		case pipeline.StepStatusRolledBack:
			fmt.Fprintf(w, "  ↩ %s (rolled back)\n", e.StepID)
		}
	}

	// ── Build uninstall options ──────────────────────────────────────────────
	opts := uninstall.Options{
		HomeDir:    params.HomeDir,
		Registry:   reg,
		OnProgress: progressFn,
	}

	// Wire RestoreManifest when strategy is restore (D6).
	// The executor reads the manifest from RestoreManifestPath so that
	// RunHeadlessUninstall is self-contained and reusable by the TUI hub.
	// The dispatch layer validates that the path is non-empty before calling us
	// (task 3.2), but we defensively skip the read when the path is empty
	// (BuildPlan will error on nil RestoreManifest, which is the correct behavior).
	if params.Intent.Strategy == uninstall.StrategyRestore && params.RestoreManifestPath != "" {
		m, err := backup.ReadManifest(params.RestoreManifestPath)
		if err != nil {
			fmt.Fprintf(w, "error: read manifest: %v\n", err)
			return 1
		}
		opts.RestoreManifest = &m
	}

	// ── Build the plan ───────────────────────────────────────────────────────
	plan, err := uninstall.BuildPlan(cat, params.Intent, opts)
	if err != nil {
		fmt.Fprintf(w, "error: build plan: %v\n", err)
		return 1
	}

	// ── Dry-run: print plan steps and exit without executing ─────────────────
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

	// ── Execute the plan via the orchestrator ───────────────────────────────
	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(progressFn),
	)
	result := orch.Execute(plan.StagePlan)

	if result.Err != nil {
		fmt.Fprintf(w, "\nUninstall failed: %v\n", result.Err)
		if result.Rollback.Stage == pipeline.StageRollback {
			if result.Rollback.Success {
				fmt.Fprintln(w, "Rollback: succeeded")
			} else {
				fmt.Fprintf(w, "Rollback: failed (%v)\n", result.Rollback.Err)
			}
		}
		return 1
	}

	return 0
}
