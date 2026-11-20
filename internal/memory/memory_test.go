package memory

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// isolateGit keeps the host's global and system Git config out of the test,
// so init.defaultBranch and similar settings come only from the test repo.
func isolateGit(t *testing.T) {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// newRepo creates a repo on branch "main" with one commit. With marker true it
// commits .engram/active-stack.json with the cloud-branch policy.
func newRepo(t *testing.T, marker bool) string {
	t.Helper()
	isolateGit(t)
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-q", "-b", "main")
	if marker {
		writeMarker(t, dir, `{"memory_policy":"cloud-branch"}`)
	} else if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func writeMarker(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".engram"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".engram", "active-stack.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// addOrigin registers an offline origin whose HEAD points at head (empty: no
// origin/HEAD). No command touches the network.
func addOrigin(t *testing.T, dir, head string) {
	t.Helper()
	git(t, dir, "remote", "add", "origin", "https://example.invalid/team/repo.git")
	git(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	if head != "" {
		git(t, dir, "update-ref", "refs/remotes/origin/"+head, "HEAD")
		git(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+head)
	}
}

func detect(t *testing.T, dir string) Result {
	t.Helper()
	res, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect(%q) error = %v", dir, err)
	}
	return res
}

func assertResult(t *testing.T, got Result, mode string, tagged bool, branch, slug string) {
	t.Helper()
	if got.Mode != mode || got.Tagged != tagged || got.Branch != branch || got.Slug != slug {
		t.Fatalf("got mode=%q tagged=%v branch=%q slug=%q; want mode=%q tagged=%v branch=%q slug=%q",
			got.Mode, got.Tagged, got.Branch, got.Slug, mode, tagged, branch, slug)
	}
}

// T-F4-1: the primary branch named by origin/HEAD is never tagged.
func TestPrimaryBranchIsNotTagged(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "main", "")
}

// T-F4-1: origin/HEAD wins over the conventional names.
func TestOriginHeadDecidesThePrimaryBranch(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "develop")
	assertResult(t, detect(t, dir), ModeCloudBranch, true, "main", "main")
	git(t, dir, "switch", "-q", "-c", "develop")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "develop", "")
}

// T-F4-2: a feature branch is tagged with its slug.
func TestFeatureBranchIsTagged(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	git(t, dir, "switch", "-q", "-c", "feature/x")
	assertResult(t, detect(t, dir), ModeCloudBranch, true, "feature/x", "feature-x")
}

// T-F4-3: slug normalization table.
func TestSlugTable(t *testing.T) {
	cases := map[string]string{
		"feature/x":      "feature-x",
		"Feature/ABC":    "feature-abc",
		"fix_bug":        "fix-bug",
		"release.v1.2.3": "release-v1-2-3",
		"mi rama nueva":  "mi-rama-nueva",
		"feat/canción":   "feat-canci-n",
		"feat//x///y":    "feat-x-y",
		"-/x/-":          "x",
		"R-08":           "r-08",
	}
	for branch, want := range cases {
		got, err := Slug(branch)
		if err != nil || got != want {
			t.Errorf("Slug(%q) = %q, %v; want %q", branch, got, err, want)
		}
	}
}

// T-F4-3: a branch with nothing left after normalization is an error.
func TestSlugEmptyIsAnError(t *testing.T) {
	for _, branch := range []string{"", "ñ", "///", "_._"} {
		if got, err := Slug(branch); !errors.Is(err, ErrEmptySlug) {
			t.Errorf("Slug(%q) = %q, %v; want ErrEmptySlug", branch, got, err)
		}
	}
}

