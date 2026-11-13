package config

import (
	"fmt"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Install runs the config installer for the given harness and adapters.
// It composes the sdd-orchestrator block from h.Toggles and injects it into
// each adapter's instructions file, taking a backup first.
//
// Harnesses registered in sectionFragments are injected as their own section
// (see installSection); everything else is the sdd-orchestrator path.
//
// Returns an error if:
//   - h.Type is not model.HarnessConfig
//   - any toggle in h.Toggles is unknown
//   - backup or file write fails for any adapter
func Install(h model.Harness, adapters []AgentAdapter, homeDir string) (Result, error) {
	return InstallWith(h, adapters, homeDir, InstallOptions{})
}

// InstallOptions tunes InstallWith for the layered Claude delivery.
type InstallOptions struct {
	// Home is the real user home: backups and the migration record live under
	// <Home>/.active-stack, also for a Project target (C8). Defaults to the base
	// directory passed to InstallWith.
	Home string
	// Target is Machine (zero value) or Project.
	Target model.InstallTarget
	// ShadowSkillsDir maps an agent to its Machine skills dir, checked for a
	// same-ID skill that would take precedence over a Project one.
	ShadowSkillsDir map[model.Agent]string
	// DetectVersion reports the installed Claude Code version; defaults to
	// claude.Detect.
	DetectVersion func() claude.VersionCheck
	// Layers builds the layered composition; defaults to ComposeLayers.
	Layers func(toggles []string) (Layers, error)
}

// InstallWith is Install with options.
func InstallWith(h model.Harness, adapters []AgentAdapter, homeDir string, opts InstallOptions) (Result, error) {
	if h.Type != model.HarnessConfig {
		return Result{}, fmt.Errorf("config.Install: harness %q has type %q, want %q",
			h.ID, h.Type, model.HarnessConfig)
	}

	if fragment, ok := sectionFragments[h.ID]; ok {
		return installSection(h, fragment, adapters, homeDir)
	}

	var written []string
	var warnings []string
	var modes map[model.Agent]string
	allAlready := true

	for _, adapter := range adapters {
		snapshotDir := filepath.Join(homeDir, ".active-stack", "backups",
			fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))

		var (
			wr          InjectResult
			writtenPath string
		)

		switch adapter.ConfigDelivery() {
		case model.ConfigDeliveryManagedSkill:
			mode, changed, warns, err := installManaged(h, adapter, homeDir, opts)
			if err != nil {
				return Result{}, fmt.Errorf("config.Install: managed skill for agent %q: %w", adapter.Agent(), err)
			}
			if modes == nil {
				modes = map[model.Agent]string{}
			}
			modes[adapter.Agent()] = mode
			warnings = append(warnings, warns...)
			if changed {
				allAlready = false
				written = append(written, adapter.InstructionsPath(homeDir))
			}
			continue

		case model.ConfigDeliveryPrimaryAgent:
			// Register as a tab-able primary agent in the settings JSON, and
			// clean any orchestrator section out of the instructions file so it
			// no longer leaks into every agent.
			settingsPath := adapter.SettingsPath(homeDir)
			if settingsPath == "" {
				// Adapter opts out of primary-agent injection.
				continue
			}
			composed, err := Compose(h.Toggles, adapter.VariantKey())
			if err != nil {
				return Result{}, fmt.Errorf("config.Install: compose for agent %q: %w", adapter.Agent(), err)
			}
			r, err := InjectPrimaryAgent(h.ID, composed, adapter.InstructionsPath(homeDir), settingsPath, snapshotDir)
			if err != nil {
				return Result{}, fmt.Errorf("config.Install: primary-agent inject for agent %q: %w", adapter.Agent(), err)
			}
			wr, writtenPath = r, settingsPath

		default:
			// Inject the composed block into the agent's instructions file.
			path := adapter.InstructionsPath(homeDir)
			if path == "" {
				// Adapter explicitly opts out of injection.
				continue
			}
			composed, err := Compose(h.Toggles, adapter.VariantKey())
			if err != nil {
				return Result{}, fmt.Errorf("config.Install: compose for agent %q: %w", adapter.Agent(), err)
			}
			r, err := Inject(path, composed, snapshotDir)
			if err != nil {
				return Result{}, fmt.Errorf("config.Install: inject for agent %q: %w", adapter.Agent(), err)
			}
			wr, writtenPath = r, path
		}

		if wr.Changed {
			allAlready = false
			written = append(written, writtenPath)
		}
	}

	return Result{
		Files:      written,
		AllAlready: allAlready && len(written) == 0,
		Warnings:   warnings,
		Modes:      modes,
	}, nil
}

