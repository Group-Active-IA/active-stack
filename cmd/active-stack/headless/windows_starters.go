package headless

import (
	"encoding/json"
	"io"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/install"
)

// windowsStarterEntry is one entry in the "windows starters list" response
// (design D2, windows-contract-hub-operations).
type windowsStarterEntry struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Includes        []string `json:"includes"`
	Harnesses       []string `json:"harnesses"`
	MCPCount        int      `json:"mcp_count"`
	LongDescription string   `json:"long_description,omitempty"`
}

type windowsStartersListResponse struct {
	Starters []windowsStarterEntry `json:"starters"`
}

// RunWindowsStartersList emits {"starters":[{id,name,description,includes,
// harnesses,mcp_count}]} for every starter in cat.AllStarters(), in catalog
// order (design D2). harnesses is the resolved id set from ResolveStarter;
// mcp_count is len(ResolveStarterMCPs). An empty catalog yields
// {"starters":[]} — never null.
func RunWindowsStartersList(cat StarterCatalog, lang i18n.Lang, w io.Writer) error {
	all := cat.AllStarters()
	resp := windowsStartersListResponse{Starters: make([]windowsStarterEntry, 0, len(all))}

	for _, s := range all {
		harnesses, err := cat.ResolveStarter(s.ID)
		if err != nil {
			return err
		}
		harnessIDs := make([]string, 0, len(harnesses))
		for _, h := range harnesses {
			harnessIDs = append(harnessIDs, h.ID)
		}

		mcps, err := cat.ResolveStarterMCPs(s.ID)
		if err != nil {
			return err
		}

		includes := s.Includes
		if includes == nil {
			includes = []string{}
		}

		resp.Starters = append(resp.Starters, windowsStarterEntry{
			ID:              s.ID,
			Name:            s.Name,
			Description:     s.Description.Localized(string(lang)),
			Includes:        includes,
			Harnesses:       harnessIDs,
			MCPCount:        len(mcps),
			LongDescription: s.LongDescription.Localized(string(lang)),
		})
	}

	return json.NewEncoder(w).Encode(resp)
}

// RunWindowsStartersInstall runs "windows starters install" — the same
// install event-stream as "windows install" (via the shared runWindowsPipeline,
// design D3), except the terminal event's type is "starter_finished". It
// builds its ParsedFlags via BuildStarterInstallParams (design D1) so it
// cannot drift from the CLI "starter add" path.
//
// buildPlanFn is forwarded to BuildStarterInstallParams verbatim (nil is
// valid: RunHeadless falls back to the default install.BuildPlan). The
// production dispatcher wires install.WithEmbeddedSkillsFS the same way it
// does for "windows install", keeping the assets package out of headless.
func RunWindowsStartersInstall(
	flags ParsedStarterAddFlags,
	cat StarterCatalog,
	reg install.Registry,
	buildPlanFn func(install.Catalog, install.Intent, install.Options) (install.Plan, error),
	w io.Writer,
) int {
	installCat, ok := cat.(install.Catalog)
	if !ok {
		_ = json.NewEncoder(w).Encode(windowsInstallEvent{
			Type:    "starter_finished",
			Phase:   "install",
			Success: false,
			Message: "internal error: catalog does not implement install.Catalog",
		})
		return 1
	}

	params, err := BuildStarterInstallParams(flags, cat, buildPlanFn)
	if err != nil {
		_ = json.NewEncoder(w).Encode(windowsInstallEvent{
			Type:    "starter_finished",
			Phase:   "install",
			Success: false,
			Message: err.Error(),
		})
		return 1
	}

	return runWindowsPipeline(params, installCat, reg, w, "starter_finished")
}
