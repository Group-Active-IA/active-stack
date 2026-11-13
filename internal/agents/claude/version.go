package claude

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// VersionState is the outcome of checking the installed Claude Code version
// against the minimum that supports the layered delivery (claude-native-opsx D7).
type VersionState int

const (
	// VersionUnknown is the zero value: the binary exists but its version could
	// not be obtained or parsed.
	VersionUnknown VersionState = iota
	// VersionCompatible means the version is >= the minimum.
	VersionCompatible
	// VersionInsufficient means the version is parseable and < the minimum.
	VersionInsufficient
	// VersionAbsent means no `claude` binary was found.
	VersionAbsent
)

// VersionCheck is the result of DetectVersion.
type VersionCheck struct {
	State   VersionState
	Version string // parsed "major.minor.patch", empty when not parsed
	Min     string
	Detail  string // human-readable reason for non-compatible states
}

// VersionProbe abstracts process lookup and execution so tests inject values.
type VersionProbe struct {
	LookPath func(name string) (string, error)
	Output   func(path string, args ...string) ([]byte, error)
}

// MinVersion is the first Claude Code release where the layered delivery works:
// ${CLAUDE_SKILL_DIR} arrives in 2.1.69 and 2.1.110 fixes the
// disable-model-invocation failure when invoking /<skill> (official changelog).
// Verified with 2.1.318 on Linux, terminal, claude -p.
const MinVersion = "2.1.110"

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func parseVersion(s string) ([3]int, string, bool) {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return [3]int{}, "", false
	}
	var v [3]int
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(m[i+1])
		if err != nil {
			return [3]int{}, "", false
		}
		v[i] = n
	}
	return v, m[0], true
}

func versionLess(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// DetectVersion runs `claude --version` through the probe and classifies it
// against min.
func DetectVersion(min string, p VersionProbe) VersionCheck {
	res := VersionCheck{State: VersionUnknown, Min: min}
	minV, _, ok := parseVersion(min)
	if !ok {
		res.Detail = fmt.Sprintf("minimum version %q is not parseable", min)
		return res
	}
	path, err := p.LookPath("claude")
	if err != nil {
		res.State = VersionAbsent
		res.Detail = "claude binary not found in PATH"
		return res
	}
	out, err := p.Output(path, "--version")
	if err != nil {
		res.Detail = fmt.Sprintf("claude --version failed: %v", err)
		return res
	}
	got, text, ok := parseVersion(string(out))
	if !ok {
		res.Detail = fmt.Sprintf("claude --version output not parseable: %q", strings.TrimSpace(string(out)))
		return res
	}
	res.Version = text
	if versionLess(got, minV) {
		res.State = VersionInsufficient
		res.Detail = fmt.Sprintf("claude %s is older than the minimum %s", text, min)
		return res
	}
	res.State = VersionCompatible
	return res
}

// Detect checks the real `claude` binary against MinVersion.
func Detect() VersionCheck {
	return DetectVersion(MinVersion, VersionProbe{
		LookPath: exec.LookPath,
		Output: func(path string, args ...string) ([]byte, error) {
			return exec.Command(path, args...).Output()
		},
	})
}