// installManaged delivers the orchestrator for an agent with the managed skill
// mode: layered (universal block + skill) when the Claude version allows it,
// the legacy single block otherwise. It returns the mode ("skill" or "legacy").
func installManaged(h model.Harness, adapter AgentAdapter, base string, opts InstallOptions) (string, bool, []string, error) {
	home := opts.Home
	if home == "" {
		home = base
	}
	mdPath := adapter.InstructionsPath(base)
	if mdPath == "" {
		return "", false, nil, nil
	}
	detect := opts.DetectVersion
	if detect == nil {
		detect = claude.Detect
	}
	check := detect()

	if check.State == claude.VersionInsufficient || check.State == claude.VersionUnknown {
		composed, err := Compose(h.Toggles, adapter.VariantKey())
		if err != nil {
			return "", false, nil, err
		}
		snap := filepath.Join(home, ".active-stack", "backups", fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))
		r, err := Inject(mdPath, composed, snap)
		if err != nil {
			return "", false, nil, err
		}
		warn := fmt.Sprintf("claude version %s is older than the minimum %s: the coordinator stays in %s (legacy mode)", check.Version, check.Min, mdPath)
		if check.State == claude.VersionUnknown {
			warn = fmt.Sprintf("could not determine the claude version (%s): the coordinator stays in %s for compatibility (legacy mode; minimum %s)", check.Detail, mdPath, check.Min)
		}
		return "legacy", r.Changed, append([]string{warn}, r.Warnings...), nil
	}

	layersFn := opts.Layers
	if layersFn == nil {
		layersFn = ComposeLayers
	}
	layers, err := layersFn(h.Toggles)
	if err != nil {
		return "", false, nil, err
	}
	in := LayeredInstall{
		HomeDir:          home,
		InstructionsPath: mdPath,
		SkillsDir:        adapter.SkillsDir(base),
		Skill:            layers.Skill,
		UniversalContent: layers.Universal,
	}
	if opts.Target == model.Project {
		in.ShadowSkillsDir = opts.ShadowSkillsDir[adapter.Agent()]
	}
	res, err := InstallLayered(in)
	if err != nil {
		return "", false, nil, err
	}
	warns := res.Warnings
	if check.State == claude.VersionAbsent {
		warns = append([]string{"claude binary not found in PATH: version not verified; the layered delivery (universal block and skill) was written"}, warns...)
	}
	if res.MigratedLegacy {
		if err := recordMigrationOnce(home, res.ManifestDir, mdPath, filepath.Join(in.SkillsDir, layers.Skill.ID)); err != nil {
			return "", false, nil, fmt.Errorf("record migration: %w", err)
		}
	}
	return "skill", res.SkillChanged || res.InstructionsChanged, warns, nil
}

// installSection injects a registered fragment into its own section
// (active-stack:<h.ID>) of each adapter's instructions file. It only touches
// agents the harness supports and never registers a primary agent: a
// section harness extends the orchestrator, it is not an agent itself.
func installSection(h model.Harness, fragment string, adapters []AgentAdapter, homeDir string) (Result, error) {
	content, err := loadFragment(fragment)
	if err != nil {
		return Result{}, fmt.Errorf("config.Install: %w", err)
	}

	var written []string
	allAlready := true

	for _, adapter := range adapters {
		if !h.SupportsAgent(adapter.Agent()) {
			continue
		}
		path := adapter.InstructionsPath(homeDir)
		if path == "" {
			// Adapter explicitly opts out of injection.
			continue
		}
		snapshotDir := filepath.Join(homeDir, ".active-stack", "backups",
			fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))

		r, err := InjectSection(path, h.ID, content, snapshotDir)
		if err != nil {
			return Result{}, fmt.Errorf("config.Install: inject %q for agent %q: %w", h.ID, adapter.Agent(), err)
		}
		if r.Changed {
			allAlready = false
			written = append(written, path)
		}
	}

	return Result{
		Files:      written,
		AllAlready: allAlready && len(written) == 0,
	}, nil
}
