package headless

import (
	"encoding/json"
	"io"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
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
	LongDescription  string `json:"long_description,omitempty"`
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
func RunWindowsUninstallOptions(homeDir string, lang i18n.Lang, w io.Writer) error {
	resp := windowsUninstallOptionsResponse{
		DetectedAgents: detectAgents(homeDir),
		Modes:          windowsModeOptions(lang),
		Strategies: []windowsUninstallStrategy{
			{
				ID:               "targeted",
				Label:            i18n.T(lang, "strategy.targeted.label"),
				Description:      i18n.T(lang, "strategy.targeted.desc"),
				LongDescription:  i18n.T(lang, "strategy.targeted.long"),
				Default:          true,
				RequiresManifest: false,
			},
			{
				ID:               "restore",
				Label:            i18n.T(lang, "strategy.restore.label"),
				Description:      i18n.T(lang, "strategy.restore.desc"),
				LongDescription:  i18n.T(lang, "strategy.restore.long"),
				Default:          false,
				RequiresManifest: true,
			},
		},
	}
	return json.NewEncoder(w).Encode(resp)
}

// windowsModeOptions returns the three install-mode options in display order
// (lite/full/custom), shared by RunWindowsOptions and RunWindowsUninstallOptions
// so the two paths cannot drift (design D6). Labels/descriptions/long
// descriptions are sourced from the i18n tables (i18n-engine-locales D3).
func windowsModeOptions(lang i18n.Lang) []windowsModeOption {
	return []windowsModeOption{
		{ID: string(model.ModeLite), Label: i18n.T(lang, "mode.lite.label"), Description: i18n.T(lang, "mode.lite.desc"), LongDescription: i18n.T(lang, "mode.lite.long")},
		{ID: string(model.ModeFull), Label: i18n.T(lang, "mode.full.label"), Description: i18n.T(lang, "mode.full.desc"), LongDescription: i18n.T(lang, "mode.full.long")},
		{ID: string(model.ModeCustom), Label: i18n.T(lang, "mode.custom.label"), Description: i18n.T(lang, "mode.custom.desc"), LongDescription: i18n.T(lang, "mode.custom.long")},
	}
}
