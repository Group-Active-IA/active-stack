package config

import (
	"fmt"
	"path/filepath"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Install runs the config installer for the given harness and adapters.
// It composes the sdd-orchestrator block from h.Toggles and injects it into
// each adapter's instructions file, taking a backup first.
//
// Returns an error if:
//   - h.Type is not model.HarnessConfig
//   - any toggle in h.Toggles is unknown
//   - backup or file write fails for any adapter
func Install(h model.Harness, adapters []AgentAdapter, homeDir string) (Result, error) {
	if h.Type != model.HarnessConfig {
		return Result{}, fmt.Errorf("config.Install: harness %q has type %q, want %q",
			h.ID, h.Type, model.HarnessConfig)
	}

	var written []string
	allAlready := true

	for _, adapter := range adapters {
		snapshotDir := filepath.Join(homeDir, ".jr-stack", "backups",
			fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))

		var (
			wr          InjectResult
			writtenPath string
		)

		switch adapter.ConfigDelivery() {
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
	}, nil
}
