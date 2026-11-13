package headless_test

import (
	"os"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/install"
)

// TestMain pins the Claude Code version so no test here depends on the claude
// binary of the machine (PR #8 review, C5).
func TestMain(m *testing.M) {
	install.SetClaudeDetectFn(func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
	})
	os.Exit(m.Run())
}
