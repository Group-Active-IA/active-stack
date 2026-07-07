package install

import "github.com/Group-Active-IA/active-stack/internal/model"

// CustomPickerHarnesses returns the harnesses a user may choose in Custom mode
// for the given agents: every catalog harness available in Custom mode,
// filtered to those supporting at least one requested agent (union, deduped by
// the catalog's own ordering). It answers "what can be picked", NOT "what
// will be installed" (that is SelectHarnesses). Read-only; it forces nothing
// and is the single source of truth for the Custom picker list shared by the
// Windows options contract (RunWindowsOptions) and the TUI Custom picker
// (design D5, windows-contract-tier-multiagent).
func CustomPickerHarnesses(cat Catalog, agents []model.Agent) []model.Harness {
	return filterByAgents(cat.ForMode(model.ModeCustom), agents)
}
