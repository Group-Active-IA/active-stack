package claude_test

import (
	"errors"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
)

func probe(out string, runErr error) claude.VersionProbe {
	return claude.VersionProbe{
		LookPath: func(string) (string, error) { return "/usr/bin/claude", nil },
		Output:   func(string, ...string) ([]byte, error) { return []byte(out), runErr },
	}
}

func TestDetectVersion_States(t *testing.T) {
	absent := claude.VersionProbe{
		LookPath: func(string) (string, error) { return "", errors.New("not found") },
		Output: func(string, ...string) ([]byte, error) {
			t.Fatal("must not run when binary is absent")
			return nil, nil
		},
	}
	tests := []struct {
		name    string
		min     string
		probe   claude.VersionProbe
		want    claude.VersionState
		version string
	}{
		{"valid above minimum", "2.1.0", probe("2.1.50 (Claude Code)\n", nil), claude.VersionCompatible, "2.1.50"},
		{"equal to minimum", "2.1.0", probe("2.1.0 (Claude Code)", nil), claude.VersionCompatible, "2.1.0"},
		{"below minimum", "2.1.0", probe("2.0.99 (Claude Code)", nil), claude.VersionInsufficient, "2.0.99"},
		{"numeric not lexical compare", "1.9.0", probe("1.10.0 (Claude Code)", nil), claude.VersionCompatible, "1.10.0"},
		{"lower major numeric", "10.0.0", probe("9.9.9", nil), claude.VersionInsufficient, "9.9.9"},
		{"unparseable output", "2.1.0", probe("hello world", nil), claude.VersionUnknown, ""},
		{"run error", "2.1.0", probe("", errors.New("boom")), claude.VersionUnknown, ""},
		{"binary absent", "2.1.0", absent, claude.VersionAbsent, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := claude.DetectVersion(tc.min, tc.probe)
			if got.State != tc.want {
				t.Errorf("State = %v, want %v (detail %q)", got.State, tc.want, got.Detail)
			}
			if got.Version != tc.version {
				t.Errorf("Version = %q, want %q", got.Version, tc.version)
			}
		})
	}
}

func TestDetectVersion_NonCompatibleStatesExplainThemselves(t *testing.T) {
	got := claude.DetectVersion("2.1.0", probe("2.0.1", nil))
	if got.Detail == "" || got.Min != "2.1.0" {
		t.Errorf("insufficient result needs Detail and Min, got %+v", got)
	}
}

// MinVersion is 2.1.110: ${CLAUDE_SKILL_DIR} arrives in 2.1.69 and 2.1.110 fixes
// disable-model-invocation when invoking /<skill> (official changelog).
func TestMinVersionIsFirstFixedRelease(t *testing.T) {
	if claude.MinVersion != "2.1.110" {
		t.Errorf("MinVersion = %q, want 2.1.110", claude.MinVersion)
	}
	if got := claude.DetectVersion(claude.MinVersion, probe("2.1.109 (Claude Code)", nil)); got.State != claude.VersionInsufficient {
		t.Errorf("2.1.109 must be insufficient, got %v", got.State)
	}
}
