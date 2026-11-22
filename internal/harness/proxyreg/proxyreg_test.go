package proxyreg

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

var claudeInputs = map[string]string{
	"compact": `{"mcpServers":{"engram":{"command":"engram","args":["mcp","--tools=agent"]}}}`,
	"pretty": "{\n  \"numStartups\": 3,\n  \"mcpServers\": {\n    \"other\": {\"command\": \"x\", \"args\": [\"a\"]},\n" +
		"    \"engram\": {\n      \"command\": \"engram\",\n      \"args\": [\n        \"mcp\",\n        \"--tools=agent\"\n      ],\n      \"env\": {\"K\": \"v\"}\n    }\n  },\n  \"projects\": {}\n}\n",
	"crlf":         "{\r\n  \"mcpServers\": {\r\n    \"engram\": {\r\n      \"command\": \"C:\\\\Program Files\\\\engram.exe\",\r\n      \"args\": [\"mcp\"]\r\n    }\r\n  }\r\n}\r\n",
	"tabs":         "{\n\t\"mcpServers\": {\n\t\t\"engram\": {\n\t\t\t\"command\": \"engram\",\n\t\t\t\"args\": [ \"mcp\" ]\n\t\t}\n\t}\n}",
	"empty args":   `{"mcpServers":{"engram":{"command":"engram","args":[ ]}}}`,
	"args first":   `{"mcpServers":{"engram":{"args":["mcp"],"type":"stdio","command":"engram"}}}`,
	"empty spaced": "{\n  \"mcpServers\": {\n    \"engram\": {\n      \"command\": \"engram\",\n      \"args\": [\n      ]\n    }\n  }\n}\n",
}

func claudeEntryOf(t *testing.T, src []byte) (string, []string) {
	t.Helper()
	var doc struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(src, &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, src)
	}
	e := doc.MCPServers["engram"]
	return e.Command, e.Args
}

func TestClaudeRoundTrip(t *testing.T) {
	for name, in := range claudeInputs {
		t.Run(name, func(t *testing.T) {
			origCmd, origArgs := claudeEntryOf(t, []byte(in))
			reg, changed, err := RegisterClaude([]byte(in), "engram")
			if err != nil || !changed {
				t.Fatalf("register: changed=%v err=%v", changed, err)
			}
			cmd, args := claudeEntryOf(t, reg)
			want := append([]string{"memory", "proxy", "--", origCmd}, origArgs...)
			if cmd != "active-stack" || strings.Join(args, "\x00") != strings.Join(want, "\x00") {
				t.Fatalf("entry = %q %q, want active-stack %q", cmd, args, want)
			}
			// idempotent
			again, changed, err := RegisterClaude(reg, "engram")
			if err != nil || changed || string(again) != string(reg) {
				t.Fatalf("second register changed the file: changed=%v err=%v", changed, err)
			}
			// same bytes on remove
			back, changed, err := UnregisterClaude(reg, "engram")
			if err != nil || !changed || string(back) != in {
				t.Fatalf("unregister: changed=%v err=%v\n got %q\nwant %q", changed, err, back, in)
			}
		})
	}
}

func TestClaudeOtherContentUntouched(t *testing.T) {
	in := claudeInputs["pretty"]
	reg, _, err := RegisterClaude([]byte(in), "engram")
	if err != nil {
		t.Fatal(err)
	}
	for _, keep := range []string{`"numStartups": 3`, `"other": {"command": "x", "args": ["a"]}`, `"env": {"K": "v"}`, `"projects": {}`} {
		if !strings.Contains(string(reg), keep) {
			t.Fatalf("lost %q:\n%s", keep, reg)
		}
	}
}

func TestClaudeNoEntryAndBadShapes(t *testing.T) {
	notFound := []string{`{}`, `{"mcpServers":{}}`, `{"mcpServers":{"other":{"command":"x","args":[]}}}`}
	for _, in := range notFound {
		if _, _, err := RegisterClaude([]byte(in), "engram"); !errors.Is(err, ErrEntryNotFound) {
			t.Fatalf("%s: err = %v", in, err)
		}
		out, changed, err := UnregisterClaude([]byte(in), "engram")
		if err != nil || changed || string(out) != in {
			t.Fatalf("unregister %s: %v %v", in, changed, err)
		}
	}
	for name, in := range map[string]string{
		"invalid json": `{"mcpServers":`,
		"http server":  `{"mcpServers":{"engram":{"type":"http","url":"https://x"}}}`,
		"no args":      `{"mcpServers":{"engram":{"command":"engram"}}}`,
		"servers list": `{"mcpServers":[]}`,
	} {
		out, changed, err := RegisterClaude([]byte(in), "engram")
		if err == nil || changed || out != nil {
			t.Fatalf("%s: err=%v changed=%v", name, err, changed)
		}
	}
}

