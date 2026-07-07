package headless

import (
	"encoding/json"
	"io"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// windowsUninstallStrategy describes one guided-uninstall strategy for the
// GUI's uninstall-options screen (design D6).
type windowsUninstallStrategy struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	Description      string `json:"description"`
	Default          bool   `json:"default"`
	RequiresManifest bool   `json:"requires_manifest"`
}

type windowsUninstallOptionsResponse struct {
	DetectedAgents []string                   `json:"detected_agents"`
	Modes          []windowsModeOption        `json:"modes"`
	Strategies     []windowsUninstallStrategy `json:"strategies"`
}

// RunWindowsUninstallOptions emits the option set consumed by the GUI's
// guided-uninstall screen (design D6): detected agents (via the shared
// detectAgents helper — same output as RunWindowsDetect), the install-mode
// options, and the two uninstall strategies. The targeted strategy is
// default:true/requires_manifest:false; restore is requires_manifest:true.
func RunWindowsUninstallOptions(homeDir string, w io.Writer) error {
	resp := windowsUninstallOptionsResponse{
		DetectedAgents: detectAgents(homeDir),
		Modes:          windowsModeOptions(),
		Strategies: []windowsUninstallStrategy{
			{
				ID:               "targeted",
				Label:            "Targeted",
				Description:      "Reverse each installed harness individually.",
				Default:          true,
				RequiresManifest: false,
			},
			{
				ID:               "restore",
				Label:            "Restore from backup",
				Description:      "Restore the full pre-install state from a backup manifest.",
				Default:          false,
				RequiresManifest: true,
			},
		},
	}
	return json.NewEncoder(w).Encode(resp)
}

// windowsModeOptions returns the three install-mode options in display order
// (lite/full/custom), shared by RunWindowsOptions and RunWindowsUninstallOptions
// so the two paths cannot drift (design D6).
func windowsModeOptions() []windowsModeOption {
	return []windowsModeOption{
		{ID: string(model.ModeLite), Label: "Quick", Description: "Fast setup to start working right away."},
		{ID: string(model.ModeFull), Label: "Complete", Description: "Full recommended setup with all key tools."},
		{ID: string(model.ModeCustom), Label: "Custom", Description: "Choose exactly what to install."},
	}
}
