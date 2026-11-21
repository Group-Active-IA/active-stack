package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func installCodex(t *testing.T, home string) Result {
	t.Helper()
	res, err := Install(home, []Adapter{codexFake})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	return res
}

type tomlHandler struct {
	Type    string `toml:"type"`
	Command string `toml:"command"`
	Timeout int    `toml:"timeout"`
}

type tomlGroup struct {
	Matcher string        `toml:"matcher"`
	Hooks   []tomlHandler `toml:"hooks"`
}

func codexHooks(t *testing.T, content string) map[string][]tomlGroup {
	t.Helper()
	var doc struct {
		Hooks map[string][]tomlGroup `toml:"hooks"`
	}
	if err := toml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("config.toml is not valid TOML: %v\n%s", err, content)
	}
	return doc.Hooks
}

func ownTOMLGroups(groups []tomlGroup, sub string) []tomlGroup {
	var out []tomlGroup
	for _, g := range groups {
		for _, h := range g.Hooks {
			if h.Command == Command+" memory hook "+sub {
				out = append(out, g)
				break
			}
		}
	}
	return out
}

// pluginConfig imitates a config.toml with the Engram MCP and the hooks of the
// Engram plugin already there (same event, overlapping matcher).
const pluginConfig = `model = "gpt-5"
approval_policy = "on-request"

[mcp_servers.engram]
command = "engram"
args = ["mcp", "--tools=agent"]

[[hooks.PreToolUse]]
matcher = "^mcp__engram__mem_.*"

[[hooks.PreToolUse.hooks]]
type = "command"
command = 'engram hook codex-pre-tool-use'
timeout = 30

[sandbox_workspace_write]
network_access = false
`

// T-F6-1: creates the file with a marked block.
func TestCodex_CreatesConfigWhenAbsent(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	res := installCodex(t, home)
	if !res.Agents[0].Changed {
		t.Error("want Changed")
	}
	got := readFile(t, configPath(home))
	if !strings.Contains(got, EntryMarkerBegin) || !strings.Contains(got, EntryMarkerEnd) {
		t.Errorf("block markers missing:\n%s", got)
	}
	hooks := codexHooks(t, got)
	if len(ownTOMLGroups(hooks["SessionStart"], "session-start")) != 1 {
		t.Errorf("SessionStart = %+v", hooks["SessionStart"])
	}
	if len(ownTOMLGroups(hooks["PreToolUse"], "pre-read")) != 1 {
		t.Errorf("PreToolUse = %+v", hooks["PreToolUse"])
	}
}

// T-F6-1 / T-F6-2: the rest of the file keeps its bytes, foreign hooks stay.
func TestCodex_KeepsForeignContentByteForByte(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, configPath(home), pluginConfig)
	installCodex(t, home)

	got := readFile(t, configPath(home))
	if !strings.HasPrefix(got, pluginConfig) {
		t.Errorf("the original content must be an unchanged prefix:\n%s", got)
	}
	hooks := codexHooks(t, got)
	if len(hooks["PreToolUse"]) != 2 {
		t.Fatalf("PreToolUse has %d groups, want the plugin's plus ours", len(hooks["PreToolUse"]))
	}
	if hooks["PreToolUse"][0].Hooks[0].Command != "engram hook codex-pre-tool-use" {
		t.Error("the plugin's group must stay first")
	}
}

// T-F6-3: idempotent; a stale block is replaced, never duplicated.
func TestCodex_IdempotentAndReplacesStaleBlock(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, configPath(home), pluginConfig)
	installCodex(t, home)
	first := readFile(t, configPath(home))
	calls := len(snaps.calls)

	res := installCodex(t, home)
	if res.Agents[0].Changed || readFile(t, configPath(home)) != first {
		t.Error("second install must change nothing")
	}
	if len(snaps.calls) != calls {
		t.Error("an unchanged install must not take a snapshot")
	}

	stale := strings.Replace(first, "timeout = 10", "timeout = 99", 1)
	if stale == first {
		t.Fatal("test setup: nothing to make stale")
	}
	writeFile(t, configPath(home), stale)
	res = installCodex(t, home)
	if !res.Agents[0].Changed {
		t.Error("a stale block must be rewritten")
	}
	if got := readFile(t, configPath(home)); got != first {
		t.Errorf("stale block not replaced by the current one:\n%s", got)
	}
	if n := strings.Count(first, EntryMarkerBegin); n != 1 {
		t.Errorf("begin markers = %d", n)
	}
}

// T-F6-4.
func TestCodex_SnapshotBeforeWriteAndFailureLeavesFileIntact(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, configPath(home), pluginConfig)
	var seen string
	snaps.onCall = func([]string) { seen = readFile(t, configPath(home)) }
	installCodex(t, home)
	if len(snaps.calls) != 1 || snaps.calls[0][0] != configPath(home) || seen != pluginConfig {
		t.Fatalf("snapshot before write: calls=%v seen-original=%v", snaps.calls, seen == pluginConfig)
	}
	if !strings.Contains(snaps.dirs[0], "memory-hooks-codex") {
		t.Errorf("snapshot dir = %q", snaps.dirs[0])
	}

	home2 := t.TempDir()
	writeFile(t, configPath(home2), pluginConfig)
	snaps.err = os.ErrPermission
	snaps.onCall = nil
	if _, err := Install(home2, []Adapter{codexFake}); err == nil {
		t.Fatal("want an error when the backup fails")
	}
	if readFile(t, configPath(home2)) != pluginConfig {
		t.Error("file touched although the backup failed")
	}
}

