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
		path := adapter.InstructionsPath(homeDir)
		if path == "" {
			// Adapter explicitly opts out of injection.
			continue
		}

		// Compose the block for this adapter's variant.
		composed, err := Compose(h.Toggles, adapter.VariantKey())
		if err != nil {
			return Result{}, fmt.Errorf("config.Install: compose for agent %q: %w", adapter.Agent(), err)
		}

		snapshotDir := filepath.Join(homeDir, ".jr-stack", "backups",
			fmt.Sprintf("%s-%s", h.ID, adapter.Agent()))

		wr, err := Inject(path, composed, snapshotDir)
		if err != nil {
			return Result{}, fmt.Errorf("config.Install: inject for agent %q: %w", adapter.Agent(), err)
		}

		if wr.Changed {
			allAlready = false
			written = append(written, path)
		}
	}

	return Result{
		Files:      written,
		AllAlready: allAlready && len(written) == 0,
	}, nil
}
