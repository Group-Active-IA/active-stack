package headless_test

import (
	"os"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/install"
)

// TestMain pins the Claude Code version so no test here depends on the claude
// binary of the machine (PR #8 review, C5). It also disables the persistent
// Windows PATH write: an isolated LOCALAPPDATA does not isolate the HKCU
// registry, so self-install tests used to prepend temp dirs to the user PATH.
func TestMain(m *testing.M) {
	install.SetClaudeDetectFn(func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
	})
	install.SetAddToUserPathFn(func(string) error { return nil })
	os.Exit(m.Run())
}
