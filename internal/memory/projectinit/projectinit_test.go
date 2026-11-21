package projectinit

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/memory"
)

// ---- helpers ---------------------------------------------------------------

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// isolateGit gives the test an empty global and system Git configuration.
// It returns the path of the global file so a test can add to it.
func isolateGit(t *testing.T) string {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return global
}

// newRepo creates a repository with one commit. remote "" means no remote.
func newRepo(t *testing.T, name, remote string) (root, home string) {
	t.Helper()
	isolateGit(t)
	root = filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "init")
	if remote != "" {
		git(t, root, "remote", "add", "origin", remote)
	}
	return root, t.TempDir()
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// snapshotTree returns every file under root (outside .git) with its content.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if fi.IsDir() {
			if rel == ".git" {
				return filepath.SkipDir
			}
			if rel != "." {
				out[filepath.ToSlash(rel)+"/"] = ""
			}
			return nil
		}
		b, _ := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func run(t *testing.T, root, home string, mut func(*Options)) (Result, error) {
	t.Helper()
	opts := Options{Dir: root, Yes: true, Home: home}
	if mut != nil {
		mut(&opts)
	}
	return Init(opts)
}

func mustRun(t *testing.T, root, home string, mut func(*Options)) Result {
	t.Helper()
	res, err := run(t, root, home, mut)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	return res
}

func actions(res Result) map[string]Action {
	m := map[string]Action{}
	for _, f := range res.Files {
		m[f.Path] = f.Action
	}
	return m
}

func jsonKeys(t *testing.T, content string) []string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, content)
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

const remote = "https://example.invalid/team/active-stack.git"

// ---- T-F7-6: name derivation -----------------------------------------------

func TestNameFromRemoteURL(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://github.com/Group-Active-IA/active-stack.git", "active-stack"},
		{"https://github.com/Group-Active-IA/active-stack", "active-stack"},
		{"https://github.com/Group-Active-IA/active-stack/", "active-stack"},
		{"https://github.com/Group-Active-IA/active-stack.git/", "active-stack"},
		{"git@github.com:org/Mi-Repo.git", "Mi-Repo"},
		{"git@github.com:Mi-Repo.git", "Mi-Repo"},
		{"ssh://git@host.example:2222/org/repo.git", "repo"},
		{"https://host.example/org/Mi%20Repo.git", "Mi-Repo"},
		{"https://host.example/org/Mi Repo.git", "Mi-Repo"},
		{"https://host.example/org/Proyecto-Ñandú.git", "Proyecto-Ñandú"},
		{`C:\Users\persona\repos\demo.git`, "demo"},
		{"file:///c:/repos/demo.git", "demo"},
		{"/srv/git/demo.git", "demo"},
	}
	for _, c := range cases {
		got, err := nameFromRemoteURL(c.url)
		if err != nil || got != c.want {
			t.Errorf("nameFromRemoteURL(%q) = %q, %v; want %q", c.url, got, err, c.want)
		}
	}
	if _, err := nameFromRemoteURL("https://host.example/"); !errors.Is(err, ErrEmptyName) {
		t.Errorf("a URL with no repository segment must be ErrEmptyName, got %v", err)
	}
}

func TestSanitizeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"active-stack", "active-stack"},
		{"Mi App", "Mi-App"},
		{"  Mi   App  ", "Mi-App"},
		{"a/b\\c", "a-b-c"},
		{"Ñandú_2", "Ñandú_2"},
		{"MAYÚSCULAS", "MAYÚSCULAS"},
	}
	for _, c := range cases {
		got, err := sanitizeName(c.in)
		if err != nil || got != c.want {
			t.Errorf("sanitizeName(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, bad := range []string{"", "   ", "/", "-"} {
		if _, err := sanitizeName(bad); !errors.Is(err, ErrEmptyName) {
			t.Errorf("sanitizeName(%q) must be ErrEmptyName, got %v", bad, err)
		}
	}
}

func TestDeriveNameIgnoresRemotesOtherThanOrigin(t *testing.T) {
	root, _ := newRepo(t, "mi repo", "")
	git(t, root, "remote", "add", "upstream", "https://example.invalid/team/otro-nombre.git")
	name, src, err := DeriveName(root)
	if err != nil || name != "mi-repo" || src != SourceDirectory {
		t.Fatalf("only origin counts (R17-C): %q %q %v", name, src, err)
	}
}

func TestDeriveNameFromOriginAndFromCheckoutDirectory(t *testing.T) {
	root, _ := newRepo(t, "carpeta distinta", remote)
	name, src, err := DeriveName(root)
	if err != nil || name != "active-stack" || src != SourceRemote {
		t.Fatalf("with origin: %q %q %v", name, src, err)
	}

	root2, _ := newRepo(t, "mi repo", "")
	name, src, err = DeriveName(root2)
	if err != nil || name != "mi-repo" || src != SourceDirectory {
		t.Fatalf("without remote: %q %q %v", name, src, err)
	}
}

func TestDeriveNameInALinkedWorktreeUsesTheMainCheckout(t *testing.T) {
	root, _ := newRepo(t, "principal", "")
	wt := filepath.Join(t.TempDir(), "otro-worktree")
	git(t, root, "worktree", "add", "-q", "-b", "feat/x", wt)
	name, src, err := DeriveName(wt)
	if err != nil || name != "principal" || src != SourceDirectory {
		t.Fatalf("worktree: %q %q %v; want principal from the main checkout", name, src, err)
	}
}

func TestNameConfirmation(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	before := snapshotTree(t, root)

	// nobody to ask and no flag: stop, show the derived name, write nothing
	_, err := run(t, root, home, func(o *Options) { o.Yes = false })
	var nc *NeedsConfirmationError
	if !errors.As(err, &nc) || nc.Derived != "active-stack" || nc.Declined {
		t.Fatalf("want NeedsConfirmationError(active-stack), got %v", err)
	}
	// the person says no
	_, err = run(t, root, home, func(o *Options) {
		o.Yes = false
		o.Confirm = func(string) (bool, error) { return false, nil }
	})
	if !errors.As(err, &nc) || !nc.Declined {
		t.Fatalf("a declined name must be a declined NeedsConfirmationError, got %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatal("nothing may be written before the name is confirmed")
	}

	// the person says yes
	var asked string
	res := mustRun(t, root, home, func(o *Options) {
		o.Yes = false
		o.Confirm = func(d string) (bool, error) { asked = d; return true, nil }
	})
	if asked != "active-stack" || res.ProjectName != "active-stack" || res.NameSource != SourceRemote {
		t.Fatalf("asked %q, result %+v", asked, res)
	}
}

func TestExplicitNameIsUsedAndSanitized(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	res := mustRun(t, root, home, func(o *Options) { o.Yes = false; o.Name = "Mi Proyecto" })
	if res.ProjectName != "Mi-Proyecto" || res.NameSource != SourceFlag {
		t.Fatalf("%+v", res)
	}
	if !strings.Contains(read(t, root, ConfigPath), `"project_name": "Mi-Proyecto"`) {
		t.Fatalf("config.json: %s", read(t, root, ConfigPath))
	}
}

// ---- T-F7-1: creation and preservation ------------------------------------

func TestInitCreatesTheThreeFiles(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	res := mustRun(t, root, home, nil)

	if got := read(t, root, ConfigPath); got != "{\n  \"project_name\": \"active-stack\"\n}\n" {
		t.Errorf("config.json = %q", got)
	}
	if got := read(t, root, MarkerPath); got != "{\n  \"memory_policy\": \"cloud-branch\"\n}\n" {
		t.Errorf("active-stack.json = %q", got)
	}
	want := "# active-stack:memory begin\n.engram/*\n!.engram/config.json\n!.engram/active-stack.json\n# active-stack:memory end\n"
	if got := read(t, root, GitignorePath); got != want {
		t.Errorf(".gitignore = %q", got)
	}
	a := actions(res)
	for _, f := range []string{ConfigPath, MarkerPath, GitignorePath} {
		if a[f] != ActionCreated {
			t.Errorf("%s: %s, want created", f, a[f])
		}
	}
}

func TestInitKeepsOtherKeysAndTheFormatOfAnExistingConfig(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\r\n    \"org\": \"acme\",\r\n    \"extra\": [1, 2]\r\n}\r\n")
	write(t, root, MarkerPath, "{\n\t\"note\": \"mine\"\n}\n")
	res := mustRun(t, root, home, nil)

	if got := read(t, root, ConfigPath); got != "{\r\n    \"org\": \"acme\",\r\n    \"extra\": [1, 2],\r\n    \"project_name\": \"active-stack\"\r\n}\r\n" {
		t.Errorf("config.json = %q", got)
	}
	if got := read(t, root, MarkerPath); got != "{\n\t\"note\": \"mine\",\n\t\"memory_policy\": \"cloud-branch\"\n}\n" {
		t.Errorf("active-stack.json = %q", got)
	}
	if a := actions(res); a[ConfigPath] != ActionUpdated || a[MarkerPath] != ActionUpdated {
		t.Errorf("actions %v", a)
	}
}

func TestInitAcceptsABOMInTheMarkerAndKeepsIt(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	// already says cloud-branch: nothing to do, not an error
	write(t, root, MarkerPath, "\xef\xbb\xbf{\"memory_policy\":\"cloud-branch\"}")
	res := mustRun(t, root, home, nil)
	if actions(res)[MarkerPath] != ActionUnchanged || read(t, root, MarkerPath) != "\xef\xbb\xbf{\"memory_policy\":\"cloud-branch\"}" {
		t.Fatalf("a marker with BOM that already says cloud-branch must stay: %q", read(t, root, MarkerPath))
	}
	// needs a key: the BOM stays
	write(t, root, MarkerPath, "\xef\xbb\xbf{\"note\":1}")
	mustRun(t, root, home, nil)
	got := read(t, root, MarkerPath)
	if !strings.HasPrefix(got, "\xef\xbb\xbf{") || !strings.Contains(got, `"memory_policy": "cloud-branch"`) {
		t.Fatalf("the BOM must be kept and the key added: %q", got)
	}
}

// ---- T-F7-2: idempotent ----------------------------------------------------

func TestInitIsIdempotentAndProposesNothingOnceCommitted(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	first := mustRun(t, root, home, nil)
	if first.Proposal == nil {
		t.Fatal("the first run must propose a commit")
	}
	tree := snapshotTree(t, root)

	second := mustRun(t, root, home, nil)
	for p, a := range actions(second) {
		if a != ActionUnchanged {
			t.Errorf("second run: %s is %s, want unchanged", p, a)
		}
	}
	if !reflect.DeepEqual(tree, snapshotTree(t, root)) {
		t.Fatal("a second run must not change a single byte")
	}
	if second.Proposal == nil {
		t.Fatal("until the person commits, the pending files are still proposed")
	}

	git(t, root, "add", "--", ".engram/config.json", ".engram/active-stack.json", ".gitignore")
	git(t, root, "commit", "-q", "-m", "x")
	third := mustRun(t, root, home, nil)
	if third.Proposal != nil {
		t.Fatalf("nothing is pending, nothing to propose: %+v", third.Proposal)
	}
}

// ---- T-F7-3: never overwrite in silence ------------------------------------

func TestConflictingProjectNameStopsAndWritesNothing(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\"project_name\": \"otro\"}\n")
	before := snapshotTree(t, root)

	_, err := run(t, root, home, nil)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("want ConflictError, got %v", err)
	}
	if ce.File != ConfigPath || ce.Key != "project_name" || ce.Have != "otro" || ce.Want != "active-stack" || ce.Reason != ReasonValue {
		t.Fatalf("%+v", ce)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatal("a conflict must leave every file as it was (no marker, no .gitignore)")
	}
	// adopting the existing name is how the person resolves it
	res := mustRun(t, root, home, func(o *Options) { o.Name = "otro" })
	if res.ProjectName != "otro" || actions(res)[ConfigPath] != ActionUnchanged {
		t.Fatalf("%+v", res)
	}
}

