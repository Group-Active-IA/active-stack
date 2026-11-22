package autosync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// [EN] Fase 9 (R-20) T-F9-9: ENGRAM_CLOUD_AUTOSYNC=1 en el entorno del MCP de
// Engram, sin token ni servidor, y reversible.

type fakeAdapter struct {
	agent model.Agent
	path  string
}

func (f fakeAdapter) Agent() model.Agent               { return f.agent }
func (f fakeAdapter) MCPConfigPath(_, _ string) string { return f.path }
func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}
func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const claudeJSON = `{
  "numStartups": 4,
  "mcpServers": {
    "engram": {
      "type": "stdio",
      "command": "engram",
      "args": [
        "mcp"
      ]
    },
    "other": {
      "command": "x"
    }
  }
}
`

func TestClaudeSetAddsOnlyAutosyncToEngramEnv(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	write(t, p, claudeJSON)
	res, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{model.AgentClaude, p}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Agents) != 1 || !res.Agents[0].Changed {
		t.Fatalf("result = %+v", res)
	}
	got := read(t, p)
	if strings.Count(got, "ENGRAM_CLOUD_AUTOSYNC") != 1 || !strings.Contains(got, `"1"`) {
		t.Fatalf("missing autosync:\n%s", got)
	}
	for _, bad := range []string{"ENGRAM_CLOUD_TOKEN", "ENGRAM_CLOUD_SERVER", "ENGRAM_CLOUD_URL"} {
		if strings.Contains(got, bad) {
			t.Errorf("%s must never be written", bad)
		}
	}
	if !strings.Contains(got, `"numStartups": 4`) || !strings.Contains(got, `"other"`) {
		t.Errorf("foreign content lost:\n%s", got)
	}
}

func TestClaudeSetIsIdempotentAndRemoveRestoresBytes(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	write(t, p, claudeJSON)
	ad := []autosync.Adapter{fakeAdapter{model.AgentClaude, p}}
	if _, err := autosync.Set(home, ad); err != nil {
		t.Fatal(err)
	}
	once := read(t, p)
	res, err := autosync.Set(home, ad)
	if err != nil || res.Agents[0].Changed {
		t.Fatalf("second Set changed=%v err=%v", res.Agents[0].Changed, err)
	}
	if read(t, p) != once {
		t.Fatal("second Set rewrote the file")
	}
	res, err = autosync.Remove(home, ad)
	if err != nil || !res.Agents[0].Changed {
		t.Fatalf("Remove changed=%v err=%v", res.Agents[0].Changed, err)
	}
	if got := read(t, p); got != claudeJSON {
		t.Fatalf("Remove did not restore the original:\n%s", got)
	}
}

func TestClaudeNoEngramEntryIsSkippedUntouched(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	const src = "{\n  \"mcpServers\": {\n    \"other\": {\"command\": \"x\"}\n  }\n}\n"
	write(t, p, src)
	res, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{model.AgentClaude, p}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Agents[0].Skipped || res.Agents[0].Changed || read(t, p) != src {
		t.Fatalf("file must stay untouched: %+v", res.Agents[0])
	}
	// A missing file is not created.
	missing := filepath.Join(home, "nope.json")
	if _, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{model.AgentClaude, missing}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("Set created a missing file")
	}
}

func TestClaudeKeepsOtherEnvAndUserValue(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	const src = `{"mcpServers":{"engram":{"command":"engram","args":["mcp"],"env":{"FOO":"bar"}}}}`
	write(t, p, src)
	ad := []autosync.Adapter{fakeAdapter{model.AgentClaude, p}}
	if _, err := autosync.Set(home, ad); err != nil {
		t.Fatal(err)
	}
	if got := read(t, p); !strings.Contains(got, `"FOO":"bar"`) || !strings.Contains(got, "ENGRAM_CLOUD_AUTOSYNC") {
		t.Fatalf("got %s", got)
	}
	if _, err := autosync.Remove(home, ad); err != nil {
		t.Fatal(err)
	}
	if got := read(t, p); got != src {
		t.Fatalf("Remove must restore bytes, got %s", got)
	}
	// A value the person set to something else is theirs: neither changed nor removed.
	const mine = `{"mcpServers":{"engram":{"command":"engram","env":{"ENGRAM_CLOUD_AUTOSYNC":"0"}}}}`
	write(t, p, mine)
	res, err := autosync.Set(home, ad)
	if err != nil || res.Agents[0].Changed || len(res.Agents[0].Warnings) == 0 {
		t.Fatalf("Set over a user value: %+v %v", res.Agents[0], err)
	}
	if res, err := autosync.Remove(home, ad); err != nil || res.Agents[0].Changed {
		t.Fatalf("Remove over a user value: %+v %v", res.Agents[0], err)
	}
	if read(t, p) != mine {
		t.Fatal("user value touched")
	}
}

