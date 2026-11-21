package hooks

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func installClaude(t *testing.T, home string) Result {
	t.Helper()
	res, err := Install(home, []Adapter{claudeFake})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	return res
}

// T-F6-1: creates the file and the hooks key when absent.
func TestClaude_CreatesSettingsWhenAbsent(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	res := installClaude(t, home)

	if !res.Agents[0].Changed {
		t.Error("want Changed on a fresh install")
	}
	hooks := claudeHooks(t, readFile(t, settingsPath(home)))
	if len(ownGroups(hooks["SessionStart"], "session-start")) != 1 {
		t.Errorf("SessionStart = %+v", hooks["SessionStart"])
	}
	if len(ownGroups(hooks["PreToolUse"], "pre-read")) != 1 {
		t.Errorf("PreToolUse = %+v", hooks["PreToolUse"])
	}
}

// T-F6-1 / T-F6-13: the rest of the file is kept, in the same order and bytes.
func TestClaude_KeepsOtherKeysInOrderAndBytes(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	original := "{\n  \"model\": \"opus\",\n  \"env\": {\"A\": \"1\", \"B\": [1, 2]},\n  \"theme\": \"dark\"\n}\n"
	writeFile(t, settingsPath(home), original)
	installClaude(t, home)

	got := readFile(t, settingsPath(home))
	for _, frag := range []string{
		"  \"model\": \"opus\",\n",
		"  \"env\": {\"A\": \"1\", \"B\": [1, 2]},\n",
		"  \"theme\": \"dark\"",
	} {
		if !strings.Contains(got, frag) {
			t.Errorf("lost or reformatted %q in:\n%s", frag, got)
		}
	}
	if !(strings.Index(got, `"model"`) < strings.Index(got, `"env"`) &&
		strings.Index(got, `"env"`) < strings.Index(got, `"theme"`) &&
		strings.Index(got, `"theme"`) < strings.Index(got, `"hooks"`)) {
		t.Errorf("key order changed:\n%s", got)
	}
}

// T-F6-2: foreign hooks (the Engram plugin's, the user's) are untouched.
func TestClaude_KeepsForeignHooks(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, settingsPath(home), pluginSettings)
	installClaude(t, home)

	got := readFile(t, settingsPath(home))
	// every byte of the original foreign hooks is still there
	for _, frag := range []string{
		"\"command\": \"engram hook claude-pre-tool-use\"",
		"\"matcher\": \"mcp__plugin_engram_engram__mem_.*\"",
		"\"UserPromptSubmit\": [",
		"\"command\": \"engram hook prompt\"",
		"\"allow\": [\n      \"Bash(git status)\"",
	} {
		if !strings.Contains(got, frag) {
			t.Errorf("foreign content lost: %q", frag)
		}
	}
	hooks := claudeHooks(t, got)
	if len(hooks["PreToolUse"]) != 2 {
		t.Fatalf("PreToolUse has %d groups, want the plugin's plus ours", len(hooks["PreToolUse"]))
	}
	if hooks["PreToolUse"][0].Hooks[0].Command != "engram hook claude-pre-tool-use" {
		t.Error("the plugin's group must stay first and unchanged")
	}
}

// T-F6-3: idempotent, no duplicates, no write the second time.
func TestClaude_Idempotent(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, settingsPath(home), pluginSettings)
	installClaude(t, home)
	first := readFile(t, settingsPath(home))
	callsAfterFirst := len(snaps.calls)

	res := installClaude(t, home)
	if res.Agents[0].Changed {
		t.Error("second install must not report a change")
	}
	if second := readFile(t, settingsPath(home)); second != first {
		t.Errorf("second install changed the file:\n%s", second)
	}
	if len(snaps.calls) != callsAfterFirst {
		t.Error("an unchanged install must not take another snapshot")
	}
	hooks := claudeHooks(t, readFile(t, settingsPath(home)))
	if n := len(ownGroups(hooks["PreToolUse"], "pre-read")); n != 1 {
		t.Errorf("own pre-read groups = %d, want 1", n)
	}
	if n := len(ownGroups(hooks["SessionStart"], "session-start")); n != 1 {
		t.Errorf("own session-start groups = %d, want 1", n)
	}
}

