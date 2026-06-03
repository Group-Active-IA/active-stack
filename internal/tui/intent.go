package tui

import (
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// Selection holds the user's choices collected across the install-flow screens.
type Selection struct {
	// Agents is the set of agents the user chose to target.
	Agents []model.Agent
	// Mode is the chosen install mode (Lite, Full, or Custom).
	Mode model.InstallMode
	// CustomHarnesses holds the harness IDs chosen in the Custom picker.
	// Populated only when Mode == ModeCustom.
	CustomHarnesses []string
	// Tier is the permission tier the user chose on ScreenPermissions.
	// Zero-value normalizes defensively to TierBalanceado (never TierBypass)
	// both in BuildIntent and defensively in the permissions installer.
	Tier model.PermissionTier
}

// BuildIntent converts the user's selections into an install.Intent.
// For Lite/Full, Custom is left nil (catalog.ForMode owns the harness set).
// For Custom, Custom is populated from CustomHarnesses.
// The Tier is normalized to TierBalanceado if zero (never TierBypass).
func (s Selection) BuildIntent() install.Intent {
	intent := install.Intent{
		Agents: s.Agents,
		Mode:   s.Mode,
		Tier:   s.Tier.Normalize(), // defensive: zero-value → balanceado
	}
	if s.Mode == model.ModeCustom {
		intent.Custom = append([]string(nil), s.CustomHarnesses...)
	}
	return intent
}

// availableAgents returns the agents that are both in the detected set AND
// have a registered adapter — the intersection the spec requires.
// It is an unexported alias for AvailableAgentsList (package-internal tests
// use the short name).
func availableAgents(detected []model.Agent, registered []model.Agent) []model.Agent {
	return AvailableAgentsList(detected, registered)
}

// AvailableAgentsList is the exported version of availableAgents, for use by
// cmd/jr-stack and other callers outside the package.
func AvailableAgentsList(detected []model.Agent, registered []model.Agent) []model.Agent {
	regSet := make(map[model.Agent]bool, len(registered))
	for _, a := range registered {
		regSet[a] = true
	}
	var out []model.Agent
	for _, a := range detected {
		if regSet[a] {
			out = append(out, a)
		}
	}
	return out
}

// filterHarnessesByAgents returns the subset of harnesses that support at
// least one of the given agents. Delegates to model.Harness.SupportsAgent.
func filterHarnessesByAgents(harnesses []model.Harness, agents []model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range harnesses {
		for _, a := range agents {
			if h.SupportsAgent(a) {
				out = append(out, h)
				break
			}
		}
	}
	return out
}
