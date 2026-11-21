package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// initRepo creates a repository and isolates Git and the home directory
// (backups go under it).
func initRepo(t *testing.T, remote string) (root, home string) {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	home = t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	root = filepath.Join(t.TempDir(), "mi repo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if remote != "" {
		if out, err := exec.Command("git", "-C", root, "remote", "add", "origin", remote).CombinedOutput(); err != nil {
			t.Fatalf("git remote add: %v\n%s", err, out)
		}
	}
	return root, home
}

func writeInitFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func initFileExists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// noTerminal makes the command believe nobody can answer a question.
func noTerminal(t *testing.T) {
	t.Helper()
	memoryInitInteractive = func() bool { return false }
	t.Cleanup(func() { memoryInitInteractive = defaultInteractive })
}

const initRemote = "https://example.invalid/team/active-stack.git"

func TestMemoryInitWritesAndProposesTheCommit(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	code, out, errOut := runMemory("init", "--project", root, "--yes", "--lang", "en")
	if code != 0 {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	for _, f := range []string{".engram/config.json", ".engram/active-stack.json", ".gitignore"} {
		if !initFileExists(root, f) {
			t.Errorf("%s was not written", f)
		}
	}
	for _, want := range []string{
		"active-stack", "created",
		"git add -- .engram/active-stack.json .engram/config.json .gitignore",
		`git commit -m "chore(engram): pin project and team memory policy" -- .engram/active-stack.json .engram/config.json .gitignore`,
		"NOT executed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout lacks %q:\n%s", want, out)
		}
	}
	// the commit is only proposed
	if staged, _ := exec.Command("git", "-C", root, "diff", "--cached", "--name-only").Output(); len(staged) != 0 {
		t.Fatalf("init staged files: %s", staged)
	}
}

func TestMemoryInitWithoutConfirmationStopsWithExit3(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	noTerminal(t)
	code, _, errOut := runMemory("init", "--project", root, "--lang", "en")
	if code != 3 {
		t.Fatalf("exit %d, want 3", code)
	}
	if !strings.Contains(errOut, `"active-stack"`) || !strings.Contains(errOut, "--yes") || !strings.Contains(errOut, "--name") {
		t.Errorf("stderr must show the derived name and both ways to confirm: %q", errOut)
	}
	if initFileExists(root, ".engram/config.json") || initFileExists(root, ".gitignore") {
		t.Fatal("nothing may be written")
	}
}

func TestMemoryInitAsksOnATerminal(t *testing.T) {
	answers := map[string]int{"y\n": 0, "s\n": 0, "n\n": 3, "\n": 3}
	for answer, wantCode := range answers {
		root, _ := initRepo(t, initRemote)
		memoryInitInteractive = func() bool { return true }
		memoryInitStdin = strings.NewReader(answer)
		code, _, errOut := runMemory("init", "--project", root, "--lang", "en")
		memoryInitInteractive, memoryInitStdin = defaultInteractive, os.Stdin
		if code != wantCode {
			t.Errorf("answer %q: exit %d, want %d (stderr %q)", answer, code, wantCode, errOut)
		}
		if !strings.Contains(errOut, `"active-stack"`) {
			t.Errorf("the prompt must show the name: %q", errOut)
		}
		if wantCode == 3 && initFileExists(root, ".engram/config.json") {
			t.Errorf("answer %q wrote files", answer)
		}
	}
}

func TestMemoryInitConflictExits3AndSaysHowToResolve(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	writeInitFile(t, root, ".engram/config.json", "{\"project_name\": \"otro\"}\n")
	code, _, errOut := runMemory("init", "--project", root, "--yes", "--lang", "en")
	if code != 3 {
		t.Fatalf("exit %d, want 3", code)
	}
	for _, want := range []string{`"otro"`, `"active-stack"`, "--name otro", "Nothing was changed"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr lacks %q: %q", want, errOut)
		}
	}
	if initFileExists(root, ".engram/active-stack.json") {
		t.Fatal("a conflict must write nothing")
	}
}