func TestConflictingPolicyAndNonStringNameStop(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, MarkerPath, "{\"memory_policy\": \"deferred\"}\n")
	var ce *ConflictError
	if _, err := run(t, root, home, nil); !errors.As(err, &ce) || ce.File != MarkerPath || ce.Key != "memory_policy" || ce.Have != "deferred" {
		t.Fatalf("policy: %v", err)
	}
	write(t, root, MarkerPath, "{\"memory_policy\": \"cloud-branch\"}\n")
	write(t, root, ConfigPath, "{\"project_name\": 7}\n")
	if _, err := run(t, root, home, nil); !errors.As(err, &ce) || ce.Key != "project_name" {
		t.Fatalf("non-string name: %v", err)
	}
	write(t, root, ConfigPath, "{\"project_name\": \"\"}\n")
	if _, err := run(t, root, home, nil); !errors.As(err, &ce) {
		t.Fatalf("an empty name is still a value of the person: %v", err)
	}
}

// ---- T-F7-4: backup ---------------------------------------------------------

func TestSnapshotHoldsTheExistingFilesBeforeAnyWrite(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\"org\": \"acme\"}\n")
	write(t, root, GitignorePath, "node_modules/\n")

	var got []string
	var sawBefore map[string]string
	restore := SetSnapshotCreate(func(dir string, paths []string) error {
		got = append(got, paths...)
		sawBefore = snapshotTree(t, root)
		if !strings.HasPrefix(dir, filepath.Join(home, ".active-stack", "backups")) {
			t.Errorf("backup dir %q is not under the home backups", dir)
		}
		return nil
	})
	defer restore()
	mustRun(t, root, home, nil)

	want := []string{filepath.Join(root, ".engram", "config.json"), filepath.Join(root, ".gitignore")}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot paths %v, want only the files that already existed: %v", got, want)
	}
	if sawBefore[".engram/config.json"] != "{\"org\": \"acme\"}\n" || sawBefore[".gitignore"] != "node_modules/\n" || sawBefore[".engram/active-stack.json"] != "" {
		t.Fatalf("the snapshot must be taken before the first write: %v", sawBefore)
	}
}

