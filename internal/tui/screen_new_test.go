package tui

import "testing"

// TestNewScreenConstants_String verifies that each new Screen constant
// added by tui-menu-hub has the correct String() representation.
func TestNewScreenConstants_String(t *testing.T) {
	tests := []struct {
		screen Screen
		want   string
	}{
		{ScreenStarters, "starters"},
		{ScreenBackups, "backups"},
		{ScreenUninstallAgents, "uninstall-agents"},
		{ScreenUninstallMode, "uninstall-mode"},
		{ScreenUninstallStrategy, "uninstall-strategy"},
		{ScreenUninstallConfirm, "uninstall-confirm"},
		{ScreenUninstalling, "uninstalling-hub"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.screen.String()
			if got != tt.want {
				t.Errorf("Screen(%d).String() = %q, want %q", tt.screen, got, tt.want)
			}
		})
	}
}