// T-F4-4: more than 40 characters → cut to 33, no trailing dash, plus a
// 6-hex SHA-256 suffix of the full branch name. Values match the R-01/R-05
// prototype (docs/research/r01-prototipo/hook.py, make_slug).
func TestSlugLongBranchIsCutAndSuffixed(t *testing.T) {
	cases := map[string]string{
		"feature/una-rama-con-un-nombre-muy-largo-para-probar": "feature-una-rama-con-un-nombre-mu-3c9b3b",
		"feature/abcdefghijklmnopqrstuvwxy-zzzzzzzzzzzzzzzzzz": "feature-abcdefghijklmnopqrstuvwxy-ed1788",
		// The cut ends in "-": the dash is dropped before the suffix.
		"feature/abcdefghijklmnopqrstuvwx-zzzzzzzzzzzzzzzzzzz": "feature-abcdefghijklmnopqrstuvwx-715897",
	}
	for branch, want := range cases {
		got, err := Slug(branch)
		if err != nil || got != want {
			t.Errorf("Slug(%q) = %q, %v; want %q", branch, got, err, want)
		}
		if len(got) > 40 {
			t.Errorf("Slug(%q) has %d characters; want at most 40", branch, len(got))
		}
	}
}

// T-F4-4: 40 characters exactly stay as they are.
func TestSlugAtFortyIsKept(t *testing.T) {
	branch := strings.Repeat("a", 40)
	if got, _ := Slug(branch); got != branch {
		t.Fatalf("Slug(40×a) = %q; want unchanged", got)
	}
}

// T-F4-4: branches sharing a long prefix do not collide, and the result is
// deterministic.
func TestSlugLongBranchesWithSamePrefixDoNotCollide(t *testing.T) {
	prefix := "feature/" + strings.Repeat("prefijo-comun-", 4)
	a, errA := Slug(prefix + "uno")
	b, errB := Slug(prefix + "dos")
	if errA != nil || errB != nil {
		t.Fatalf("errors: %v, %v", errA, errB)
	}
	if a == b {
		t.Fatalf("Slug collision: both %q", a)
	}
	again, _ := Slug(prefix + "uno")
	if again != a {
		t.Fatalf("Slug is not deterministic: %q then %q", a, again)
	}
}

// T-F4-4 (TBD): "feat/una" and "feat-una" still share a slug. The owner has
// not decided the fix yet (docs/equipo/fichas/R-08.md, "Colisión de slugs").
// This test pins the current, known behavior so the decision changes it on
// purpose: when the owner decides, invert this assertion.
func TestSlugShortCollisionIsPendingOwnerDecision(t *testing.T) {
	a, _ := Slug("feat/una")
	b, _ := Slug("feat-una")
	if a != b {
		t.Fatalf("the (TBD) collision changed without a recorded decision: %q vs %q", a, b)
	}
}

// T-F4-5: a detached HEAD is tagged as detached-<sha7>.
func TestDetachedHead(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	sha := git(t, dir, "rev-parse", "HEAD")
	git(t, dir, "switch", "-q", "--detach", "HEAD")
	want := "detached-" + sha[:7]
	assertResult(t, detect(t, dir), ModeCloudBranch, true, want, want)
}

// T-F4-6: outside a repository there is nothing to tag.
func TestNoRepository(t *testing.T) {
	isolateGit(t)
	assertResult(t, detect(t, t.TempDir()), ModeOff, false, "", "")
}

// T-F4-6: a repository without remotes is not tagged.
func TestNoRemoteIsNotTagged(t *testing.T) {
	dir := newRepo(t, true)
	git(t, dir, "switch", "-q", "-c", "feature/x")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "feature/x", "")
}

// T-F4-6: with a remote but no origin/HEAD, init.defaultBranch names the
// primary branch when that branch exists.
func TestNoOriginHeadUsesInitDefaultBranch(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "")
	git(t, dir, "config", "init.defaultBranch", "trunk")
	git(t, dir, "switch", "-q", "-c", "trunk")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "trunk", "")
	git(t, dir, "switch", "-q", "main")
	assertResult(t, detect(t, dir), ModeCloudBranch, true, "main", "main")
}

// T-F4-6: with no origin/HEAD and no init.defaultBranch, main or master is the
// primary branch.
func TestNoOriginHeadFallsBackToMainOrMaster(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "main", "")

	legacy := newRepo(t, true)
	git(t, legacy, "branch", "-q", "-m", "master")
	git(t, legacy, "remote", "add", "origin", "https://example.invalid/team/legacy.git")
	assertResult(t, detect(t, legacy), ModeCloudBranch, false, "master", "")
}