func TestNoSnapshotWhenNothingExistedOrNothingChanges(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	calls := 0
	restore := SetSnapshotCreate(func(string, []string) error { calls++; return nil })
	defer restore()
	mustRun(t, root, home, nil) // everything is new
	if calls != 0 {
		t.Fatalf("snapshot called %d times with nothing to protect", calls)
	}
	mustRun(t, root, home, nil) // nothing changes
	if calls != 0 {
		t.Fatalf("snapshot called %d times with nothing to change", calls)
	}
}

func TestFailedSnapshotTouchesNothing(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\"org\": \"acme\"}\n")
	before := snapshotTree(t, root)
	restore := SetSnapshotCreate(func(string, []string) error { return errors.New("disk full") })
	defer restore()

	_, err := run(t, root, home, nil)
	var be *BackupError
	if !errors.As(err, &be) {
		t.Fatalf("want BackupError, got %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatal("if the backup fails no file may be touched")
	}
}

// ---- T-F7-5: malformed JSON ------------------------------------------------

func TestMalformedJSONIsAClearErrorAndWritesNothing(t *testing.T) {
	bad := map[string]string{
		"invalid":        "{not json",
		"comment":        "{\n// hi\n\"a\": 1}",
		"trailing comma": "{\"a\": 1,}",
		"array root":     "[1, 2]",
		"text":           "hello",
	}
	for name, content := range bad {
		for _, file := range []string{ConfigPath, MarkerPath} {
			t.Run(name+" "+file, func(t *testing.T) {
				root, home := newRepo(t, "r", remote)
				write(t, root, file, content)
				before := snapshotTree(t, root)
				_, err := run(t, root, home, nil)
				var ie *InvalidFileError
				if !errors.As(err, &ie) || ie.File != file {
					t.Fatalf("want InvalidFileError for %s, got %v", file, err)
				}
				if !reflect.DeepEqual(before, snapshotTree(t, root)) {
					t.Fatal("an invalid file must stop the run before any write")
				}
			})
		}
	}
}