var codexInputs = map[string]string{
	"simple":                    "model = \"x\"\n\n[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\", \"--tools=agent\"]\n\n[other]\nk = 1\n",
	"with env and other server": "[mcp_servers.other]\ncommand = \"o\"\n\n[mcp_servers.engram]\ncommand = \"/opt/engram\"\nargs = [\"mcp\"]\nstartup_timeout_sec = 20\n\n[mcp_servers.engram.env]\nK = \"v\"\n",
	"crlf":                      "[mcp_servers.engram]\r\ncommand = \"engram\"\r\nargs = [\"mcp\"]\r\n\r\n[x]\r\ny = 1\r\n",
	"multiline args":            "[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\n  \"mcp\",\n  \"--tools=agent\",\n]\nenabled = true\n",
	"no eol at eof":             "[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]",
	"no args":                   "[mcp_servers.engram]\ncommand = \"engram\"\n",
	"comments":                  "# my servers\n[mcp_servers.engram] # engram\n# keep\ncommand = \"engram\"\nargs = [\"mcp\"]\n",
	"quoted name":               "[mcp_servers.\"engram\"]\ncommand = \"engram\"\nargs = [\"mcp\"]\n",
}

func codexEntryOf(t *testing.T, src string) (string, []string, map[string]any) {
	t.Helper()
	var doc map[string]any
	if err := toml.Unmarshal([]byte(src), &doc); err != nil {
		t.Fatalf("invalid TOML: %v\n%s", err, src)
	}
	e := doc["mcp_servers"].(map[string]any)["engram"].(map[string]any)
	var args []string
	if l, ok := e["args"].([]any); ok {
		for _, a := range l {
			args = append(args, a.(string))
		}
	}
	rest := map[string]any{}
	for k, v := range e {
		if k != "command" && k != "args" {
			rest[k] = v
		}
	}
	return e["command"].(string), args, rest
}

func TestCodexRoundTrip(t *testing.T) {
	for name, in := range codexInputs {
		t.Run(name, func(t *testing.T) {
			origCmd, origArgs, origRest := codexEntryOf(t, in)
			reg, changed, err := RegisterCodex(in, "engram")
			if err != nil || !changed {
				t.Fatalf("register: changed=%v err=%v\n%s", changed, err, reg)
			}
			cmd, args, rest := codexEntryOf(t, reg)
			want := append([]string{"memory", "proxy", "--", origCmd}, origArgs...)
			if cmd != "active-stack" || strings.Join(args, "\x00") != strings.Join(want, "\x00") {
				t.Fatalf("entry = %q %q, want %q\n%s", cmd, args, want, reg)
			}
			if len(rest) != len(origRest) {
				t.Fatalf("other keys changed: %v vs %v", rest, origRest)
			}
			if strings.Count(reg, MarkerBegin) != 1 || strings.Count(reg, MarkerEnd) != 1 {
				t.Fatalf("markers:\n%s", reg)
			}
			again, changed, err := RegisterCodex(reg, "engram")
			if err != nil || changed || again != reg {
				t.Fatalf("second register changed the file: %v %v", changed, err)
			}
			back, changed, err := UnregisterCodex(reg, "engram")
			if err != nil || !changed || back != in {
				t.Fatalf("unregister: changed=%v err=%v\n got %q\nwant %q", changed, err, back, in)
			}
		})
	}
}

func TestCodexNoEntryAndInvalid(t *testing.T) {
	for _, in := range []string{"", "model = \"x\"\n", "[mcp_servers.other]\ncommand = \"o\"\n"} {
		if _, _, err := RegisterCodex(in, "engram"); !errors.Is(err, ErrEntryNotFound) {
			t.Fatalf("%q: err = %v", in, err)
		}
		if out, changed, err := UnregisterCodex(in, "engram"); err != nil || changed || out != in {
			t.Fatalf("unregister %q: %v %v", in, changed, err)
		}
	}
	for name, in := range map[string]string{
		"invalid":      "[mcp_servers.engram\ncommand =",
		"remote":       "[mcp_servers.engram]\nurl = \"https://x\"\n",
		"two begins":   MarkerBegin + "\n" + MarkerBegin + "\n" + MarkerEnd + "\n[mcp_servers.engram]\ncommand = \"e\"\n",
		"args not str": "[mcp_servers.engram]\ncommand = \"e\"\nargs = [1]\n",
	} {
		if out, changed, err := RegisterCodex(in, "engram"); err == nil || changed || out != "" {
			t.Fatalf("%s: err=%v changed=%v", name, err, changed)
		}
	}
}