// T-F4-7: a linked worktree reports its own branch.
func TestLinkedWorktreeUsesItsBranch(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	wt := filepath.Join(filepath.Dir(dir), "wt")
	git(t, dir, "worktree", "add", "-q", "-b", "feature/wt", wt)
	assertResult(t, detect(t, wt), ModeCloudBranch, true, "feature/wt", "feature-wt")
	assertResult(t, detect(t, dir), ModeCloudBranch, false, "main", "")
}

// T-F4-8: no marker → off.
func TestMarkerAbsentIsOff(t *testing.T) {
	dir := newRepo(t, false)
	addOrigin(t, dir, "main")
	git(t, dir, "switch", "-q", "-c", "feature/x")
	res := detect(t, dir)
	assertResult(t, res, ModeOff, false, "feature/x", "")
	if len(res.Warnings) != 0 {
		t.Fatalf("absent marker must not warn, got %+v", res.Warnings)
	}
}

// T-F4-8: an invalid marker or an unknown policy → off with a warning.
func TestMarkerInvalidIsOffWithWarning(t *testing.T) {
	cases := map[string]string{
		`{not json`:                         WarnMarkerInvalid,
		`["cloud-branch"]`:                  WarnMarkerInvalid,
		`{"memory_policy":"cloud-forever"}`: WarnMarkerUnknownPolicy,
		`{}`:                                WarnMarkerUnknownPolicy,
	}
	for content, code := range cases {
		dir := newRepo(t, false)
		addOrigin(t, dir, "main")
		git(t, dir, "switch", "-q", "-c", "feature/x")
		writeMarker(t, dir, content)
		res := detect(t, dir)
		assertResult(t, res, ModeOff, false, "feature/x", "")
		if len(res.Warnings) != 1 || res.Warnings[0].Code != code {
			t.Errorf("marker %q: warnings = %+v; want one %q", content, res.Warnings, code)
		}
	}
}

// T-F4-8: the marker is read from the repository root, also from a subfolder.
func TestMarkerIsReadFromRepositoryRoot(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	git(t, dir, "switch", "-q", "-c", "feature/x")
	sub := filepath.Join(dir, "docs", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	assertResult(t, detect(t, sub), ModeCloudBranch, true, "feature/x", "feature-x")
}

// T-F4-10: ENGRAM_* variables do not change the result.
func TestEngramEnvironmentIsIgnored(t *testing.T) {
	dir := newRepo(t, true)
	addOrigin(t, dir, "main")
	git(t, dir, "switch", "-q", "-c", "feature/x")
	before := detect(t, dir)
	t.Setenv("ENGRAM_PROJECT", "otro")
	t.Setenv("ENGRAM_CLOUD_TOKEN", "x")
	t.Setenv("ENGRAM_REMOTE", "https://example.invalid")
	after := detect(t, dir)
	assertResult(t, after, before.Mode, before.Tagged, before.Branch, before.Slug)
}

// T-F4-11: paths with spaces, and a marker written on Windows (CRLF, BOM).
func TestWindowsPathsAndMarkerLineEndings(t *testing.T) {
	isolateGit(t)
	dir := filepath.Join(t.TempDir(), "carpeta con espacios", "mi repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-q", "-b", "main")
	writeMarker(t, dir, "\xef\xbb\xbf{\r\n  \"memory_policy\": \"cloud-branch\"\r\n}\r\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "init")
	addOrigin(t, dir, "main")
	git(t, dir, "switch", "-q", "-c", "feature/x")
	res := detect(t, dir)
	assertResult(t, res, ModeCloudBranch, true, "feature/x", "feature-x")
	if len(res.Warnings) != 0 {
		t.Fatalf("CRLF/BOM marker must be valid, got %+v", res.Warnings)
	}
}

// A directory that does not exist is an error, not an "off" result.
func TestMissingDirectoryIsAnError(t *testing.T) {
	if _, err := Detect(filepath.Join(t.TempDir(), "no-existe")); !errors.Is(err, ErrNoDirectory) {
		t.Fatalf("Detect(missing) error = %v; want ErrNoDirectory", err)
	}
}