func TestEmptyFilesAreFilled(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "")
	write(t, root, MarkerPath, " \n")
	res := mustRun(t, root, home, nil)
	if a := actions(res); a[ConfigPath] != ActionUpdated || a[MarkerPath] != ActionUpdated {
		t.Fatalf("%v", a)
	}
	if !reflect.DeepEqual(jsonKeys(t, read(t, root, ConfigPath)), []string{"project_name"}) {
		t.Fatal(read(t, root, ConfigPath))
	}
}

// ---- T-F7-7: .gitignore ----------------------------------------------------

func TestGitignoreKeepsTheUsersRulesAndTheirLineEndings(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	user := "node_modules/\r\n*.log\r\n"
	write(t, root, GitignorePath, user)
	mustRun(t, root, home, nil)
	got := read(t, root, GitignorePath)
	want := user + "\r\n# active-stack:memory begin\r\n.engram/*\r\n!.engram/config.json\r\n!.engram/active-stack.json\r\n# active-stack:memory end\r\n"
	if got != want {
		t.Fatalf(".gitignore = %q\nwant        %q", got, want)
	}
	// a file with no final newline
	write(t, root, GitignorePath, "build")
	mustRun(t, root, home, nil)
	if got := read(t, root, GitignorePath); !strings.HasPrefix(got, "build\n\n# active-stack:memory begin\n") {
		t.Fatalf("%q", got)
	}
}

func TestGitignoreBlockIsNotDuplicatedAndIsRefreshed(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, GitignorePath, "a\n# active-stack:memory begin\n.engram/old\n# active-stack:memory end\nb\n")
	mustRun(t, root, home, nil)
	got := read(t, root, GitignorePath)
	if strings.Count(got, BlockBegin) != 1 || strings.Contains(got, ".engram/old") || !strings.HasPrefix(got, "a\n") || !strings.HasSuffix(got, "# active-stack:memory end\nb\n") {
		t.Fatalf("%q", got)
	}
	mustRun(t, root, home, nil)
	if read(t, root, GitignorePath) != got {
		t.Fatal("second run changed the file")
	}
}

func TestGitignoreBrokenMarkersAreRefused(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, GitignorePath, "# active-stack:memory begin\n.engram/*\n")
	before := snapshotTree(t, root)
	_, err := run(t, root, home, nil)
	var ie *InvalidFileError
	if !errors.As(err, &ie) || ie.File != GitignorePath {
		t.Fatalf("a begin without end must be refused, got %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatal("nothing may be written")
	}
}

func TestAUserRuleThatIgnoresTheWholeDirectoryIsAConflict(t *testing.T) {
	for _, rule := range []string{".engram/", ".engram", "/.engram", "/.engram/", "**/.engram/", ".engram/\r"} {
		t.Run(rule, func(t *testing.T) {
			root, home := newRepo(t, "r", remote)
			write(t, root, GitignorePath, "x\n"+rule+"\n")
			before := snapshotTree(t, root)
			_, err := run(t, root, home, nil)
			var ce *ConflictError
			if !errors.As(err, &ce) || ce.File != GitignorePath || ce.Reason != ReasonIgnoredDir {
				t.Fatalf("want ConflictError ignored_dir, got %v", err)
			}
			if !reflect.DeepEqual(before, snapshotTree(t, root)) {
				t.Fatal("the person's line must not be edited and nothing written")
			}
		})
	}
}

func TestGitConfirmsOnlyTheTwoFilesAreVersionable(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	mustRun(t, root, home, nil)
	write(t, root, ".engram/engram.db", "db")
	write(t, root, ".engram/chunks/a.jsonl", "x")

	status := git(t, root, "status", "--porcelain", "-uall")
	var files []string
	for _, l := range strings.Split(status, "\n") {
		files = append(files, strings.TrimSpace(l[2:]))
	}
	sort.Strings(files)
	want := []string{".engram/active-stack.json", ".engram/config.json", ".gitignore"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("git would version %v, want exactly %v", files, want)
	}
}