func TestMemoryInitErrorsExit1(t *testing.T) {
	root, home := initRepo(t, initRemote)
	writeInitFile(t, root, ".engram/config.json", "{broken")
	if code, _, errOut := runMemory("init", "--project", root, "--yes", "--lang", "en"); code != 1 || !strings.Contains(errOut, ".engram/config.json") {
		t.Errorf("invalid JSON: exit %d, %q", code, errOut)
	}
	if code, _, _ := runMemory("init", "--project", t.TempDir(), "--yes"); code != 1 {
		t.Errorf("not a repository: exit %d, want 1", code)
	}
	if code, _, _ := runMemory("init", "--project", filepath.Join(home, "no-existe"), "--yes"); code != 1 {
		t.Errorf("missing directory: exit %d, want 1", code)
	}
}

func TestMemoryInitRefusesTheHomeDirectory(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	t.Setenv("USERPROFILE", root)
	t.Setenv("HOME", root)
	code, _, errOut := runMemory("init", "--project", root, "--yes", "--lang", "en")
	if code != 1 || !strings.Contains(errOut, "home") {
		t.Fatalf("exit %d, %q", code, errOut)
	}
	if initFileExists(root, ".engram/config.json") {
		t.Fatal("wrote in the home directory")
	}
}

func TestMemoryInitUsageErrorsExit2(t *testing.T) {
	if code, _, _ := runMemory("init", "--no-such-flag"); code != 2 {
		t.Errorf("unknown flag: exit %d, want 2", code)
	}
	if code, _, _ := runMemory("init", "extra-arg"); code != 2 {
		t.Errorf("positional arg: exit %d, want 2", code)
	}
	if code, _, _ := runMemory("init", "--lang", "fr"); code != 2 {
		t.Errorf("bad language: exit %d, want 2", code)
	}
}

