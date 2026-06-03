// Package tui provides the slim Bubbletea install-flow TUI for jr-stack.
// It is a thin driver: collect intent → call install.BuildPlan → run the plan
// via pipeline.NewOrchestrator with a ProgressFunc bridge.
// No orchestration, backup, or rollback logic lives here.
package tui

// Screen identifies which step of the install flow is currently displayed.
type Screen int

const (
	ScreenUnknown      Screen = iota
	ScreenWelcome             // entry point
	ScreenDetection           // show OS/agent detection results
	ScreenAgents              // agent multi-select
	ScreenMode                // Lite / Full / Custom radio
	ScreenCustomPicker        // per-harness checkbox (Custom only)
	ScreenPermissions         // permission-tier radio: estricto / balanceado / bypass
	ScreenReview              // show resolved plan; BuildPlan error shown here
	ScreenInstalling          // live progress subscribed to ProgressFunc
	ScreenComplete            // success / failure summary
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
	default:
		return "unknown"
	}
}