func TestAGlobalIgnoreThatStillHidesTheFilesRollsEverythingBack(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	// a rule outside the repository that ignores .engram/ as a whole: the
	// .gitignore of the project cannot undo it
	global := filepath.Join(t.TempDir(), "global-ignore")
	if err := os.WriteFile(global, []byte(".engram/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "config", "core.excludesFile", global)
	before := snapshotTree(t, root)

	_, err := run(t, root, home, nil)
	var ce *ConflictError
	if !errors.As(err, &ce) || ce.Reason != ReasonStillIgnored {
		t.Fatalf("want ConflictError still_ignored, got %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatal("after a failed verification every file must be put back")
	}
}

// ---- T-F7-8: only proposes -------------------------------------------------

func TestProposalIsPrintedNotExecuted(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	headBefore := git(t, root, "rev-parse", "HEAD")
	res := mustRun(t, root, home, nil)

	if git(t, root, "rev-parse", "HEAD") != headBefore {
		t.Fatal("init must not commit")
	}
	if staged := git(t, root, "diff", "--cached", "--name-only"); staged != "" {
		t.Fatalf("init must not stage anything, staged: %q", staged)
	}
	p := res.Proposal
	if p == nil || p.Message != "chore(engram): pin project and team memory policy" {
		t.Fatalf("%+v", p)
	}
	wantFiles := []string{".engram/active-stack.json", ".engram/config.json", ".gitignore"}
	if !reflect.DeepEqual(p.Files, wantFiles) {
		t.Fatalf("files %v, want %v", p.Files, wantFiles)
	}
	wantCmds := []string{
		`git add -- .engram/active-stack.json .engram/config.json .gitignore`,
		`git commit -m "chore(engram): pin project and team memory policy" -- .engram/active-stack.json .engram/config.json .gitignore`,
	}
	if !reflect.DeepEqual(p.Commands, wantCmds) {
		t.Fatalf("commands %q", p.Commands)
	}
}

func TestProposalCarriesOnlyPendingFiles(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	// the person already versioned their .gitignore with our block and the other two files
	mustRun(t, root, home, nil)
	git(t, root, "add", "--", ".gitignore", ".engram/active-stack.json")
	git(t, root, "commit", "-q", "-m", "partial")
	res := mustRun(t, root, home, nil)
	if res.Proposal == nil || !reflect.DeepEqual(res.Proposal.Files, []string{".engram/config.json"}) {
		t.Fatalf("%+v", res.Proposal)
	}
}

func TestWarnsAboutUncommittedChangesAndOtherVersionedFiles(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, GitignorePath, "a\n")
	git(t, root, "add", ".gitignore")
	git(t, root, "commit", "-q", "-m", "gi")
	write(t, root, GitignorePath, "a\nmine\n") // uncommitted change of the person
	write(t, root, ".engram/engram.db", "db")
	git(t, root, "add", "-f", ".engram/engram.db")
	git(t, root, "commit", "-q", "-m", "db")

	res := mustRun(t, root, home, nil)
	byCode := map[string][]string{}
	for _, w := range res.Warnings {
		byCode[w.Code] = w.Files
	}
	if !reflect.DeepEqual(byCode[WarnDirtyFiles], []string{".gitignore"}) {
		t.Errorf("dirty: %v", byCode)
	}
	if !reflect.DeepEqual(byCode[WarnTrackedOther], []string{".engram/engram.db"}) {
		t.Errorf("tracked: %v", byCode)
	}
}

// ---- T-F7-9: scope ---------------------------------------------------------

func TestRefusesAnythingThatIsNotAProjectRoot(t *testing.T) {
	root, home := newRepo(t, "r", remote)

	// the repository root is the person's home
	_, err := Init(Options{Dir: root, Yes: true, Home: root})
	if !errors.Is(err, ErrHomeDirectory) {
		t.Fatalf("home as root: %v", err)
	}
	if exists(root, ".engram") || exists(root, GitignorePath) {
		t.Fatal("nothing may be written in the home directory")
	}

	// not a repository
	plain := t.TempDir()
	if _, err := Init(Options{Dir: plain, Yes: true, Home: home}); !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("no repo: %v", err)
	}
	if _, err := Init(Options{Dir: filepath.Join(plain, "no-existe"), Yes: true, Home: home}); err == nil {
		t.Fatal("a missing directory must fail")
	}
}

func TestWritesOnlyInsideTheRootAndTheBackupFolder(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, GitignorePath, "a\n") // so that a backup exists
	outside := filepath.Dir(root)
	beforeOutside := snapshotTree(t, outside)
	mustRun(t, root, home, nil)

	afterOutside := snapshotTree(t, outside)
	for p := range afterOutside {
		if _, ok := beforeOutside[p]; !ok && !strings.HasPrefix(p, filepath.Base(root)+"/") {
			t.Errorf("wrote outside the project: %s", p)
		}
	}
	for p := range snapshotTree(t, home) {
		if !strings.HasPrefix(p, ".active-stack/") {
			t.Errorf("wrote in home outside .active-stack/: %s", p)
		}
	}
}

