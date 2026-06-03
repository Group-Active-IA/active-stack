package tui

import "testing"

// TestScreenPermissionsString verifies the new ScreenPermissions has the
// correct string representation (spec: tui-install 5.1).
func TestScreenPermissionsString(t *testing.T) {
	got := ScreenPermissions.String()
	want := "permissions"
	if got != want {
		t.Errorf("ScreenPermissions.String() = %q, want %q", got, want)
	}
}
