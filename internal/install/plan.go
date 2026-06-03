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
	selected, err := SelectHarnesses(cat, intent)
	if err != nil {
		return empty, err
	}

	// When there are no harnesses to install, we may still have MCP steps
	// from a starter (C-28). Only return early if there's truly nothing to do.
	if len(selected) == 0 && (opts.Starter == nil || len(opts.Starter.MCPs) == 0) {
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

	// Resolve the effective base directory for this install target.
	// Machine (zero-value) → homeDir: identical to pre-C-27 behaviour.
	// Project              → projectRoot: routes writes to the project.
	effectiveBase := opts.HomeDir
	if opts.Target == model.Project && opts.ProjectRoot != "" {
		effectiveBase = opts.ProjectRoot
	}

	// 5. Build the Apply steps in topological order.
	// Normalize the tier defensively so no intent zero-value can produce bypass behavior.
	tier := intent.Tier.Normalize()
	applySteps := make([]pipeline.Step, 0, len(resolved.OrderedIDs))
	for _, id := range resolved.OrderedIDs {
		h, ok := cat.ByID(id)
		if !ok {
			return empty, fmt.Errorf("install: harness %q not found in catalog", id)
		}
		step, err := buildHarnessStep(h, adapters, opts, effectiveBase, tier)
		if err != nil {
			return empty, fmt.Errorf("install: build step for harness %q: %w", id, err)
		}
		applySteps = append(applySteps, step)
	}

	// 5b. Emit MCP write steps from Starter.MCPs[] (D5).
	// One step per (MCP × focused-agent), each with a Rollback() (governance ALTO).
	mcpSteps := buildMCPSteps(opts.Starter, adapters, effectiveBase, opts.Target)
	applySteps = append(applySteps, mcpSteps...)

	// 6. Collect all paths that Apply steps will write; build the snapshot step.
	// Snapshot dir follows the same effective base (project root for Project target).
	snapshotDir := filepath.Join(effectiveBase, ".jr-stack", "backups", "install")
	writePaths := collectWritePaths(adapters, effectiveBase, opts.Target, resolved.OrderedIDs, cat)
	// Also include MCP write paths from the starter (C-28 D5).
	collectStarterMCPPaths(opts.Starter, adapters, effectiveBase, opts.Target, &writePaths)
	dirHints := writePaths.DirHints
	prepareSteps := []pipeline.Step{
		&snapshotStep{
			id:      "snapshot",
			paths:   writePaths.Paths,
			snapDir: snapshotDir,
			snapCreate: func(dir string, paths []string) (backup.Manifest, error) {
				return snapshotCreateWithHints(dir, paths, dirHints)
			},
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

// SecurityFirstHarnessID is the id of the harness that is security-first: it is
// non-disableable in Custom mode and is always forced into the selection (C-21).
//
// This is the SINGLE SOURCE OF TRUTH for "what is forced". Every other layer —
// the TUI gate (selectTUIHarnesses), the Custom picker (tui/model.go) and the
// verify hook (cmd collectSelectedHarnesses) — references this const and/or
// delegates to SelectHarnesses, so the rule lives in exactly one place (C-24).
const SecurityFirstHarnessID = "permissions"

// SelectHarnesses returns the harness set to install for the given intent.
//
// It is the canonical, exported selector and the single source of truth for the
// security-first rule that forces SecurityFirstHarnessID in Custom mode
// (C-21/C-24). Semantics are STRICT: an unknown id in Custom mode is an error
// (safer than silently ignoring it). All callers — headless BuildPlan, the TUI
// gate, the verify hook — must go through this function.
func SelectHarnesses(cat Catalog, intent Intent) ([]model.Harness, error) {
	switch intent.Mode {
	case model.ModeCustom:
		var out []model.Harness
		seen := make(map[string]struct{}, len(intent.Custom)+1)
		for _, id := range intent.Custom {
			h, ok := cat.ByID(id)
			if !ok {
				return nil, fmt.Errorf("install: custom harness %q not found in catalog", id)
			}
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, h)
		}
		// C-21/C-24: the security-first harness is non-disableable in Custom.
		// We force it into the set even when the user didn't pick it. The final
		// filterByAgents drops it if the selected agent does not support it
		// (correct boundary: you cannot force an overlay that does not exist for
		// that agent). THIS is the only place the "force permissions" rule lives.
		if _, picked := seen[SecurityFirstHarnessID]; !picked {
			if perm, ok := cat.ByID(SecurityFirstHarnessID); ok {
				out = append(out, perm)
			}
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

// writePathsResult holds the paths and directory hints collected by collectWritePaths.
type writePathsResult struct {
	// Paths is the ordered, deduplicated list of filesystem paths the Apply
	// steps will write (both files and directories).
	Paths []string
	// DirHints is the set of paths known to be directories (even if they may
	// not exist on disk yet). Used by the snapshotter to record IsDir=true so
	// the restore can call RemoveAll instead of Remove on rollback.
	DirHints map[string]bool
}

// collectWritePaths enumerates the filesystem paths that the Apply steps will
// write, so the snapshot in Prepare captures exactly those paths.
// It also records which paths are directories (DirHints) so the snapshotter
// can distinguish them from files even when they don't exist on disk yet.
//
// For Machine target (zero-value), the existing per-method resolution is used
// (zero regression from C-27). For Project target, PathsFor is called with the
// effective base so each adapter resolves its project-specific layout.
func collectWritePaths(adapters []AgentAdapter, effectiveBase string, target model.InstallTarget, orderedIDs []string, cat Catalog) writePathsResult {
	seen := make(map[string]struct{})
	result := writePathsResult{DirHints: make(map[string]bool)}

	addFile := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			result.Paths = append(result.Paths, p)
		}
	}

	addDir := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			result.Paths = append(result.Paths, p)
		}
		result.DirHints[filepath.Clean(p)] = true
	}

	for _, id := range orderedIDs {
		h, ok := cat.ByID(id)
		if !ok {
			continue
		}
		switch h.Type {
		case model.HarnessConfig:
			for _, a := range adapters {
				addFile(resolvedInstructionsPath(a, effectiveBase, target))
			}
		case model.HarnessSkill:
			for _, a := range adapters {
				// SkillsDir is a directory path — must be tracked as a dir hint.
				addDir(resolvedSkillsDir(a, effectiveBase, target))
			}
		case model.HarnessExternal:
			for _, a := range adapters {
				if h.External != nil && h.External.Method == "mcp" {
					addFile(resolvedMCPConfigPath(a, effectiveBase, target, h.ID))
				}
			}
		case model.HarnessCommand:
			for _, a := range adapters {
				// CommandsDir is a directory — track as dir hint so rollback can RemoveAll.
				addDir(resolvedCommandsDir(a, effectiveBase, target))
			}
		}
		// permissions harness (id == "permissions") — settings paths.
		if id == "permissions" {
			for _, a := range adapters {
				addFile(resolvedSettingsPath(a, effectiveBase, target))
			}
		}
	}

	return result
}

// collectStarterMCPPaths appends the MCP config paths for all (MCP × adapter)
// pairs from the starter into the given writePathsResult (deduplicating).
// This ensures the snapshot captures these files before any MCP write steps run.
func collectStarterMCPPaths(starter *model.Starter, adapters []AgentAdapter, base string, target model.InstallTarget, result *writePathsResult) {
	if starter == nil || len(starter.MCPs) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(result.Paths))
	for _, p := range result.Paths {
		seen[p] = struct{}{}
	}
	for _, mcp := range starter.MCPs {
		for _, adapter := range adapters {
			p := adapter.PathsFor(base, target).MCPConfigPath(mcp.Name)
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			result.Paths = append(result.Paths, p)
		}
	}
}