// T-F6-5: invalid TOML, a static array that cannot be appended to, and broken
// markers are refused and left intact.
func TestCodex_RefusesInvalidOrUnexpectedFiles(t *testing.T) {
	cases := map[string]string{
		"invalid toml":        "model = \n[[hooks\n",
		"static array":        "[hooks]\nPreToolUse = [{ matcher = \"x\", hooks = [] }]\n",
		"begin without end":   "model = \"x\"\n\n" + EntryMarkerBegin + "\n[[hooks.SessionStart]]\n",
		"end without begin":   "model = \"x\"\n\n" + EntryMarkerEnd + "\n",
		"two blocks":          "\n" + EntryMarkerBegin + "\n" + EntryMarkerEnd + "\n" + EntryMarkerBegin + "\n" + EntryMarkerEnd + "\n",
		"end before begin":    EntryMarkerEnd + "\n" + EntryMarkerBegin + "\n",
		"hooks is a scalar":   "hooks = 3\n",
		"event is not arrays": "[hooks]\nSessionStart = 3\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			snaps := stubSnapshots(t)
			stubLookPath(t, true)
			home := t.TempDir()
			writeFile(t, configPath(home), content)
			_, err := Install(home, []Adapter{codexFake})
			if err == nil {
				t.Fatal("want an error")
			}
			if !strings.Contains(err.Error(), "config.toml") {
				t.Errorf("the error must name the file, got %v", err)
			}
			if readFile(t, configPath(home)) != content {
				t.Error("file changed")
			}
			if len(snaps.calls) != 0 {
				t.Error("no snapshot for a refused file")
			}
			if _, err := Remove(home, []Adapter{codexFake}); err == nil && name != "static array" && name != "hooks is a scalar" && name != "event is not arrays" {
				t.Error("Remove on a broken block must also refuse")
			}
		})
	}
}

// T-F6-6 + T-F6-12: Remove gives back the original bytes (files that end in a
// newline, LF or CRLF).
func TestCodex_RemoveRestoresOriginalBytes(t *testing.T) {
	cases := map[string]string{
		"plugin config":       pluginConfig,
		"plugin config crlf":  toCRLF(pluginConfig),
		"top level only":      "model = \"gpt-5\"\n",
		"ends with blank":     "model = \"gpt-5\"\n\n",
		"only a table":        "[mcp_servers.engram]\ncommand = \"engram\"\n",
		"empty file":          "",
		"comment only":        "# my notes\n",
		"hooks table of mine": "[[hooks.Stop]]\n\n[[hooks.Stop.hooks]]\ntype = \"command\"\ncommand = 'echo x'\n",
	}
	for name, original := range cases {
		t.Run(name, func(t *testing.T) {
			stubSnapshots(t)
			stubLookPath(t, true)
			home := t.TempDir()
			writeFile(t, configPath(home), original)
			installCodex(t, home)
			if readFile(t, configPath(home)) == original {
				t.Fatal("install changed nothing")
			}
			res, err := Remove(home, []Adapter{codexFake})
			if err != nil {
				t.Fatalf("Remove: %v", err)
			}
			if !res.Agents[0].Changed {
				t.Error("Remove must report a change")
			}
			if got := readFile(t, configPath(home)); got != original {
				t.Errorf("round trip differs.\nwant: %q\ngot:  %q", original, got)
			}
		})
	}
}

// A file with CRLF gets a CRLF block (no mixed line endings).
func TestCodex_BlockUsesTheFileLineEnding(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, configPath(home), toCRLF(pluginConfig))
	installCodex(t, home)
	got := readFile(t, configPath(home))
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Error("bare LF mixed into a CRLF file")
	}
}

func TestCodex_RemoveWithNothingInstalledIsANoOp(t *testing.T) {
	snaps := stubSnapshots(t)
	home := t.TempDir()
	if res, err := Remove(home, []Adapter{codexFake}); err != nil || res.Agents[0].Changed {
		t.Fatalf("absent: %+v %v", res, err)
	}
	writeFile(t, configPath(home), pluginConfig)
	if res, err := Remove(home, []Adapter{codexFake}); err != nil || res.Agents[0].Changed {
		t.Fatalf("nothing of ours: %+v %v", res, err)
	}
	if readFile(t, configPath(home)) != pluginConfig || len(snaps.calls) != 0 {
		t.Error("file changed or a snapshot was taken for a no-op")
	}
}

