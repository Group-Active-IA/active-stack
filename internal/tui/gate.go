package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// tuiDetectDepsForFn is the function the TUI uses to detect dependencies
// before starting the install. It defaults to system.DetectDepsFor and can be
// replaced in tests via setTUIDetectDepsForFn.
var tuiDetectDepsForFn = func(ctx context.Context, deps []system.Dependency) system.DependencyReport {
	return system.DetectDepsFor(ctx, deps)
}

// setTUIDetectDepsForFn replaces the TUI dependency-detection function for
// testing. It returns a restore function that resets the original.
func setTUIDetectDepsForFn(fn func(ctx context.Context, deps []system.Dependency) system.DependencyReport) (restore func()) {
	old := tuiDetectDepsForFn
	tuiDetectDepsForFn = fn
	return func() { tuiDetectDepsForFn = old }
}

// checkPreflightDeps derives the required deps for the intent's selected
// harnesses and detects them. Returns an error if any required dep is missing;
// returns nil if the gate passes (nothing required, or all present).
func (m Model) checkPreflightDeps() error {
	if m.deps.Catalog == nil {
		return nil
	}

	intent := m.Selection.BuildIntent()
	selected := selectTUIHarnesses(m.deps.Catalog, intent)
	// Use an empty profile — the harness→deps mapping is profile-independent
	// for the set of required names (node/npm/npx/git). Profile only affects
	// InstallHint text, which comes from defineDependencies inside RequiredDependencies.
	profile := system.PlatformProfile{}
	reqDeps := system.RequiredDependencies(selected, profile)
	if len(reqDeps) == 0 {
		return nil
	}

	report := tuiDetectDepsForFn(context.Background(), reqDeps)
	if len(report.MissingRequired) > 0 {
		return fmt.Errorf("missing required dependencies: %s\n\n%s",
			strings.Join(report.MissingRequired, ", "),
			system.RenderDependencyReport(report),
		)
	}
	return nil
}

// selectTUIHarnesses derives the harness set matching the given intent from
// the catalog. It mirrors selectHarnessesForGate in the headless package,
// keeping the tui package self-contained.
func selectTUIHarnesses(cat install.Catalog, intent install.Intent) []model.Harness {
	switch intent.Mode {
	case model.ModeCustom:
		var out []model.Harness
		for _, id := range intent.Custom {
			if h, ok := cat.ByID(id); ok {
				out = append(out, h)
			}
		}
		return filterHarnessesByAgents(out, intent.Agents)
	default:
		candidates := cat.ForMode(intent.Mode)
		return filterHarnessesByAgents(candidates, intent.Agents)
	}
}