// T-F6-4: snapshot first; if it fails the file is not touched.
func TestClaude_SnapshotBeforeWrite(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, settingsPath(home), pluginSettings)
	var seen string
	snaps.onCall = func([]string) { seen = readFile(t, settingsPath(home)) }
	installClaude(t, home)

	if len(snaps.calls) != 1 || len(snaps.calls[0]) != 1 || snaps.calls[0][0] != settingsPath(home) {
		t.Fatalf("snapshot calls = %v", snaps.calls)
	}
	if seen != pluginSettings {
		t.Error("the snapshot must run before the file is written")
	}
	if !strings.Contains(snaps.dirs[0], "memory-hooks-claude") {
		t.Errorf("snapshot dir = %q", snaps.dirs[0])
	}
}

func TestClaude_SnapshotFailureLeavesFileIntact(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	snaps.err = os.ErrPermission
	home := t.TempDir()
	writeFile(t, settingsPath(home), pluginSettings)

	if _, err := Install(home, []Adapter{claudeFake}); err == nil {
		t.Fatal("want an error when the backup fails")
	}
	if got := readFile(t, settingsPath(home)); got != pluginSettings {
		t.Error("the file was touched although the backup failed")
	}
	if _, err := Remove(home, []Adapter{claudeFake}); err != nil {
		// nothing of ours is installed: Remove is a no-op and needs no snapshot
		t.Errorf("Remove with nothing installed: %v", err)
	}
}

// A file that does not exist has nothing to back up.
func TestClaude_NoSnapshotWhenFileAbsent(t *testing.T) {
	snaps := stubSnapshots(t)
	stubLookPath(t, true)
	installClaude(t, t.TempDir())
	if len(snaps.calls) != 0 {
		t.Errorf("snapshot calls = %v, want none for a new file", snaps.calls)
	}
}

// T-F6-5: invalid or unexpected files are refused and left intact.
func TestClaude_RefusesInvalidOrUnexpectedFiles(t *testing.T) {
	cases := map[string]string{
		"comments":            "{\n  // my note\n  \"model\": \"opus\"\n}\n",
		"trailing comma":      "{\"model\": \"opus\",}\n",
		"truncated":           "{\"model\": ",
		"top level array":     "[1, 2]\n",
		"hooks is an array":   "{\"hooks\": []}\n",
		"event is an object":  "{\"hooks\": {\"PreToolUse\": {}}}\n",
		"event is a string":   "{\"hooks\": {\"SessionStart\": \"x\"}}\n",
		"not json at all":     "model = 'opus'\n",
		"utf8 bom":            "\xef\xbb\xbf{}\n",
		"two top level items": "{} {}\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			snaps := stubSnapshots(t)
			stubLookPath(t, true)
			home := t.TempDir()
			writeFile(t, settingsPath(home), content)
			_, err := Install(home, []Adapter{claudeFake})
			if err == nil {
				t.Fatal("want an error")
			}
			if !strings.Contains(err.Error(), "settings.json") {
				t.Errorf("the error must name the file, got %v", err)
			}
			if got := readFile(t, settingsPath(home)); got != content {
				t.Error("file changed")
			}
			if len(snaps.calls) != 0 {
				t.Error("no snapshot for a file that is refused")
			}
			if _, err := Remove(home, []Adapter{claudeFake}); err == nil {
				t.Error("Remove on an invalid file must also refuse")
			}
			if _, err := Status(home, []Adapter{claudeFake}); err == nil {
				t.Error("Status on an invalid file must say so")
			}
		})
	}
}

// T-F6-6 + T-F6-12: Remove takes only our entries and gives back the original
// bytes, in the formats a real settings.json comes in.
func TestClaude_RemoveRestoresOriginalBytes(t *testing.T) {
	cases := map[string]string{
		"empty object":            "{}\n",
		"empty object no newline": "{}",
		"compact one line":        "{\"model\":\"opus\"}",
		"two spaces":              "{\n  \"model\": \"opus\"\n}\n",
		"four spaces":             "{\n    \"model\": \"opus\",\n    \"env\": {\n        \"A\": \"1\"\n    }\n}\n",
		"tabs":                    "{\n\t\"model\": \"opus\"\n}\n",
		"crlf":                    toCRLF("{\n  \"model\": \"opus\"\n}\n"),
		"plugin hooks":            pluginSettings,
		"plugin hooks crlf":       toCRLF(pluginSettings),
		"hooks object only":       "{\n  \"hooks\": {\n    \"Stop\": [\n      {\n        \"hooks\": [\n          {\n            \"type\": \"command\",\n            \"command\": \"x\"\n          }\n        ]\n      }\n    ]\n  }\n}\n",
		"foreign session start":   "{\n  \"hooks\": {\n    \"SessionStart\": [\n      {\n        \"hooks\": [\n          {\n            \"type\": \"command\",\n            \"command\": \"echo hi\"\n          }\n        ]\n      }\n    ]\n  }\n}\n",
	}
	for name, original := range cases {
		t.Run(name, func(t *testing.T) {
			stubSnapshots(t)
			stubLookPath(t, true)
			home := t.TempDir()
			writeFile(t, settingsPath(home), original)
			installClaude(t, home)
			if got := readFile(t, settingsPath(home)); got == original {
				t.Fatal("install changed nothing")
			}
			res, err := Remove(home, []Adapter{claudeFake})
			if err != nil {
				t.Fatalf("Remove: %v", err)
			}
			if !res.Agents[0].Changed {
				t.Error("Remove must report a change")
			}
			if got := readFile(t, settingsPath(home)); got != original {
				t.Errorf("round trip differs.\nwant:\n%q\ngot:\n%q", original, got)
			}
		})
	}
}

