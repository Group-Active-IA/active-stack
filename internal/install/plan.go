package install

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/planner"
)

// BuildPlan converts an Intent into a Plan ready for execution.
//
// Stages:
//   - Prepare: one snapshotStep that captures all config paths that Apply will write.
//   - Apply: one harnessStep per resolved harness, in topological dependency order.
//
// The plan is headless: no TUI, no progress display. Progress and verify are
// extension points supplied via Options.
//
// The returned Plan also carries the ProgressFunc from opts so callers can wire
// it into pipeline.NewOrchestrator via pipeline.WithProgressFunc(plan.OnProgress).
func BuildPlan(cat Catalog, intent Intent, opts Options) (Plan, error) {
	empty := Plan{}

	// 1. Select harnesses from catalog.
	selected, err := selectHarnesses(cat, intent)
	if err != nil {
		return empty, err
	}

	if len(selected) == 0 {
		return Plan{OnProgress: opts.OnProgress}, nil
	}

	// 2. Collect all harnesses for the dependency resolver.
	allHarnesses := make(map[string]model.Harness, len(cat.ForMode(model.ModeCustom)))
	for _, h := range cat.ForMode(model.ModeCustom) {
		allHarnesses[h.ID] = h
	}

	// 3. Resolve dependencies and topological order.
	resolved, err := planner.NewResolver().Resolve(selected, allHarnesses)
	if err != nil {
		return empty, fmt.Errorf("install: resolve dependencies: %w", err)
	}

	// 4. Collect adapters for the intent's agents.
	adapters := make([]AgentAdapter, 0, len(intent.Agents))
	for _, agent := range intent.Agents {
		adapter, ok := opts.Registry.Get(agent)
		if !ok {
			return empty, fmt.Errorf("install: no adapter registered for agent %q", agent)
		}
		adapters = append(adapters, adapter)
	}

	// 5. Build the Apply steps in topological order.
	applySteps := make([]pipeline.Step, 0, len(resolved.OrderedIDs))
	for _, id := range resolved.OrderedIDs {
		h, ok := cat.ByID(id)
		if !ok {
			return empty, fmt.Errorf("install: harness %q not found in catalog", id)
		}
		step, err := buildHarnessStep(h, adapters, opts)
		if err != nil {
			return empty, fmt.Errorf("install: build step for harness %q: %w", id, err)
		}
		applySteps = append(applySteps, step)
	}

	// 6. Collect all paths that Apply steps will write; build the snapshot step.
	snapshotDir := filepath.Join(opts.HomeDir, ".jr-stack", "backups", "install")
	prepareSteps := []pipeline.Step{
		&snapshotStep{
			id:         "snapshot",
			paths:      collectWritePaths(adapters, opts.HomeDir, resolved.OrderedIDs, cat),
			snapDir:    snapshotDir,
			snapCreate: snapshotCreate,
		},
	}

	// 7. Wire the snapshot manifest into RollbackStep Apply steps.
	// The snapshot is passed by pointer so the Apply steps see it after Prepare runs.
	manifestPtr := new(backup.Manifest)
	if ss, ok := prepareSteps[0].(*snapshotStep); ok {
		ss.manifestOut = manifestPtr
	}
	for _, s := range applySteps {
		if rs, ok := s.(manifestReceiver); ok {
			rs.setManifest(manifestPtr)
		}
	}

	// 8. Wrap the verify hook as the final Apply step if provided.
	if opts.VerifyHook != nil {
		applySteps = append(applySteps, &verifyStep{fn: opts.VerifyHook})
	}

	return Plan{
		StagePlan: pipeline.StagePlan{
			Prepare: prepareSteps,
			Apply:   applySteps,
		},
		OnProgress: opts.OnProgress,
	}, nil
}

// selectHarnesses returns the harness set to install for the given intent.
func selectHarnesses(cat Catalog, intent Intent) ([]model.Harness, error) {
	switch intent.Mode {
	case model.ModeCustom:
		var out []model.Harness
		for _, id := range intent.Custom {
			h, ok := cat.ByID(id)
			if !ok {
				return nil, fmt.Errorf("install: custom harness %q not found in catalog", id)
			}
			out = append(out, h)
		}
		return filterByAgents(out, intent.Agents), nil

	default:
		// Lite or Full: select by mode, then intersect with agents.
		candidates := cat.ForMode(intent.Mode)
		return filterByAgents(candidates, intent.Agents), nil
	}
}