// resolvedInstructionsPath returns the instructions file path for the given
// adapter, base directory, and install target.
// Machine: delegates to the existing InstructionsPath method (zero regression).
// Project: uses PathsFor to resolve the project layout.
func resolvedInstructionsPath(a AgentAdapter, base string, t model.InstallTarget) string {
	if t == model.Project {
		return a.PathsFor(base, t).InstructionsPath
	}
	return a.InstructionsPath(base)
}

// resolvedSkillsDir returns the skills directory path for the given adapter,
// base directory, and install target.
func resolvedSkillsDir(a AgentAdapter, base string, t model.InstallTarget) string {
	if t == model.Project {
		return a.PathsFor(base, t).SkillsDir
	}
	return a.SkillsDir(base)
}

// resolvedSettingsPath returns the settings file path for the given adapter,
// base directory, and install target.
func resolvedSettingsPath(a AgentAdapter, base string, t model.InstallTarget) string {
	if t == model.Project {
		return a.PathsFor(base, t).SettingsPath
	}
	return a.SettingsPath(base)
}

// resolvedMCPConfigPath returns the MCP config path for the given adapter,
// base directory, install target, and server name.
func resolvedMCPConfigPath(a AgentAdapter, base string, t model.InstallTarget, serverName string) string {
	if t == model.Project {
		return a.PathsFor(base, t).MCPConfigPath(serverName)
	}
	return a.MCPConfigPath(base, serverName)
}

