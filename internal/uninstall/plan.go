package uninstall

import (
	"fmt"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/command"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// BuildPlan converts an Intent into a Plan ready for execution.
//
// Stages:
//   - Prepare: one snapshotStep capturing all paths the Apply steps will touch.
//   - Apply: one step per resolved harness, typed by the harness's type.
//
// The plan is headless: no TUI, no progress display. Progress is an extension
// point supplied via Options.
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

	// 2. Collect adapters for the intent's agents.
	adapters := make([]AgentAdapter, 0, len(intent.Agents))
	for _, agent := range intent.Agents {
		adapter, ok := opts.Registry.Get(agent)
		if !ok {
			return empty, fmt.Errorf("uninstall: no adapter registered for agent %q", agent)
		}
		adapters = append(adapters, adapter)
	}

	// 3. Handle StrategyRestore: a single restoreStep replays the install-time backup.
	if intent.Strategy == StrategyRestore {
		if opts.RestoreManifest == nil {
			return empty, fmt.Errorf("uninstall: StrategyRestore requires RestoreManifest in Options")
		}
		applySteps := []pipeline.Step{
			&restoreStep{installManifest: *opts.RestoreManifest},
		}
		return Plan{
			StagePlan: pipeline.StagePlan{
				Apply: applySteps,
			},
			OnProgress: opts.OnProgress,
		}, nil
	}

	// 4. StrategyTargeted: build per-harness reversal steps.
	applySteps := make([]pipeline.Step, 0, len(selected))
	for _, h := range selected {
		step, err := buildUninstallStep(h, adapters, opts)
		if err != nil {
			return empty, fmt.Errorf("uninstall: build step for harness %q: %w", h.ID, err)
		}
		applySteps = append(applySteps, step)
	}

	// 5. Collect all paths that Apply steps will touch; build the snapshot step.
	snapshotDir := filepath.Join(opts.HomeDir, ".jr-stack", "backups", "uninstall")
	prepareSteps := []pipeline.Step{
		&snapshotStep{
			id:         "uninstall-snapshot",
			paths:      collectUninstallPaths(adapters, opts.HomeDir, selected),
			snapDir:    snapshotDir,
			snapCreate: snapshotCreate,
		},
	}

	// 6. Wire the snapshot manifest into RollbackStep Apply steps via pointer.
	// The snapshot is passed by pointer so Apply steps see it after Prepare runs.
	manifestPtr := new(backup.Manifest)
	if ss, ok := prepareSteps[0].(*snapshotStep); ok {
		ss.manifestOut = manifestPtr
	}
	for _, s := range applySteps {
		if rs, ok := s.(manifestReceiver); ok {
			rs.setManifest(manifestPtr)
		}
	}

	return Plan{
		StagePlan: pipeline.StagePlan{
			Prepare: prepareSteps,
			Apply:   applySteps,
		},
		OnProgress: opts.OnProgress,
	}, nil
}

// selectHarnesses returns the harness set to uninstall for the given intent.
func selectHarnesses(cat Catalog, intent Intent) ([]model.Harness, error) {
	switch intent.Mode {
	case model.ModeCustom:
		var out []model.Harness
		for _, id := range intent.Custom {
			h, ok := cat.ByID(id)
			if !ok {
				return nil, fmt.Errorf("uninstall: custom harness %q not found in catalog", id)
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

// collectUninstallPaths enumerates the filesystem paths that the Apply steps
// will touch, so the snapshot in Prepare captures exactly those paths.
func collectUninstallPaths(adapters []AgentAdapter, homeDir string, harnesses []model.Harness) []string {
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

	for _, h := range harnesses {
		switch h.Type {
		case model.HarnessConfig:
			if h.ID == "permissions" {
				for _, a := range adapters {
					add(a.SettingsPath(homeDir))
				}
			} else {
				for _, a := range adapters {
					add(a.InstructionsPath(homeDir))
					// Primary-agent delivery also touches the settings JSON, so
					// snapshot it too for a complete rollback point.
					if a.ConfigDelivery() == model.ConfigDeliveryPrimaryAgent {
						add(a.SettingsPath(homeDir))
					}
				}
			}
		case model.HarnessSkill:
			for _, a := range adapters {
				add(filepath.Join(a.SkillsDir(homeDir), h.ID))
			}
		case model.HarnessExternal:
			// External harnesses are skipped — no paths to snapshot.
		case model.HarnessCommand:
			// Command harnesses write a single file per adapter under CommandsDir.
			// Skip adapters with empty CommandsDir or unknown VariantKey.
			for _, a := range adapters {
				commandsDir := a.CommandsDir(homeDir)
				if commandsDir == "" {
					continue
				}
				relPath := command.RelPathForVariant(a.VariantKey())
				if relPath == "" {
					continue
				}
				add(filepath.Join(commandsDir, relPath))
			}
		}
	}

	return paths
}

// buildUninstallStep constructs the correct pipeline.Step for a single harness.
func buildUninstallStep(h model.Harness, adapters []AgentAdapter, opts Options) (pipeline.Step, error) {
	switch h.Type {
	case model.HarnessExternal:
		return &externalSkipStep{h: h}, nil

	case model.HarnessSkill:
		return &skillRemovalStep{
			h:        h,
			adapters: adapters,
			homeDir:  opts.HomeDir,
		}, nil

	case model.HarnessConfig:
		if h.ID == "permissions" {
			return &permissionsRemovalStep{
				h:        h,
				adapters: adapters,
				homeDir:  opts.HomeDir,
			}, nil
		}
		return &markerRemovalStep{
			h:        h,
			adapters: adapters,
			homeDir:  opts.HomeDir,
		}, nil

	case model.HarnessCommand:
		return &commandRemovalStep{
			h:        h,
			adapters: adapters,
			homeDir:  opts.HomeDir,
		}, nil

	default:
		return nil, fmt.Errorf("unknown harness type %q for harness %q", h.Type, h.ID)
	}
}