// ---- file level ----

type fakeAdapter struct {
	agent model.Agent
	path  string
}

func (f fakeAdapter) Agent() model.Agent { return f.agent }
func (f fakeAdapter) MCPConfigPath(home, _ string) string {
	if f.path == "" {
		return ""
	}
	return filepath.Join(home, filepath.FromSlash(f.path))
}

var (
	claudeFake = fakeAdapter{model.AgentClaude, ".claude.json"}
	codexFake  = fakeAdapter{model.AgentCodex, ".codex/config.toml"}
	adapters   = []Adapter{claudeFake, codexFake, fakeAdapter{model.AgentOpenCode, ".config/opencode/opencode.json"}, fakeAdapter{model.AgentGemini, ""}}
)

func writeFile(t *testing.T, home, rel, content string) string {
	t.Helper()
	p := filepath.Join(home, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func stubSnapshots(t *testing.T, err error) *[][]string {
	t.Helper()
	var calls [][]string
	restore := SetSnapshotterCreate(func(_ string, paths []string) error {
		calls = append(calls, paths)
		return err
	})
	t.Cleanup(restore)
	return &calls
}

func TestRegisterAndUnregisterFiles(t *testing.T) {
	home := t.TempDir()
	calls := stubSnapshots(t, nil)
	cj := writeFile(t, home, ".claude.json", claudeInputs["pretty"])
	ct := writeFile(t, home, ".codex/config.toml", codexInputs["simple"])
	oc := writeFile(t, home, ".config/opencode/opencode.json", `{"mcp":{"engram":{"command":["engram","mcp"]}}}`)

	res, err := Register(home, adapters)
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].Changed || !res[1].Changed || !res[2].Skipped || !res[3].Skipped {
		t.Fatalf("results = %+v", res)
	}
	if len(*calls) != 2 {
		t.Fatalf("snapshots = %v", *calls)
	}
	if b, _ := os.ReadFile(oc); string(b) != `{"mcp":{"engram":{"command":["engram","mcp"]}}}` {
		t.Fatal("OpenCode file was touched")
	}
	// second run: no change, no snapshot
	res, err = Register(home, adapters)
	if err != nil || res[0].Changed || res[1].Changed || len(*calls) != 2 {
		t.Fatalf("not idempotent: %+v %v snapshots=%d", res, err, len(*calls))
	}
	res, err = Unregister(home, adapters)
	if err != nil || !res[0].Changed || !res[1].Changed {
		t.Fatalf("unregister: %+v %v", res, err)
	}
	if b, _ := os.ReadFile(cj); string(b) != claudeInputs["pretty"] {
		t.Fatalf("claude file not restored:\n%s", b)
	}
	if b, _ := os.ReadFile(ct); string(b) != codexInputs["simple"] {
		t.Fatalf("codex file not restored:\n%s", b)
	}
}

func TestSnapshotBeforeWriteAndFailureLeavesFile(t *testing.T) {
	home := t.TempDir()
	stubSnapshots(t, errors.New("disk full"))
	cj := writeFile(t, home, ".claude.json", claudeInputs["compact"])
	_, err := Register(home, []Adapter{claudeFake})
	if err == nil || !strings.Contains(err.Error(), "backup failed") {
		t.Fatalf("err = %v", err)
	}
	if b, _ := os.ReadFile(cj); string(b) != claudeInputs["compact"] {
		t.Fatal("file touched although the backup failed")
	}
}

func TestInvalidFileLeftIntactAndMissingFilesSkipped(t *testing.T) {
	home := t.TempDir()
	calls := stubSnapshots(t, nil)
	bad := writeFile(t, home, ".claude.json", `{"mcpServers": `)
	if _, err := Register(home, []Adapter{claudeFake}); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if b, _ := os.ReadFile(bad); string(b) != `{"mcpServers": ` || len(*calls) != 0 {
		t.Fatal("invalid file touched or snapshotted")
	}
	res, err := Register(t.TempDir(), []Adapter{codexFake})
	if err != nil || !res[0].Skipped {
		t.Fatalf("missing file: %+v %v", res, err)
	}
}
