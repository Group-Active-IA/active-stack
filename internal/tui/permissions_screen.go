package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// tierCapableAgents is the single source of truth for agents that can
// meaningfully differentiate the three permission tiers (claude and opencode).
// When a future agent gains a deny-list or equivalent, add it here and the
// ScreenPermissions will appear automatically.
//
// D8 (design.md): agents NOT in this set get ScreenPermissions skipped for them;
// the intent still carries the zero-value tier (normalized to TierBalanceado).
var tierCapableAgents = map[model.Agent]bool{
	model.AgentClaude:    true,
	model.AgentOpenCode:  true,
}

// anyTierCapable returns true when at least one agent in the set is tier-capable.
// It is a pure function of the agent set — deterministic and testable in both directions.
func anyTierCapable(agents []model.Agent) bool {
	for _, a := range agents {
		if tierCapableAgents[a] {
			return true
		}
	}
	return false
}

// tierOrder defines the display order of permission tiers in ScreenPermissions.
var tierOrder = []model.PermissionTier{
	model.TierEstricto,
	model.TierBalanceado,
	model.TierBypass,
}

// defaultTierCursor returns the cursor index that corresponds to TierBalanceado.
// Computed from tierOrder so a reorder cannot leave a stale hardcoded index.
func defaultTierCursor() int {
	for i, tier := range tierOrder {
		if tier == model.TierBalanceado {
			return i
		}
	}
	return 0
}

// bypassWarning is the text shown when the cursor is on the bypass tier.
const bypassWarning = "⚠ Bypass: autonomous mode — the security floor still applies (C-21)"

// enterPermissions transitions the model to ScreenPermissions, preselecting
// TierBalanceado. Called from ScreenMode and ScreenCustomPicker transitions.
func (m Model) enterPermissions() Model {
	m.Screen = ScreenPermissions
	m.Cursor = defaultTierCursor() // preselect balanceado
	return m
}

// viewPermissions renders the permission-tier radio list.
func (m Model) viewPermissions() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Permission tier") + "\n\n")
	for i, tier := range tierOrder {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		radio := "( )"
		if m.Selection.Tier == tier || (m.Selection.Tier == "" && tier == model.TierBalanceado) {
			radio = selectedStyle.Render("(•)")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, radio, tier))
	}

	// Show bypass warning when cursor is on the bypass tier.
	if m.Cursor < len(tierOrder) && tierOrder[m.Cursor] == model.TierBypass {
		sb.WriteString("\n" + errorStyle.Render(bypassWarning) + "\n")
	}

	sb.WriteString("\nSpace = pick  Enter = confirm  Esc = back\n")
	return sb.String()
}

// handlePermissionsKey handles keyboard input on ScreenPermissions.
func (m Model) handlePermissionsKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyDown:
		if m.Cursor < len(tierOrder)-1 {
			m.Cursor++
		}
	case tea.KeySpace:
		// Pick the tier under the cursor without advancing — mirrors the
		// Space toggle of the multi-select screens. Enter then confirms.
		m.Selection.Tier = tierOrder[m.Cursor]
	case tea.KeyEnter:
		// Commit the tier (cursor wins even if Space was not pressed) and advance.
		m.Selection.Tier = tierOrder[m.Cursor]
		return m.enterReview()
	case tea.KeyEsc:
		// Go back to wherever we came from.
		if m.Selection.Mode == model.ModeCustom {
			m.Screen = ScreenCustomPicker
		} else {
			m.Screen = ScreenMode
			m.Cursor = defaultModeCursor()
		}
	}
	return m, nil
}
