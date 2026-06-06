// Package tui provides the slim Bubbletea install-flow TUI for jr-stack.
// It is a thin driver: collect intent → call install.BuildPlan → run the plan
// via pipeline.NewOrchestrator with a ProgressFunc bridge.
// No orchestration, backup, or rollback logic lives here.
package tui

// Screen identifies which step of the install flow is currently displayed.
type Screen int

const (
	ScreenUnknown      Screen = iota
	ScreenWelcome             // entry point / hub
	ScreenDetection           // show OS/agent detection results
	ScreenAgents              // agent multi-select
	ScreenMode                // Lite / Full / Custom radio
	ScreenCustomPicker        // per-harness checkbox (Custom only)
	ScreenPermissions         // permission-tier radio: estricto / balanceado / bypass
	ScreenReview              // show resolved plan; BuildPlan error shown here
	ScreenInstalling          // live progress subscribed to ProgressFunc
	ScreenComplete            // success / failure summary

	// Hub child screens (tui-menu-hub).
	ScreenStarters          // starters list + install
	ScreenBackups           // backups list with restore/rename/delete
	ScreenUninstallAgents   // uninstall: agent selection
	ScreenUninstallMode     // uninstall: mode selection
	ScreenUninstallStrategy // uninstall: strategy selection
	ScreenUninstallConfirm  // uninstall: confirmation gate
	ScreenUninstalling      // uninstall: live progress
)

// String returns a display label for the screen (used in tests/logs).
func (s Screen) String() string {
	switch s {
	case ScreenWelcome:
		return "welcome"
	case ScreenDetection:
		return "detection"
	case ScreenAgents:
		return "agents"
	case ScreenMode:
		return "mode"
	case ScreenCustomPicker:
		return "custom-picker"
	case ScreenPermissions:
		return "permissions"
	case ScreenReview:
		return "review"
	case ScreenInstalling:
		return "installing"
	case ScreenComplete:
		return "complete"
	case ScreenStarters:
		return "starters"
	case ScreenBackups:
		return "backups"
	case ScreenUninstallAgents:
		return "uninstall-agents"
	case ScreenUninstallMode:
		return "uninstall-mode"
	case ScreenUninstallStrategy:
		return "uninstall-strategy"
	case ScreenUninstallConfirm:
		return "uninstall-confirm"
	case ScreenUninstalling:
		return "uninstalling-hub"
	default:
		return "unknown"
	}
}