// T-F6-7: never declared in hooks.json too. If it is already there, config.toml
// is not written and the user is told.
func TestCodex_HooksJSONHasTheSameHook_NotDuplicated(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	hj := filepath.Join(home, ".codex", "hooks.json")
	body := `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"active-stack memory hook session-start"}]}]}}`
	writeFile(t, hj, body)
	writeFile(t, configPath(home), pluginConfig)

	res := installCodex(t, home)
	if res.Agents[0].Changed || !res.Agents[0].Skipped {
		t.Errorf("want skipped without changes, got %+v", res.Agents[0])
	}
	if readFile(t, configPath(home)) != pluginConfig || readFile(t, hj) != body {
		t.Error("neither file may be touched")
	}
	if !strings.Contains(strings.Join(res.Agents[0].Warnings, " "), "hooks.json") {
		t.Errorf("warnings = %v", res.Agents[0].Warnings)
	}
}

func TestCodex_NeverWritesHooksJSON(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	installCodex(t, home)
	if exists(filepath.Join(home, ".codex", "hooks.json")) {
		t.Error("hooks.json must not be created")
	}
}

// T-F6-8: same matcher rules as Claude, and no save hook.
func TestCodex_PreReadMatcherAndNoSaveHook(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	installCodex(t, home)
	got := readFile(t, configPath(home))
	g := ownTOMLGroups(codexHooks(t, got)["PreToolUse"], "pre-read")
	if len(g) != 1 {
		t.Fatal("no pre-read group")
	}
	mustMatch(t, g[0].Matcher, true,
		"mcp__engram__mem_search", "mcp__engram__mem_context",
		"mcp__plugin_engram_engram__mem_search", "mcp__plugin_engram_engram__mem_context")
	mustMatch(t, g[0].Matcher, false,
		"mcp__engram__mem_save", "mcp__engram__mem_session_summary", "mcp__engram__mem_capture_passive",
		"mcp__other__mem_search", "xmcp__engram__mem_search", "mcp__engram__mem_search_prompts", "Bash")
	if g := ownTOMLGroups(codexHooks(t, got)["SessionStart"], "session-start"); len(g) != 1 || g[0].Matcher != "" {
		t.Errorf("SessionStart group = %+v", g)
	}
	for _, banned := range []string{"pre-save", "mem_save", "http", "token", "ENGRAM_"} {
		if strings.Contains(got, banned) {
			t.Errorf("file contains %q", banned)
		}
	}
	for _, h := range append(codexHooks(t, got)["SessionStart"], codexHooks(t, got)["PreToolUse"]...) {
		for _, x := range h.Hooks {
			if strings.Contains(x.Command, Command) && x.Timeout != 10 {
				t.Errorf("timeout = %d", x.Timeout)
			}
		}
	}
}

// T-F6-11: Codex does not run the hook until the person trusts it.
func TestCodex_NoticeToReviewAndTrust(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	res := installCodex(t, home)
	n := strings.Join(res.Notices, " ")
	if !strings.Contains(n, "/hooks") || !strings.Contains(n, "Codex") {
		t.Errorf("notices = %v", res.Notices)
	}
	// still there when nothing changed: it is the person's pending step
	res = installCodex(t, home)
	if !strings.Contains(strings.Join(res.Notices, " "), "/hooks") {
		t.Errorf("notice must stay while the hook is installed: %v", res.Notices)
	}
	// a Claude-only install does not mention Codex
	res, err := Install(t.TempDir(), []Adapter{claudeFake})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(res.Notices, " "), "/hooks") {
		t.Errorf("Claude-only install must not mention /hooks: %v", res.Notices)
	}
}

func TestCodex_InstallsWithWarningWhenBinaryMissing(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, false)
	home := t.TempDir()
	res := installCodex(t, home)
	w := strings.Join(res.Agents[0].Warnings, " ")
	if !res.Agents[0].Changed || !strings.Contains(w, "PATH") {
		t.Errorf("result = %+v", res.Agents[0])
	}
}

func TestCodex_Status(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	if st, _ := Status(home, []Adapter{codexFake}); st[0].Installed || len(st[0].Missing) != 2 {
		t.Errorf("absent: %+v", st[0])
	}
	installCodex(t, home)
	if st, _ := Status(home, []Adapter{codexFake}); !st[0].Installed || !st[0].BinaryOnPath {
		t.Errorf("installed: %+v", st[0])
	}
	// the person removes the markers but leaves the hooks: not ours any more
	noMarkers := strings.ReplaceAll(strings.ReplaceAll(readFile(t, configPath(home)), EntryMarkerBegin, "# x"), EntryMarkerEnd, "# y")
	writeFile(t, configPath(home), noMarkers)
	if st, _ := Status(home, []Adapter{codexFake}); st[0].Installed {
		t.Errorf("without markers it is not our block: %+v", st[0])
	}
	// half of the markers is a broken block: it is reported, not guessed
	writeFile(t, configPath(home), strings.Replace(noMarkers, "# x", EntryMarkerBegin, 1))
	if _, err := Status(home, []Adapter{codexFake}); err == nil {
		t.Error("a begin marker without its end must be reported as an error")
	}
}
