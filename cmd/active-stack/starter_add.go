package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// starterCatalog is the subset of *catalog.Catalog methods needed by the
// starter add handler and dispatch. This allows both to be tested with fakes or
// the real embedded catalog.
type starterCatalog interface {
	install.Catalog
	// StarterByID looks up a starter by its id.
	StarterByID(id string) (model.Starter, bool)
	// AllStarters returns all starters in the catalog (used in error messages).
	AllStarters() []model.Starter
	// ResolveStarter expands a starter into its total harness set.
	ResolveStarter(id string) ([]model.Harness, error)
	// ResolveStarterMCPs aggregates MCPs across includes.
	ResolveStarterMCPs(id string) ([]model.MCP, error)
}

// runStarterAdd implements the "starter add <id>" handler.
// It is extracted from main() to allow headless testing.
//
// Parameters:
//   - flags: parsed flags from ParseStarterAddFlags (ProjectPath must already be
//     absolutized and validated by the caller via ResolveProjectRoot).
//   - cat: the (embedded) catalog — must implement starterCatalog.
//   - reg: the agent registry (install.Registry).
//   - buildPlanFn: optional override for install.BuildPlan (inject SkillsFS, etc.).
//     When nil, the real install.BuildPlan with assets.SkillsFS is used.
//   - w: the output writer (os.Stdout in production).
//
// Returns the process exit code (0 = success, 1 = failure).
func runStarterAdd(
	flags headless.ParsedStarterAddFlags,
	cat starterCatalog,
	reg install.Registry,
	buildPlanFn func(install.Catalog, install.Intent, install.Options) (install.Plan, error),
	w io.Writer,
) int {
	// 1. Look up the starter by id; error with available list if unknown.
	starter, ok := cat.StarterByID(flags.StarterID)
	if !ok {
		allStarters := cat.AllStarters()
		available := make([]string, 0, len(allStarters))
		for _, s := range allStarters {
			available = append(available, s.ID)
		}
		fmt.Fprintf(w, "error: unknown starter %q. Available starters: %s\n",
			flags.StarterID, strings.Join(available, ", "))
		return 1
	}

	// 2. Resolve harnesses via ResolveStarter (expands includes, dedup, stable order).
	harnesses, err := cat.ResolveStarter(flags.StarterID)
	if err != nil {
		fmt.Fprintf(w, "error: resolve starter %q: %v\n", flags.StarterID, err)
		return 1
	}

	// 3. Aggregate MCPs across includes (D3a).
	mcps, err := cat.ResolveStarterMCPs(flags.StarterID)
	if err != nil {
		fmt.Fprintf(w, "error: resolve starter MCPs for %q: %v\n", flags.StarterID, err)
		return 1
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

	// 7. Wire BuildPlanFn: inject SkillsFS and a no-op verify hook when no override
	// is provided. The real verify hook (with agent adapters) is wired by the
	// production call site in main.go; tests that pass nil get a safe no-op.
	if buildPlanFn == nil {
		buildPlanFn = func(c install.Catalog, i install.Intent, opts install.Options) (install.Plan, error) {
			opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)
			if opts.VerifyHook == nil {
				// No-op verify hook: no harnesses to check here — verify is only
				// meaningful after a non-dry-run install, and is wired from main.go.
				opts.VerifyHook = verify.BuildHook(nil, nil, opts.HomeDir)
			}
			return install.BuildPlan(c, i, opts)
		}
	}

	// 8. Build ParsedFlags for RunHeadless (reuses existing headless path — D5-A).
	// HomeDir is intentionally empty for project-target installs: RunHeadless
	// falls back to os.UserHomeDir() for any machine-level dependency check, but
	// the install paths resolve from ProjectRoot (via Target=Project in Options).
	params := headless.ParsedFlags{
		TUI:         false,
		DryRun:      flags.DryRun,
		Yes:         flags.Yes,
		HomeDir:     "",
		Target:      model.Project,
		ProjectRoot: flags.ProjectPath,
		Starter:     effectiveStarter,
		Intent:      intent,
		BuildPlanFn: buildPlanFn,
	}

	// 9. Execute via RunHeadless (gate + snapshot + orchestrator + rollback + dry-run).
	exitCode := headless.RunHeadless(params, cat, reg, w)
	if exitCode == 0 && !flags.DryRun {
		fmt.Fprintf(w, "\nStarter %q applied to %s (agents: %s)\n",
			flags.StarterID, flags.ProjectPath, agentListStr(flags.Agents))
	}
	return exitCode
}

// agentListStr formats a slice of agents as a comma-separated string.
func agentListStr(agents []model.Agent) string {
	parts := make([]string, len(agents))
	for i, a := range agents {
		parts[i] = string(a)
	}
	return strings.Join(parts, ", ")
}