func TestWritesInTheRootOfTheCurrentWorktree(t *testing.T) {
	root, home := newRepo(t, "principal", remote)
	wt := filepath.Join(t.TempDir(), "wt")
	git(t, root, "worktree", "add", "-q", "-b", "feat/y", wt)
	sub := filepath.Join(wt, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	res := mustRun(t, root, home, func(o *Options) { o.Dir = sub })
	if !exists(wt, ConfigPath) || exists(root, ConfigPath) {
		t.Fatal("the files belong to the worktree where the command ran, at its root")
	}
	if got, _ := filepath.EvalSymlinks(res.Root); got != func() string { g, _ := filepath.EvalSymlinks(wt); return g }() {
		t.Fatalf("root %q", res.Root)
	}
}

// ---- T-F7-10: rollback -----------------------------------------------------

func TestAFailureAfterTheFirstWriteRestoresEverything(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\"org\": \"acme\"}\n") // exists: must come back
	// marker and .gitignore do not exist: must disappear
	before := snapshotTree(t, root)

	restore := SetWriteFile(func(path string, content []byte) error {
		if strings.HasSuffix(filepath.ToSlash(path), GitignorePath) {
			return errors.New("injected failure")
		}
		return defaultWriteFile(path, content)
	})
	defer restore()

	_, err := run(t, root, home, nil)
	if err == nil || !strings.Contains(err.Error(), "injected failure") {
		t.Fatalf("want the injected failure, got %v", err)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatalf("rollback must restore the original tree\nbefore %v\nafter  %v", before, snapshotTree(t, root))
	}
}

func TestAFailureOnTheFirstWriteRemovesTheDirectoryItCreated(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	before := snapshotTree(t, root)
	n := 0
	restore := SetWriteFile(func(path string, content []byte) error {
		n++
		if n == 2 {
			return errors.New("injected failure")
		}
		return defaultWriteFile(path, content)
	})
	defer restore()
	_, err := run(t, root, home, nil)
	if err == nil || !strings.Contains(err.Error(), "injected failure") {
		t.Fatalf("want the injected failure, got %v", err)
	}
	if n != 2 {
		t.Fatalf("the writer was called %d times, want 2 (the first write must have happened and been undone)", n)
	}
	if !reflect.DeepEqual(before, snapshotTree(t, root)) {
		t.Fatalf("the .engram directory created by the run must be gone\nafter %v", snapshotTree(t, root))
	}
}

// ---- T-F7-11: no secrets ---------------------------------------------------

func TestWritesOnlyTheAllowedKeysAndNoSecrets(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	t.Setenv("ENGRAM_CLOUD_TOKEN", "secret-token-value")
	t.Setenv("ENGRAM_CLOUD_SERVER", "https://server.example")
	mustRun(t, root, home, nil)

	if !reflect.DeepEqual(jsonKeys(t, read(t, root, ConfigPath)), []string{"project_name"}) {
		t.Errorf("config.json keys: %s", read(t, root, ConfigPath))
	}
	if !reflect.DeepEqual(jsonKeys(t, read(t, root, MarkerPath)), []string{"memory_policy"}) {
		t.Errorf("marker keys: %s", read(t, root, MarkerPath))
	}
	for _, f := range []string{ConfigPath, MarkerPath, GitignorePath} {
		c := strings.ToLower(read(t, root, f))
		for _, bad := range []string{"secret-token-value", "server.example", "token", "server", "engram_"} {
			if strings.Contains(c, bad) {
				t.Errorf("%s contains %q", f, bad)
			}
		}
	}
}