// resolvedCommandsDir returns the command directory path for the given adapter,
// base directory, and install target. Mirrors resolvedSkillsDir exactly.
// Machine: delegates to adapter.CommandsDir (machine-target method).
// Project: uses PathsFor(base, Project).CommandsDir (project-specific layout).
// Added in C-31 (D1 + §6.3).
func resolvedCommandsDir(a AgentAdapter, base string, t model.InstallTarget) string {
	if t == model.Project {
		return a.PathsFor(base, t).CommandsDir
	}
	return a.CommandsDir(base)
}

// snapshotCreate is the backing function for creating snapshots. It is a
// package-level variable so tests can inject a fake without reopening backup.
var snapshotCreate = func(snapshotDir string, paths []string) (backup.Manifest, error) {
	return backup.NewSnapshotter().Create(snapshotDir, paths)
}

// snapshotCreateWithHints is the full-featured backing function used by BuildPlan.
// It passes directory hints to the snapshotter so it can record IsDir=true for
// skill-dir entries, enabling safe RemoveAll rollback for dirs created by the install.
// Tests that inject snapshotCreate via SetSnapshotCreate automatically override this too.
var snapshotCreateWithHints = func(snapshotDir string, paths []string, dirHints map[string]bool) (backup.Manifest, error) {
	return backup.NewSnapshotter().CreateWithDirHints(snapshotDir, paths, dirHints)
}

// restoreFn is the backing function for restoring from a snapshot.
// It is a package-level variable so tests can inject a fake.
var restoreFn = func(m backup.Manifest) error {
	return backup.RestoreService{}.Restore(m)
}

// buildMCPSteps constructs one mcpWriteStep per (MCP × adapter) pair from the
// starter's MCPs slice (D5). When starter is nil or has no MCPs, returns nil.
//
// Each step is routed via the target-aware resolvedMCPConfigPath so the correct
// per-agent, per-target path is used (e.g. <root>/.mcp.json for Claude project).
// Each step carries a Rollback() (governance ALTO — writes into user project).
func buildMCPSteps(starter *model.Starter, adapters []AgentAdapter, effectiveBase string, target model.InstallTarget) []pipeline.Step {
	if starter == nil || len(starter.MCPs) == 0 {
		return nil
	}

	steps := make([]pipeline.Step, 0, len(starter.MCPs)*len(adapters))
	for _, mcp := range starter.MCPs {
		for _, adapter := range adapters {
			paths := adapter.PathsFor(effectiveBase, target)
			configPath := paths.MCPConfigPath(mcp.Name)
			strategy := paths.MCPStrategy

			stepID := fmt.Sprintf("mcp:%s/%s", mcp.Name, adapter.Agent())
			steps = append(steps, &mcpWriteStep{
				id:         stepID,
				mcp:        mcp,
				configPath: configPath,
				strategy:   strategy,
			})
		}
	}
	return steps
}

// buildHarnessStep constructs the correct pipeline.Step for a single harness.
// The "permissions" config harness is special: it uses the permissions installer
// and receives the permission tier from the install intent.
//
// effectiveBase is the resolved base directory for this install (homeDir for
// Machine target, projectRoot for Project target). Steps store it as homeDir
// for backward compatibility with existing step implementations.
func buildHarnessStep(h model.Harness, adapters []AgentAdapter, opts Options, effectiveBase string, tier model.PermissionTier) (pipeline.Step, error) {
	switch h.Type {
	case model.HarnessExternal:
		return &externalStep{
			h:        h,
			adapters: adapters,
			homeDir:  effectiveBase,
			profile:  opts.Profile,
		}, nil

	case model.HarnessSkill:
		runner := opts.cmdRunner
		if runner == nil {
			runner = defaultCmdRunner{}
		}
		return &skillStep{
			h:          h,
			adapters:   adapters,
			homeDir:    effectiveBase,
			backupDir:  filepath.Join(effectiveBase, ".jr-stack", "backups", "skills", h.ID),
			embeddedFS: opts.embeddedSkillsFS,
			runner:     runner,
			bestEffort: h.BestEffort,
			onProgress: opts.OnProgress,
		}, nil

	case model.HarnessConfig:
		if h.ID == "permissions" {
			return &permissionsStep{
				h:        h,
				adapters: adapters,
				homeDir:  effectiveBase,
				tier:     tier,
			}, nil
		}
		return &configStep{
			h:        h,
			adapters: adapters,
			homeDir:  effectiveBase,
		}, nil

	case model.HarnessCommand:
		return &commandStep{
			h:         h,
			adapters:  adapters,
			homeDir:   effectiveBase,
			backupDir: filepath.Join(effectiveBase, ".jr-stack", "backups", "commands", h.ID),
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

func (s *verifyStep) ID() string { return "verify-hook" }
func (s *verifyStep) Run() error { return s.fn() }

// skillRunner is the interface required by skill.NewInstaller.
type skillRunner interface {
	Run(ctx context.Context, args []string) error
}