// Remove empties what only we filled, and leaves alone what was already empty.
func TestClaude_RemoveCleansOnlyWhatItEmptied(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, settingsPath(home), "{\n  \"hooks\": {\n    \"Stop\": []\n  }\n}\n")
	installClaude(t, home)
	if _, err := Remove(home, []Adapter{claudeFake}); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, settingsPath(home))
	if !strings.Contains(got, "\"Stop\": []") {
		t.Errorf("a foreign empty array must stay:\n%s", got)
	}
	if strings.Contains(got, "SessionStart") || strings.Contains(got, "PreToolUse") {
		t.Errorf("our events must be gone:\n%s", got)
	}
}

func TestClaude_RemoveWithNothingInstalledIsANoOp(t *testing.T) {
	snaps := stubSnapshots(t)
	home := t.TempDir()
	if res, err := Remove(home, []Adapter{claudeFake}); err != nil || res.Agents[0].Changed {
		t.Fatalf("absent file: %+v %v", res, err)
	}
	writeFile(t, settingsPath(home), pluginSettings)
	if res, err := Remove(home, []Adapter{claudeFake}); err != nil || res.Agents[0].Changed {
		t.Fatalf("nothing of ours: %+v %v", res, err)
	}
	if got := readFile(t, settingsPath(home)); got != pluginSettings {
		t.Error("file changed")
	}
	if len(snaps.calls) != 0 {
		t.Error("no snapshot when there is nothing to remove")
	}
}

// A handler of ours next to the user's own handler in the same group: only ours goes.
func TestClaude_RemoveKeepsForeignHandlerInSharedGroup(t *testing.T) {
	stubSnapshots(t)
	home := t.TempDir()
	shared := `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "echo mine" },
          { "type": "command", "command": "active-stack", "args": ["memory", "hook", "session-start"], "timeout": 10 }
        ]
      }
    ]
  }
}
`
	writeFile(t, settingsPath(home), shared)
	if _, err := Remove(home, []Adapter{claudeFake}); err != nil {
		t.Fatal(err)
	}
	hooks := claudeHooks(t, readFile(t, settingsPath(home)))
	g := hooks["SessionStart"]
	if len(g) != 1 || len(g[0].Hooks) != 1 || g[0].Hooks[0].Command != "echo mine" {
		t.Errorf("SessionStart after Remove = %+v", g)
	}
}

// T-F6-8: the matcher covers both tool prefixes, only the two read tools, and
// is anchored. No save tool is covered.
func TestClaude_PreReadMatcher(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	installClaude(t, home)
	g := ownGroups(claudeHooks(t, readFile(t, settingsPath(home)))["PreToolUse"], "pre-read")
	if len(g) != 1 {
		t.Fatal("no pre-read group")
	}
	mustMatch(t, g[0].Matcher, true,
		"mcp__engram__mem_search", "mcp__engram__mem_context",
		"mcp__plugin_engram_engram__mem_search", "mcp__plugin_engram_engram__mem_context")
	mustMatch(t, g[0].Matcher, false,
		"mcp__engram__mem_save", "mcp__engram__mem_update", "mcp__engram__mem_session_summary",
		"mcp__engram__mem_capture_passive", "mcp__plugin_engram_engram__mem_save",
		"mcp__other__mem_search", "xmcp__engram__mem_search", "mcp__engram__mem_search_prompts",
		"mcp__engram__mem_context2", "Bash", "")
}

