package hooks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type fakeAdapter struct {
	agent model.Agent
	path  string // relative to home; empty means no settings file
}

func (f fakeAdapter) Agent() model.Agent { return f.agent }
func (f fakeAdapter) SettingsPath(home string) string {
	if f.path == "" {
		return ""
	}
	return filepath.Join(home, filepath.FromSlash(f.path))
}

var (
	claudeFake = fakeAdapter{model.AgentClaude, ".claude/settings.json"}
	codexFake  = fakeAdapter{model.AgentCodex, ".codex/config.toml"}
)

// snapshots records every snapshot request and can fail on demand.
type snapshots struct {
	calls [][]string
	dirs  []string
	err   error
	// onCall runs at snapshot time (to observe the file before it is written).
	onCall func(paths []string)
}

func stubSnapshots(t *testing.T) *snapshots {
	t.Helper()
	s := &snapshots{}
	restore := SetSnapshotterCreate(func(dir string, paths []string) error {
		s.calls = append(s.calls, paths)
		s.dirs = append(s.dirs, dir)
		if s.onCall != nil {
			s.onCall(paths)
		}
		return s.err
	})
	t.Cleanup(restore)
	return s
}

func stubLookPath(t *testing.T, found bool) {
	t.Helper()
	restore := SetLookPath(func(string) (string, error) {
		if found {
			return "/usr/local/bin/active-stack", nil
		}
		return "", errors.New("not found")
	})
	t.Cleanup(restore)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func settingsPath(home string) string { return claudeFake.SettingsPath(home) }
func configPath(home string) string   { return codexFake.SettingsPath(home) }

// pluginSettings imitates the settings.json an Engram plugin leaves: hooks
// already declared for the same events, with a matcher that overlaps ours.
const pluginSettings = `{
  "model": "opus",
  "env": {
    "A": "1"
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "mcp__plugin_engram_engram__mem_.*",
        "hooks": [
          {
            "type": "command",
            "command": "engram hook claude-pre-tool-use"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "engram hook prompt"
          }
        ]
      }
    ]
  },
  "permissions": {
    "allow": [
      "Bash(git status)"
    ]
  }
}
`

// handler is the part of a hook handler the tests inspect.
type handler struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout int      `json:"timeout"`
	Shell   string   `json:"shell"`
}

type group struct {
	Matcher string    `json:"matcher"`
	Hooks   []handler `json:"hooks"`
}

func claudeHooks(t *testing.T, content string) map[string][]group {
	t.Helper()
	var doc struct {
		Hooks map[string][]group `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v\n%s", err, content)
	}
	return doc.Hooks
}

// ownGroups returns the groups of an event that hold one of our handlers.
func ownGroups(groups []group, sub string) []group {
	var out []group
	for _, g := range groups {
		for _, h := range g.Hooks {
			if h.Command == Command && strings.Join(h.Args, " ") == "memory hook "+sub {
				out = append(out, g)
				break
			}
		}
	}
	return out
}

func mustMatch(t *testing.T, matcher string, want bool, tools ...string) {
	t.Helper()
	re, err := regexp.Compile(matcher)
	if err != nil {
		t.Fatalf("matcher %q is not a valid regex: %v", matcher, err)
	}
	for _, tool := range tools {
		if got := re.MatchString(tool); got != want {
			t.Errorf("matcher %q on %q = %v, want %v", matcher, tool, got, want)
		}
	}
}

func toCRLF(s string) string { return strings.ReplaceAll(s, "\n", "\r\n") }