// filterByAgents returns the subset of harnesses that support at least one of
// the given agents. If agents is empty, all harnesses are returned as-is.
func filterByAgents(harnesses []model.Harness, agents []model.Agent) []model.Harness {
	if len(agents) == 0 {
		return harnesses
	}
	var out []model.Harness
	for _, h := range harnesses {
		for _, agent := range agents {
			if h.SupportsAgent(agent) {
				out = append(out, h)
				break
			}
		}
	}
	return out
}

// collectWritePaths enumerates the filesystem paths that the Apply steps will
// write, so the snapshot in Prepare captures exactly those paths.
func collectWritePaths(adapters []AgentAdapter, homeDir string, orderedIDs []string, cat Catalog) []string {
	seen := make(map[string]struct{})
	var paths []string

	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			paths = append(paths, p)
		}
	}

	for _, id := range orderedIDs {
		h, ok := cat.ByID(id)
		if !ok {
			continue
		}
		switch h.Type {
		case model.HarnessConfig:
			for _, a := range adapters {
				add(a.InstructionsPath(homeDir))
			}
		case model.HarnessSkill:
			for _, a := range adapters {
				add(a.SkillsDir(homeDir))
			}
		case model.HarnessExternal:
			for _, a := range adapters {
				if h.External != nil && h.External.Method == "mcp" {
					add(a.MCPConfigPath(homeDir, h.ID))
				}
			}
		}
		// permissions harness (id == "permissions") — settings paths.
		if id == "permissions" {
			for _, a := range adapters {
				add(a.SettingsPath(homeDir))
			}
		}
	}

	return paths
}

// snapshotCreate is the backing function for creating snapshots. It is a
// package-level variable so tests can inject a fake without reopening backup.
var snapshotCreate = func(snapshotDir string, paths []string) (backup.Manifest, error) {
	return backup.NewSnapshotter().Create(snapshotDir, paths)
}

// restoreFn is the backing function for restoring from a snapshot.
// It is a package-level variable so tests can inject a fake.
var restoreFn = func(m backup.Manifest) error {
	return backup.RestoreService{}.Restore(m)
}

// buildHarnessStep constructs the correct pipeline.Step for a single harness.
// The "permissions" config harness is special: it uses the permissions installer.
func buildHarnessStep(h model.Harness, adapters []AgentAdapter, opts Options) (pipeline.Step, error) {
	switch h.Type {
	case model.HarnessExternal:
		return &externalStep{
			h:        h,
			adapters: adapters,
			homeDir:  opts.HomeDir,
		}, nil

	case model.HarnessSkill:
		return &skillStep{
			h:          h,
			adapters:   adapters,
			homeDir:    opts.HomeDir,
			backupDir:  filepath.Join(opts.HomeDir, ".jr-stack", "backups", "skills", h.ID),
			embeddedFS: opts.embeddedSkillsFS,
		}, nil

	case model.HarnessConfig:
		if h.ID == "permissions" {
			return &permissionsStep{
				h:        h,
				adapters: adapters,
				homeDir:  opts.HomeDir,
			}, nil
		}
		return &configStep{
			h:        h,
			adapters: adapters,
			homeDir:  opts.HomeDir,
		}, nil

	default:
		return nil, fmt.Errorf("unknown harness type %q for harness %q", h.Type, h.ID)
	}
}

// manifestReceiver is implemented by steps that need the snapshot manifest for
// rollback. The manifest is set after Prepare runs.
type manifestReceiver interface {
	setManifest(m *backup.Manifest)
}

// verifyStep wraps the optional verify hook as the final Apply step.
type verifyStep struct {
	fn func() error
}

func (s *verifyStep) ID() string   { return "verify-hook" }
func (s *verifyStep) Run() error   { return s.fn() }

// skillRunner is the interface required by skill.NewInstaller.
type skillRunner interface {
	Run(ctx context.Context, args []string) error
}