func TestInvalidJSONIsRefusedAndIntact(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	write(t, p, `{"mcpServers": {"engram": `)
	if _, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{model.AgentClaude, p}}); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if read(t, p) != `{"mcpServers": {"engram": ` {
		t.Fatal("invalid file modified")
	}
}

const codexTOML = `model = "gpt-5"

[mcp_servers.engram]
command = "engram"
args = ["mcp"]

[projects."x"]
trust = "trusted"
`

func TestCodexSetAndRemove(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".codex", "config.toml")
	write(t, p, codexTOML)
	ad := []autosync.Adapter{fakeAdapter{model.AgentCodex, p}}
	res, err := autosync.Set(home, ad)
	if err != nil || !res.Agents[0].Changed {
		t.Fatalf("Set: %+v %v", res, err)
	}
	got := read(t, p)
	if !strings.Contains(got, "[mcp_servers.engram.env]") || !strings.Contains(got, `ENGRAM_CLOUD_AUTOSYNC = "1"`) {
		t.Fatalf("missing env table:\n%s", got)
	}
	if !strings.Contains(got, `trust = "trusted"`) || !strings.Contains(got, `model = "gpt-5"`) {
		t.Fatalf("foreign content lost:\n%s", got)
	}
	if res, _ := autosync.Set(home, ad); res.Agents[0].Changed {
		t.Fatal("second Set changed the file")
	}
	if _, err := autosync.Remove(home, ad); err != nil {
		t.Fatal(err)
	}
	after := read(t, p)
	if strings.Contains(after, "AUTOSYNC") || strings.Contains(after, ".env]") {
		t.Fatalf("Remove left traces:\n%s", after)
	}
	if !strings.Contains(after, `command = "engram"`) || !strings.Contains(after, `trust = "trusted"`) {
		t.Fatalf("Remove damaged foreign content:\n%s", after)
	}
}

func TestCodexWithoutEngramEntryIsSkipped(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".codex", "config.toml")
	write(t, p, "model = \"gpt-5\"\n")
	res, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{model.AgentCodex, p}})
	if err != nil || !res.Agents[0].Skipped || read(t, p) != "model = \"gpt-5\"\n" {
		t.Fatalf("%+v %v", res, err)
	}
}

// OpenCode and Antigravity are out of scope: nothing is read or written.
func TestOtherAgentsAreSkipped(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, "opencode.json")
	const src = `{"mcp":{"engram":{"type":"local","command":["engram","mcp"],"enabled":true}}}`
	write(t, p, src)
	for _, a := range []model.Agent{model.AgentOpenCode, model.AgentAntigravity} {
		res, err := autosync.Set(home, []autosync.Adapter{fakeAdapter{a, p}})
		if err != nil || !res.Agents[0].Skipped {
			t.Fatalf("%s: %+v %v", a, res, err)
		}
		if _, err := autosync.Remove(home, []autosync.Adapter{fakeAdapter{a, p}}); err != nil {
			t.Fatal(err)
		}
	}
	if read(t, p) != src {
		t.Fatal("out-of-scope agent file modified")
	}
}

func TestStatusReportsEnabled(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, ".claude.json")
	write(t, p, claudeJSON)
	ad := []autosync.Adapter{fakeAdapter{model.AgentClaude, p}}
	st, err := autosync.Status(home, ad)
	if err != nil || len(st) != 1 || st[0].Enabled {
		t.Fatalf("before: %+v %v", st, err)
	}
	if _, err := autosync.Set(home, ad); err != nil {
		t.Fatal(err)
	}
	if st, _ = autosync.Status(home, ad); !st[0].Enabled {
		t.Fatal("not enabled after Set")
	}
}
