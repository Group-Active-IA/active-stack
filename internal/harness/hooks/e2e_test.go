package hooks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
)

const realisticClaudeSettings = `{
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

const realisticCodexConfig = `model = "gpt-5"

[mcp_servers.engram]
command = "engram"
args = ["mcp", "--tools=agent"]

[[hooks.PreToolUse]]
matcher = "^mcp__engram__mem_.*"

[[hooks.PreToolUse.hooks]]
type = "command"
command = 'engram hook codex-pre-tool-use'
`

// T-F6-14: end to end with the REAL adapters and a throwaway HOME. What Install
// wrote at the adapters' own paths is what Status reads, foreign hooks survive,
// and Remove gives the files back byte for byte.
func TestEndToEnd_RealAdapters(t *testing.T) {
	restoreSnap := hooks.SetSnapshotterCreate(func(string, []string) error { return nil })
	defer restoreSnap()
	restorePath := hooks.SetLookPath(func(string) (string, error) { return "active-stack", nil })
	defer restorePath()

	home := t.TempDir()
	adapters := []hooks.Adapter{claude.NewAdapter(), codex.NewAdapter()}
	claudePath := claude.NewAdapter().SettingsPath(home)
	codexPath := codex.NewAdapter().SettingsPath(home)
	for p, c := range map[string]string{claudePath: realisticClaudeSettings, codexPath: realisticCodexConfig} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	res, err := hooks.Install(home, adapters)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Agents) != 2 || !res.Agents[0].Changed || !res.Agents[1].Changed {
		t.Fatalf("result = %+v", res)
	}
	if res.Agents[0].File != claudePath || res.Agents[1].File != codexPath {
		t.Errorf("files = %q, %q: they must come from the adapters", res.Agents[0].File, res.Agents[1].File)
	}

	st, err := hooks.Status(home, adapters)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range st {
		if !s.Installed || len(s.Missing) != 0 || !s.BinaryOnPath {
			t.Errorf("status %s = %+v", s.Agent, s)
		}
	}

	gotClaude, _ := os.ReadFile(claudePath)
	gotCodex, _ := os.ReadFile(codexPath)
	for _, keep := range []string{"engram hook claude-pre-tool-use", "engram hook prompt", "Bash(git status)"} {
		if !strings.Contains(string(gotClaude), keep) {
			t.Errorf("Claude lost %q", keep)
		}
	}
	if !strings.HasPrefix(string(gotCodex), realisticCodexConfig) {
		t.Error("Codex original content must stay as an unchanged prefix")
	}

	// installing again is a no-op
	if res2, err := hooks.Install(home, adapters); err != nil || res2.Agents[0].Changed || res2.Agents[1].Changed {
		t.Errorf("second install: %+v %v", res2, err)
	}

	if _, err := hooks.Remove(home, adapters); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	backClaude, _ := os.ReadFile(claudePath)
	backCodex, _ := os.ReadFile(codexPath)
	if string(backClaude) != realisticClaudeSettings {
		t.Errorf("Claude not restored byte for byte:\n%s", backClaude)
	}
	if string(backCodex) != realisticCodexConfig {
		t.Errorf("Codex not restored byte for byte:\n%s", backCodex)
	}
	if st, _ := hooks.Status(home, adapters); st[0].Installed || st[1].Installed {
		t.Errorf("status after Remove = %+v", st)
	}
}

// The real snapshotter is used by default: the backup left under the home of
// the person, restored with the real restore service of the backup package,
// gives the original settings.json back.
func TestEndToEnd_RealSnapshotRestoresTheOriginal(t *testing.T) {
	home := t.TempDir()
	claudePath := claude.NewAdapter().SettingsPath(home)
	if err := os.MkdirAll(filepath.Dir(claudePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, []byte(realisticClaudeSettings), 0o644); err != nil {
		t.Fatal(err)
	}
	restorePath := hooks.SetLookPath(func(string) (string, error) { return "active-stack", nil })
	defer restorePath()

	if _, err := hooks.Install(home, []hooks.Adapter{claude.NewAdapter()}); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(claudePath); string(got) == realisticClaudeSettings {
		t.Fatal("install changed nothing")
	}

	manifest, err := backup.ReadManifest(filepath.Join(home, ".active-stack", "backups", "memory-hooks-claude", backup.ManifestFilename))
	if err != nil {
		t.Fatalf("no readable backup manifest under ~/.active-stack/backups/memory-hooks-claude: %v", err)
	}
	if err := (backup.RestoreService{}).Restore(manifest); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got, _ := os.ReadFile(claudePath); string(got) != realisticClaudeSettings {
		t.Errorf("the backup did not give the original back: %q", got)
	}
}
