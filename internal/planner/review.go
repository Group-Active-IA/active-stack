package planner

import "github.com/Group-Active-IA/active-stack/internal/model"

// BuildReviewPayload constructs the data shown on the TUI confirmation screen
// before installation begins. It labels each harness as "selected" (user
// explicitly chose it) or "auto-dependency" (pulled in transitively).
func BuildReviewPayload(
	selected []model.Harness,
	resolved ResolvedPlan,
	agents []model.Agent,
	mode model.InstallMode,
) ReviewPayload {
	addedSet := make(map[string]struct{}, len(resolved.AddedIDs))
	for _, id := range resolved.AddedIDs {
		addedSet[id] = struct{}{}
	}

	actions := make([]HarnessAction, 0, len(resolved.OrderedIDs))
	for _, id := range resolved.OrderedIDs {
		action := "selected"
		if _, isAdded := addedSet[id]; isAdded {
			action = "auto-dependency"
		}
		actions = append(actions, HarnessAction{ID: id, Action: action})
	}

	return ReviewPayload{
		HarnessActions: actions,
		AddedIDs:       resolved.AddedIDs,
		Agents:         agents,
		Mode:           mode,
	}
}
