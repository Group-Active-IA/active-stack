package headless

import (
	"bytes"
	"strings"
	"testing"
)

// 6.7: install --help carries the CLAUDE.md user-content warning, in both languages.
func TestInstallHelpWarnsAboutUserClaudeMD(t *testing.T) {
	for args, want := range map[string]string{
		"--help":           "whatever you write in your own CLAUDE.md reaches every session and subagent",
		"--lang=es --help": "lo que escribas en tu propio CLAUDE.md llega a todas las sesiones y subagentes",
	} {
		var buf bytes.Buffer
		old := installFlagOutput
		installFlagOutput = &buf
		_, _ = ParseInstallFlags(strings.Fields(args))
		installFlagOutput = old
		if !strings.Contains(buf.String(), want) {
			t.Errorf("%s: help missing %q:\n%s", args, want, buf.String())
		}
	}
}
