package proxyreg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

type statusAdapter struct {
	agent model.Agent
	path  string
}

func (a statusAdapter) Agent() model.Agent               { return a.agent }
func (a statusAdapter) MCPConfigPath(_, _ string) string { return a.path }

func TestStatusClaudeAndCodex(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	claudeRaw := claudeInputs["pretty"]
	reg, _, err := RegisterClaude([]byte(claudeRaw), ServerName)
	if err != nil {
		t.Fatal(err)
	}
	codexRaw := "[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]\n"
	codexReg, _, err := RegisterCodex(codexRaw, ServerName)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name         string
		a            statusAdapter
		hasEntry, ok bool
	}{
		{"claude plain", statusAdapter{model.AgentClaude, write("c1.json", claudeRaw)}, true, false},
		{"claude proxied", statusAdapter{model.AgentClaude, write("c2.json", string(reg))}, true, true},
		{"claude no entry", statusAdapter{model.AgentClaude, write("c3.json", `{"mcpServers":{}}`)}, false, false},
		{"claude no file", statusAdapter{model.AgentClaude, filepath.Join(dir, "none.json")}, false, false},
		{"codex plain", statusAdapter{model.AgentCodex, write("x1.toml", codexRaw)}, true, false},
		{"codex proxied", statusAdapter{model.AgentCodex, write("x2.toml", codexReg)}, true, true},
		{"codex no entry", statusAdapter{model.AgentCodex, write("x3.toml", "model = \"x\"\n")}, false, false},
	}
	for _, c := range cases {
		before, _ := os.ReadFile(c.a.path)
		got, err := Status(dir, []Adapter{c.a})
		if err != nil || len(got) != 1 {
			t.Fatalf("%s: %v %v", c.name, got, err)
		}
		if got[0].HasEntry != c.hasEntry || got[0].Registered != c.ok {
			t.Errorf("%s: %+v", c.name, got[0])
		}
		after, _ := os.ReadFile(c.a.path)
		if string(before) != string(after) {
			t.Errorf("%s: Status wrote the file", c.name)
		}
	}
	if got, _ := Status(dir, []Adapter{statusAdapter{model.AgentOpenCode, "x"}}); len(got) != 0 {
		t.Errorf("opencode not skipped: %v", got)
	}
}
