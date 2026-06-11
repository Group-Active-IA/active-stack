package headless_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
)

// TestParseInstallFlags_NoSelfInstall verifies that --no-self-install sets
// ParsedFlags.NoSelfInstall == true.
func TestParseInstallFlags_NoSelfInstall(t *testing.T) {
	result, err := headless.ParseInstallFlags([]string{
		"--mode", "lite",
		"--no-self-install",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NoSelfInstall {
		t.Error("expected NoSelfInstall == true when --no-self-install is passed")
	}
}

// TestParseInstallFlags_DefaultSelfInstallOn verifies that without the flag,
// NoSelfInstall defaults to false (self-install ON).
func TestParseInstallFlags_DefaultSelfInstallOn(t *testing.T) {
	result, err := headless.ParseInstallFlags([]string{"--mode", "lite"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NoSelfInstall {
		t.Error("expected NoSelfInstall == false by default (self-install ON)")
	}
}