// A file the person created and never committed is not "uncommitted changes to a
// versioned file": the commit would not drag anything else in.
func TestAnUntrackedExistingFileIsNotReportedAsDirty(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, ConfigPath, "{\"org\": \"acme\"}\n")
	write(t, root, GitignorePath, "build/\n")
	res := mustRun(t, root, home, nil)
	for _, w := range res.Warnings {
		if w.Code == WarnDirtyFiles {
			t.Fatalf("untracked files must not be reported as dirty: %v", w.Files)
		}
	}
}

// ---- the contract with R-08 / R-15 and the real backup ---------------------

// What init writes is what memory.Detect (R-08) and the hooks (R-15) read: the
// project must come out in cloud-branch mode, tagged on a feature branch.
func TestWhatInitWritesIsWhatMemoryDetectReads(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	git(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	git(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	mustRun(t, root, home, nil)

	onMain, err := memory.Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if onMain.Mode != memory.ModeCloudBranch || onMain.Tagged || len(onMain.Warnings) != 0 {
		t.Fatalf("on the primary branch: mode cloud-branch, not tagged, no warnings; got %+v", onMain)
	}

	git(t, root, "switch", "-q", "-c", "feature/x")
	onFeature, err := memory.Detect(root)
	if err != nil {
		t.Fatal(err)
	}
	if onFeature.Mode != memory.ModeCloudBranch || !onFeature.Tagged || onFeature.Slug != "feature-x" || len(onFeature.Warnings) != 0 {
		t.Fatalf("on a feature branch: cloud-branch and tagged feature-x; got %+v", onFeature)
	}
}

// T-F7-4 with the REAL snapshotter: the archive it leaves gives back the exact
// bytes of the files init was about to change.
func TestRealSnapshotRestoresTheOriginalFiles(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	origConfig := "{\r\n    \"org\": \"acme\"\r\n}\r\n"
	origIgnore := "node_modules/\r\n*.log\r\n"
	write(t, root, ConfigPath, origConfig)
	write(t, root, GitignorePath, origIgnore)

	res := mustRun(t, root, home, nil)
	if read(t, root, ConfigPath) == origConfig || read(t, root, GitignorePath) == origIgnore {
		t.Fatal("init changed nothing")
	}
	if res.BackupDir == "" || !strings.HasPrefix(res.BackupDir, filepath.Join(home, ".active-stack", "backups", "memory-init-")) {
		t.Fatalf("BackupDir %q", res.BackupDir)
	}
	manifest, err := backup.ReadManifest(filepath.Join(res.BackupDir, backup.ManifestFilename))
	if err != nil {
		t.Fatalf("no readable manifest in %s: %v", res.BackupDir, err)
	}
	if err := (backup.RestoreService{}).Restore(manifest); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if read(t, root, ConfigPath) != origConfig || read(t, root, GitignorePath) != origIgnore {
		t.Fatalf("the backup did not give the originals back:\nconfig  %q\nignore  %q", read(t, root, ConfigPath), read(t, root, GitignorePath))
	}
}

// R17-G, pinned so the ficha stays true: the backup folder is fixed per
// project, so each run that writes replaces the previous backup.
func TestEachWritingRunReplacesThePreviousBackup(t *testing.T) {
	root, home := newRepo(t, "r", remote)
	write(t, root, GitignorePath, "a\n")
	first := mustRun(t, root, home, nil)

	// the person's block drifts, so the next run has to rewrite the file
	drifted := strings.Replace(read(t, root, GitignorePath), ".engram/*", ".engram/old", 1)
	write(t, root, GitignorePath, drifted)
	second := mustRun(t, root, home, nil)
	if first.BackupDir != second.BackupDir {
		t.Fatalf("the backup folder must be the same for the same project: %q vs %q", first.BackupDir, second.BackupDir)
	}

	manifest, err := backup.ReadManifest(filepath.Join(second.BackupDir, backup.ManifestFilename))
	if err != nil {
		t.Fatal(err)
	}
	if err := (backup.RestoreService{}).Restore(manifest); err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, GitignorePath); got != drifted {
		t.Fatalf("the backup must hold the state before the LAST write (the drifted block), not the first original:\n%q", got)
	}
}
