package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// memoryRepo creates an offline repo on feature/x with the cloud-branch
// marker and an origin whose HEAD is main.
func memoryRepo(t *testing.T, marker string) string {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	dir := filepath.Join(t.TempDir(), "mi repo")
	if err := os.MkdirAll(filepath.Join(dir, ".engram"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".engram", "active-stack.json"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "-A"},
		{"commit", "-q", "-m", "init"},
		{"remote", "add", "origin", "https://example.invalid/team/repo.git"},
		{"update-ref", "refs/remotes/origin/main", "HEAD"},
		{"symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main"},
		{"switch", "-q", "-c", "feature/x"},
	} {
		full := append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func runMemory(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := runMemoryDispatch(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// T-F4-9: --json prints exactly the stable schema {"mode","tagged","branch","slug"}.
func TestMemoryBranchJSONSchema(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	code, out, errOut := runMemory("branch", "--project", dir, "--json")
	if code != 0 {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	want := map[string]any{"mode": "cloud-branch", "tagged": true, "branch": "feature/x", "slug": "feature-x"}
	if len(got) != len(want) {
		t.Fatalf("keys = %v; want exactly %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v; want %v", k, got[k], v)
		}
	}
}

// T-F4-9: absent values are JSON null, so the keys never change.
func TestMemoryBranchJSONNullsWhenNotTagged(t *testing.T) {
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	code, out, _ := runMemory("branch", "--project", t.TempDir(), "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	want := `{"mode":"off","tagged":false,"branch":null,"slug":null}`
	if strings.TrimSpace(out) != want {
		t.Fatalf("stdout = %s; want %s", out, want)
	}
}

// T-F4-9 and T-F4-12: readable output in both languages.
func TestMemoryBranchReadableOutputIsBilingual(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	_, es, _ := runMemory("branch", "--project", dir, "--lang", "es")
	for _, want := range []string{"Modo de memoria: cloud-branch", "Etiqueta: sí", "Rama: feature/x", "Slug: feature-x"} {
		if !strings.Contains(es, want) {
			t.Errorf("es output missing %q:\n%s", want, es)
		}
	}
	_, en, _ := runMemory("branch", "--project", dir, "--lang", "en")
	for _, want := range []string{"Memory mode: cloud-branch", "Tagged: yes", "Branch: feature/x", "Slug: feature-x"} {
		if !strings.Contains(en, want) {
			t.Errorf("en output missing %q:\n%s", want, en)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(es), "{") {
		t.Errorf("readable output must not be JSON:\n%s", es)
	}
}

// T-F4-8 and T-F4-12: an invalid marker warns on stderr, in the chosen
// language, and stdout keeps the stable JSON.
func TestMemoryBranchMarkerWarningGoesToStderr(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-forever"}`)
	code, out, errOut := runMemory("branch", "--project", dir, "--json", "--lang", "es")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, `"mode":"off"`) {
		t.Errorf("stdout = %s; want mode off", out)
	}
	if !strings.Contains(errOut, "cloud-forever") || !strings.Contains(errOut, "Se ignora") {
		t.Errorf("stderr = %q; want the Spanish warning naming the policy", errOut)
	}
}

// T-F4-12: exit codes. 0 ok, 1 cannot decide, 2 usage error.
func TestMemoryExitCodes(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	cases := []struct {
		name string
		args []string
		code int
	}{
		{"ok", []string{"branch", "--project", dir}, 0},
		{"missing directory", []string{"branch", "--project", filepath.Join(dir, "no-existe")}, 1},
		{"no subcommand", nil, 2},
		{"unknown subcommand", []string{"tag"}, 2},
		{"unknown flag", []string{"branch", "--nope"}, 2},
		{"bad language", []string{"branch", "--lang", "fr"}, 2},
		{"extra argument", []string{"branch", "extra"}, 2},
	}
	for _, c := range cases {
		if code, _, errOut := runMemory(c.args...); code != c.code {
			t.Errorf("%s: exit %d (stderr %q); want %d", c.name, code, errOut, c.code)
		}
	}
}

// T-F4-12: a branch with no letters or digits cannot be tagged: exit 1 with
// a message in the chosen language.
func TestMemoryBranchEmptySlugIsAnError(t *testing.T) {
	dir := memoryRepo(t, `{"memory_policy":"cloud-branch"}`)
	if out, err := exec.Command("git", "-C", dir, "switch", "-q", "-c", "ñ").CombinedOutput(); err != nil {
		t.Fatalf("git switch: %v\n%s", err, out)
	}
	code, out, errOut := runMemory("branch", "--project", dir, "--lang", "es")
	if code != 1 || out != "" {
		t.Fatalf("exit %d stdout %q; want 1 and no stdout", code, out)
	}
	if !strings.Contains(errOut, "ñ") || !strings.Contains(errOut, "etiqueta") {
		t.Errorf("stderr = %q; want the Spanish empty-slug message", errOut)
	}
}

// The command is reachable from the top-level dispatch.
func TestMemoryCommandIsRouted(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `case "memory":`) || !strings.Contains(string(src), "runMemoryDispatch(os.Args[2:], os.Stdout, os.Stderr)") {
		t.Fatal(`main.go must route "memory" to runMemoryDispatch`)
	}
}