// Without shell (T-F6-10), with a timeout, and no save hook anywhere.
func TestClaude_HandlerShapeAndNoSaveHook(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	installClaude(t, home)
	got := readFile(t, settingsPath(home))
	hooks := claudeHooks(t, got)

	for ev, subs := range map[string]string{"SessionStart": "session-start", "PreToolUse": "pre-read"} {
		g := ownGroups(hooks[ev], subs)
		if len(g) != 1 || len(g[0].Hooks) != 1 {
			t.Fatalf("%s groups = %+v", ev, g)
		}
		h := g[0].Hooks[0]
		if h.Type != "command" || h.Command != "active-stack" || len(h.Args) != 3 || h.Timeout != 10 || h.Shell != "" {
			t.Errorf("%s handler = %+v", ev, h)
		}
	}
	if g := ownGroups(hooks["SessionStart"], "session-start"); g[0].Matcher != "" {
		t.Errorf("SessionStart must apply to every source, matcher = %q", g[0].Matcher)
	}
	for _, banned := range []string{"pre-save", "mem_save", "http", "token", "ENGRAM_"} {
		if strings.Contains(got, banned) {
			t.Errorf("file contains %q: no save hook, URL or secret may be written", banned)
		}
	}
}

// D-C: no active-stack in PATH: it is installed and warned about.
func TestClaude_InstallsWithWarningWhenBinaryMissing(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, false)
	home := t.TempDir()
	res := installClaude(t, home)
	if !res.Agents[0].Changed || !exists(settingsPath(home)) {
		t.Fatal("the hooks must be installed anyway")
	}
	w := strings.Join(res.Agents[0].Warnings, " ")
	if !strings.Contains(w, "PATH") || !strings.Contains(w, "active-stack") {
		t.Errorf("warnings = %v", res.Agents[0].Warnings)
	}
	st, err := Status(home, []Adapter{claudeFake})
	if err != nil || !st[0].Installed || st[0].BinaryOnPath {
		t.Errorf("status = %+v, %v", st, err)
	}
}

func TestClaude_StatusReportsWhatIsMissing(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	if st, _ := Status(home, []Adapter{claudeFake}); st[0].Installed || len(st[0].Missing) != 2 || !st[0].BinaryOnPath {
		t.Errorf("absent file: %+v", st[0])
	}
	installClaude(t, home)
	if st, _ := Status(home, []Adapter{claudeFake}); !st[0].Installed || len(st[0].Missing) != 0 {
		t.Errorf("installed: %+v", st[0])
	}
	// the user deletes one of our handlers by hand
	writeFile(t, settingsPath(home), strings.Replace(readFile(t, settingsPath(home)), "session-start", "other", 1))
	st, _ := Status(home, []Adapter{claudeFake})
	if st[0].Installed || len(st[0].Missing) != 1 || st[0].Missing[0] != "session-start" {
		t.Errorf("partial: %+v", st[0])
	}
}

// Existing file permissions and the parent directory are left as they are.
func TestClaude_KeepsFileAndDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits")
	}
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, settingsPath(home), pluginSettings)
	dir := home + "/.claude"
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(settingsPath(home), 0o600); err != nil {
		t.Fatal(err)
	}
	installClaude(t, home)
	if fi, _ := os.Stat(settingsPath(home)); fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v", fi.Mode().Perm())
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Errorf("directory mode = %v", fi.Mode().Perm())
	}
}

// Identity is command AND args: another active-stack handler of the user's is
// not ours, survives Install and Remove, and does not count as installed.
func TestClaude_ForeignActiveStackHandlerIsNotOurs(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	theirs := "{\n  \"hooks\": {\n    \"SessionStart\": [\n      {\n        \"hooks\": [\n          {\n            \"type\": \"command\",\n            \"command\": \"active-stack\",\n            \"args\": [\n              \"memory\",\n              \"hook\",\n              \"session-start\",\n              \"--debug\"\n            ]\n          }\n        ]\n      }\n    ]\n  }\n}\n"
	writeFile(t, settingsPath(home), theirs)
	if st, _ := Status(home, []Adapter{claudeFake}); st[0].Installed || len(st[0].Missing) != 2 {
		t.Fatalf("a look-alike handler must not count as installed: %+v", st[0])
	}
	installClaude(t, home)
	if len(claudeHooks(t, readFile(t, settingsPath(home)))["SessionStart"]) != 2 {
		t.Error("ours must be added next to the look-alike")
	}
	if _, err := Remove(home, []Adapter{claudeFake}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, settingsPath(home)); got != theirs {
		t.Errorf("the look-alike handler must survive byte for byte:\n%s", got)
	}
}
