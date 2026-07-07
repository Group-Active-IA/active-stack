package headless

import (
	"fmt"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// StarterCatalog is the subset of *catalog.Catalog needed to resolve a
// starter id into install params. It lets BuildStarterInstallParams be
// exercised against fakes or the real embedded catalog (mirrors the
// package-main starterCatalog interface used by "starter add").
type StarterCatalog interface {
	// StarterByID looks up a starter by its id.
	StarterByID(id string) (model.Starter, bool)
	// AllStarters returns all starters in the catalog (used in error messages).
	AllStarters() []model.Starter
	// ResolveStarter expands a starter into its total harness set.
	ResolveStarter(id string) ([]model.Harness, error)
	// ResolveStarterMCPs aggregates MCPs across includes.
	ResolveStarterMCPs(id string) ([]model.MCP, error)
}

// BuildStarterInstallParams builds the ParsedFlags needed to run a starter
// install through RunHeadless / the windows install event-stream pipeline.
// It is the pure params-construction step extracted from runStarterAdd
// (design D1, windows-contract-hub-operations): resolve the starter, its
// harnesses and aggregated MCPs, then build the Project-target Custom-mode
// intent. Both the CLI "starter add" path and "windows starters install"
// share this helper so they cannot drift.
//
// buildPlanFn is forwarded verbatim into ParsedFlags.BuildPlanFn; callers
// that need a production BuildPlan (SkillsFS, verify hook) build it and pass
// it in — this helper does not wire defaults, keeping it a pure function of
// its inputs.
func BuildStarterInstallParams(
	flags ParsedStarterAddFlags,
	cat StarterCatalog,
	buildPlanFn func(install.Catalog, install.Intent, install.Options) (install.Plan, error),
) (ParsedFlags, error) {
	// 1. Look up the starter by id; error with available list if unknown.
	starter, ok := cat.StarterByID(flags.StarterID)
	if !ok {
		allStarters := cat.AllStarters()
		available := make([]string, 0, len(allStarters))
		for _, s := range allStarters {
			available = append(available, s.ID)
		}
		return ParsedFlags{}, fmt.Errorf("unknown starter %q. Available starters: %s",
			flags.StarterID, strings.Join(available, ", "))
	}

	// 2. Resolve harnesses via ResolveStarter (expands includes, dedup, stable order).
	harnesses, err := cat.ResolveStarter(flags.StarterID)
	if err != nil {
		return ParsedFlags{}, fmt.Errorf("resolve starter %q: %w", flags.StarterID, err)
	}

	// 3. Aggregate MCPs across includes.
	mcps, err := cat.ResolveStarterMCPs(flags.StarterID)
	if err != nil {
		return ParsedFlags{}, fmt.Errorf("resolve starter MCPs for %q: %w", flags.StarterID, err)
	}

	// 4. Build an effective starter that carries the fully aggregated MCP list
	// (root + includes, deduped by name). This is the value stored in
	// Options.Starter and consumed by BuildPlan to emit MCP write steps.
	effectiveStarter := &model.Starter{
		ID:          starter.ID,
		Name:        starter.Name,
		Description: starter.Description,
		Harnesses:   starter.Harnesses,
		Includes:    starter.Includes,
		MCPs:        mcps,
	}

	// 5. Derive harness ids from the resolved harness set.
	harnessIDs := make([]string, 0, len(harnesses))
	for _, h := range harnesses {
		harnessIDs = append(harnessIDs, h.ID)
	}

	// 6. Build install.Intent (Custom mode, resolved harness ids, focal agents).
	intent := install.Intent{
		Mode:   model.ModeCustom,
		Custom: harnessIDs,
		Agents: flags.Agents,
	}

	// 7. Build ParsedFlags. HomeDir is intentionally empty for project-target
	// installs: RunHeadless falls back to os.UserHomeDir() for any
	// machine-level dependency check, but the install paths resolve from
	// ProjectRoot (via Target=Project in Options).
	return ParsedFlags{
		TUI:         false,
		DryRun:      flags.DryRun,
		Yes:         flags.Yes,
		HomeDir:     "",
		Target:      model.Project,
		ProjectRoot: flags.ProjectPath,
		Starter:     effectiveStarter,
		Intent:      intent,
		BuildPlanFn: buildPlanFn,
	}, nil
}
