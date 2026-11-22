package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/memory"
)

// runHook runs "active-stack memory hook <name>" with event on stdin.
func runHook(t *testing.T, name, event string) (int, string, string) {
	t.Helper()
	old := memoryStdin
	memoryStdin = strings.NewReader(event)
	defer func() { memoryStdin = old }()
	var stdout, stderr bytes.Buffer
	code := runMemoryDispatch([]string{"hook", name}, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func hookEvent(t *testing.T, fields map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func sessionEvent(t *testing.T, dir string) string {
	return hookEvent(t, map[string]any{"hook_event_name": "SessionStart", "source": "startup", "session_id": "s-1", "cwd": dir})
}

func readEvent(t *testing.T, dir string) string {
	return hookEvent(t, map[string]any{"hook_event_name": "PreToolUse", "session_id": "s-1", "cwd": dir,
		"tool_name": "mcp__engram__mem_search", "tool_input": map[string]any{"query": "login"}})
}

// additionalContext decodes the hook output and checks its event name.
func additionalContext(t *testing.T, out, event string) string {
	t.Helper()
	var got struct {
		HookSpecificOutput map[string]string `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not JSON: %q", out)
	}
	h := got.HookSpecificOutput
	if len(h) != 2 || h["hookEventName"] != event {
		t.Fatalf("hookSpecificOutput = %v; want only hookEventName=%s and additionalContext", h, event)
	}
	return h["additionalContext"]
}

func switchBranch(t *testing.T, dir, branch string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", dir, "switch", "-q", branch).CombinedOutput(); err != nil {
		t.Fatalf("git switch %s: %v\n%s", branch, err, out)
	}
}

// T-F5-11: session-start adds the team-memory line only with the marker.
func TestHookSessionStart(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	code, out, errOut := runHook(t, "session-start", sessionEvent(t, dir))
	if code != 0 || errOut != "" {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
	if got := additionalContext(t, out, "SessionStart"); got != "Team memory: cloud-branch. Current branch: feature/x (slug: feature-x-217d2b)." {
		t.Fatalf("additionalContext = %q", got)
	}
	switchBranch(t, dir, "main")
	_, out, _ = runHook(t, "session-start", sessionEvent(t, dir))
	if got := additionalContext(t, out, "SessionStart"); got != "Team memory: cloud-branch. Current branch: main (untagged)." {
		t.Fatalf("on main: additionalContext = %q", got)
	}
}

// T-F5-10: pre-read adds the current branch only with the marker and off the primary branch.
func TestHookPreRead(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	code, out, errOut := runHook(t, "pre-read", readEvent(t, dir))
	if code != 0 || errOut != "" {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
	if got := additionalContext(t, out, "PreToolUse"); got != "Current branch: feature/x (slug: feature-x-217d2b)." {
		t.Fatalf("additionalContext = %q", got)
	}
	switchBranch(t, dir, "main")
	if code, out, _ := runHook(t, "pre-read", readEvent(t, dir)); code != 0 || out != "" {
		t.Fatalf("on main: exit %d stdout %q; want no output", code, out)
	}
}

// T-F5-1: without the marker neither hook prints anything.
func TestHooksWithoutMarkerPrintNothing(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	if err := os.Remove(filepath.Join(dir, ".engram", "active-stack.json")); err != nil {
		t.Fatal(err)
	}
	for name, event := range map[string]string{"session-start": sessionEvent(t, dir), "pre-read": readEvent(t, dir)} {
		if code, out, errOut := runHook(t, name, event); code != 0 || out != "" || errOut != "" {
			t.Errorf("%s: exit %d stdout %q stderr %q; want nothing", name, code, out, errOut)
		}
	}
}

// T-F5-9: an event from a sub-agent (agent_id) is handled the same way.
func TestHookSubagentEventIsTheSame(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	_, plain, _ := runHook(t, "pre-read", readEvent(t, dir))
	_, sub, _ := runHook(t, "pre-read", hookEvent(t, map[string]any{"hook_event_name": "PreToolUse", "cwd": dir,
		"agent_id": "agent-7", "agent_type": "Explore", "tool_name": "mcp__engram__mem_search", "tool_input": map[string]any{}}))
	if plain == "" || sub != plain {
		t.Fatalf("sub-agent output %q; want %q", sub, plain)
	}
}

// T-F5-8: bad input, git errors, a timeout or a wrong hook never block: exit 0,
// no stdout, a message on stderr.
func TestHooksNeverBlock(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	cases := map[string]struct{ hook, event string }{
		"malformed JSON":     {"pre-read", "{not json"},
		"empty stdin":        {"session-start", ""},
		"wrong event":        {"session-start", readEvent(t, dir)},
		"missing directory":  {"pre-read", readEvent(t, filepath.Join(dir, "no-existe"))},
		"unknown hook":       {"post-save", readEvent(t, dir)},
		"pre-save not wired": {"pre-save", readEvent(t, dir)},
	}
	for name, c := range cases {
		code, out, errOut := runHook(t, c.hook, c.event)
		if code != 0 || out != "" || errOut == "" {
			t.Errorf("%s: exit %d stdout %q stderr %q; want 0, no stdout and a message", name, code, out, errOut)
		}
	}
}

// T-F5-8: a slow Git does not hold the agent: the hook gives up within its budget.
func TestHookTimeoutNeverBlocks(t *testing.T) {
	oldDetect, oldBudget := memoryDetect, hookBudget
	defer func() { memoryDetect, hookBudget = oldDetect, oldBudget }()
	memoryDetect = func(string) (memory.Result, error) {
		time.Sleep(2 * time.Second)
		return memory.Result{Mode: memory.ModeCloudBranch, Tagged: true, Branch: "feature/x", Slug: "feature-x-217d2b"}, nil
	}
	hookBudget = 50 * time.Millisecond
	start := time.Now()
	code, out, errOut := runHook(t, "pre-read", readEvent(t, t.TempDir()))
	if code != 0 || out != "" || errOut == "" || time.Since(start) > time.Second {
		t.Fatalf("exit %d stdout %q stderr %q after %v", code, out, errOut, time.Since(start))
	}
}

// T-F5-12: the hook code uses no network and reads no environment variable.
func TestHookCodeIsOffline(t *testing.T) {
	src, err := os.ReadFile("memory_hook.go")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), "memory_hook.go", src, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, imp := range parsed.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		if path == "net" || strings.HasPrefix(path, "net/") {
			t.Errorf("memory_hook.go imports %q", path)
		}
	}
	for _, call := range []string{"os.Getenv", "os.LookupEnv", "os.Environ", `"ENGRAM_`} {
		if strings.Contains(string(src), call) {
			t.Errorf("memory_hook.go uses %s", call)
		}
	}
}