func TestMemoryInitJSONSchema(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	code, out, errOut := runMemory("init", "--project", root, "--yes", "--json")
	if code != 0 {
		t.Fatalf("exit %d, %q", code, errOut)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	want := []string{"backup_dir", "detail", "files", "name_source", "project_name", "proposed_commit", "reason", "root", "status", "warnings"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("keys %v, want %v", keys, want)
	}
	if got["status"] != "ok" || got["project_name"] != "active-stack" || got["name_source"] != "remote" || got["reason"] != nil || got["detail"] != nil {
		t.Errorf("%v", got)
	}
	pc := got["proposed_commit"].(map[string]any)
	if pc["message"] != "chore(engram): pin project and team memory policy" || len(pc["commands"].([]any)) != 2 {
		t.Errorf("proposed_commit %v", pc)
	}
	files := got["files"].([]any)
	if len(files) != 3 || files[0].(map[string]any)["action"] != "created" {
		t.Errorf("files %v", files)
	}
}

func TestMemoryInitJSONWhenItNeedsADecision(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	writeInitFile(t, root, ".engram/config.json", "{\"project_name\": \"otro\"}\n")
	code, out, _ := runMemory("init", "--project", root, "--yes", "--json")
	if code != 3 {
		t.Fatalf("exit %d", code)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	detail, _ := got["detail"].(map[string]any)
	if got["status"] != "needs_decision" || got["reason"] != "conflict_value" || detail["have"] != "otro" || detail["want"] != "active-stack" || detail["file"] != ".engram/config.json" {
		t.Errorf("%v", got)
	}

	// without confirmation
	root2, _ := initRepo(t, initRemote)
	noTerminal(t)
	code, out, _ = runMemory("init", "--project", root2, "--json")
	if code != 3 || !strings.Contains(out, `"name_not_confirmed"`) || !strings.Contains(out, `"derived_name":"active-stack"`) {
		t.Errorf("exit %d, %s", code, out)
	}
}

func TestMemoryInitIsBilingualAndKnownToTheUsage(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	writeInitFile(t, root, ".engram/config.json", "{\"project_name\": \"otro\"}\n")
	_, _, es := runMemory("init", "--project", root, "--yes", "--lang", "es")
	if !strings.Contains(es, "No se cambió nada") {
		t.Errorf("Spanish message missing: %q", es)
	}
	_, _, usage := runMemory("init", "--bogus", "--lang", "en")
	if !strings.Contains(usage, "memory init") {
		t.Errorf("the usage must list memory init: %q", usage)
	}
}

func TestMemoryInitIsIdempotent(t *testing.T) {
	root, _ := initRepo(t, initRemote)
	runMemory("init", "--project", root, "--yes")
	_, out, _ := runMemory("init", "--project", root, "--yes", "--lang", "en")
	if strings.Count(out, "unchanged") != 3 {
		t.Errorf("a second run must report 3 unchanged files:\n%s", out)
	}
}

// Every message of "memory init", in both languages, must be fully formatted:
// a wrong number of arguments prints "%!" or "%!(EXTRA", and go vet cannot see
// it through i18n.Tf.
func TestMemoryInitMessagesAreFullyFormatted(t *testing.T) {
	type scenario struct {
		name  string
		setup func(root string)
		args  []string
		tty   bool
	}
	scenarios := []scenario{
		{"ok", func(string) {}, []string{"--yes"}, false},
		{"ok with warnings", func(r string) {
			writeInitFile(t, r, ".gitignore", "a\n")
			for _, a := range [][]string{{"add", ".gitignore"}, {"-c", "user.name=T", "-c", "user.email=t@e.invalid", "commit", "-q", "-m", "x"}} {
				if out, err := exec.Command("git", append([]string{"-C", r}, a...)...).CombinedOutput(); err != nil {
					t.Fatalf("%v\n%s", err, out)
				}
			}
			writeInitFile(t, r, ".gitignore", "a\nb\n")
		}, []string{"--yes"}, false},
		{"conflict name", func(r string) { writeInitFile(t, r, ".engram/config.json", "{\"project_name\": \"otro\"}") }, []string{"--yes"}, false},
		{"conflict policy", func(r string) { writeInitFile(t, r, ".engram/active-stack.json", "{\"memory_policy\": \"deferred\"}") }, []string{"--yes"}, false},
		{"ignored dir", func(r string) { writeInitFile(t, r, ".gitignore", ".engram/\n") }, []string{"--yes"}, false},
		{"invalid file", func(r string) { writeInitFile(t, r, ".engram/config.json", "{bad") }, []string{"--yes"}, false},
		{"not confirmed", func(string) {}, nil, false},
		{"declined", func(string) {}, nil, true},
		{"not a repo", func(string) {}, []string{"--yes", "--project", "NOT-A-REPO"}, false},
	}
	for _, lang := range []string{"en", "es"} {
		for _, sc := range scenarios {
			t.Run(lang+" "+sc.name, func(t *testing.T) {
				root, _ := initRepo(t, initRemote)
				sc.setup(root)
				args := append([]string{"init", "--lang", lang, "--project", root}, sc.args...)
				for i, a := range args { // "NOT-A-REPO" stands for a directory that is not a repository
					if a == "NOT-A-REPO" {
						args[i-1], args[i] = "--project", t.TempDir()
					}
				}
				memoryInitInteractive = func() bool { return sc.tty }
				memoryInitStdin = strings.NewReader("n\n")
				defer func() { memoryInitInteractive, memoryInitStdin = defaultInteractive, os.Stdin }()
				_, out, errOut := runMemory(args...)
				for _, text := range []string{out, errOut} {
					if strings.Contains(text, "%!") || strings.Contains(text, "EXTRA") || strings.Contains(text, "memory.init.") {
						t.Errorf("badly formatted or missing message:\n%s", text)
					}
				}
			})
		}
	}
}
