// Package memory decides how Engram memory is tagged for the current Git
// branch ([EN] Fase 4). It is pure local logic: it runs read-only Git
// commands, reads the project marker .engram/active-stack.json and never uses
// the network or ENGRAM_* variables.
//
// The rules are ported from the prototype validated in R-01 and R-05
// (docs/research/r01-prototipo/hook.py).
package memory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Memory modes.
const (
	ModeOff         = "off"
	ModeCloudBranch = "cloud-branch"
)

// Warning codes.
const (
	WarnMarkerInvalid       = "marker_invalid"
	WarnMarkerUnknownPolicy = "marker_unknown_policy"
)

// MarkerPath is the marker location relative to the repository root.
var MarkerPath = filepath.Join(".engram", "active-stack.json")

const (
	maxSlug      = 40
	cutSlug      = 33
	suffixHexLen = 6
	gitTimeout   = 5 * time.Second
)

var (
	// ErrEmptySlug means the branch name has no letter or digit left.
	ErrEmptySlug = errors.New("branch name has no usable characters for a slug")
	// ErrNoDirectory means the requested directory does not exist.
	ErrNoDirectory = errors.New("directory does not exist")
	// ErrGitMissing means the git executable is not available.
	ErrGitMissing = errors.New("git is not installed or not in PATH")
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Warning reports a marker that was ignored.
type Warning struct {
	Code   string
	Path   string
	Policy string
}

// Result is what Detect decides for a directory. Branch is empty outside a
// repository; Slug is set only when Tagged.
type Result struct {
	Mode     string
	Tagged   bool
	Branch   string
	Slug     string
	Warnings []Warning
}

// Slug turns a branch name into a tag: lower case, every run of characters
// outside [a-z0-9] becomes "-", no dashes at the ends. When the slug loses
// information with respect to the branch name (it had "/", "_", capitals or
// other characters that were rewritten) or is longer than 40 characters, it
// keeps at most its first 33 characters (without a trailing dash) plus "-" and
// the first 6 hex digits of the SHA-256 of the full branch name, so "feat/una"
// and "feat-una" do not share a tag. A branch already written as a slug is
// returned as is.
func Slug(branch string) (string, error) {
	slug := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(branch), "-"), "-")
	if slug == "" {
		return "", ErrEmptySlug
	}
	if slug != branch || len(slug) > maxSlug {
		sum := sha256.Sum256([]byte(branch))
		if len(slug) > cutSlug {
			slug = strings.TrimRight(slug[:cutSlug], "-")
		}
		slug = slug + "-" + hex.EncodeToString(sum[:])[:suffixHexLen]
	}
	return slug, nil
}

// Detect decides the memory mode and tag for dir:
//   - outside a repository: mode off, no branch;
//   - without a valid marker: mode off (with a warning if the marker is
//     invalid or names an unknown policy);
//   - without remotes, or on the primary branch: not tagged;
//   - otherwise tagged with the branch slug (detached HEAD: detached-<sha7>).
func Detect(dir string) (Result, error) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return Result{}, ErrNoDirectory
	}
	g := gitRunner{dir: dir}

	top, err := g.run("rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, ErrGitMissing) {
			return Result{}, err
		}
		return Result{Mode: ModeOff}, nil
	}

	branch, err := currentBranch(g)
	if err != nil {
		return Result{}, err
	}

	res := Result{Mode: ModeOff, Branch: branch}
	mode, warning := readMarker(filepath.Join(filepath.FromSlash(top), MarkerPath))
	if warning != nil {
		res.Warnings = append(res.Warnings, *warning)
	}
	if mode != ModeCloudBranch {
		return res, nil
	}
	res.Mode = mode

	remotes, err := g.run("remote")
	if err != nil {
		return Result{}, err
	}
	if remotes == "" || branch == primaryBranch(g) {
		return res, nil
	}

	slug, err := Slug(branch)
	if err != nil {
		// Keep mode and branch so the caller can name the branch in its message.
		return res, err
	}
	res.Tagged = true
	res.Slug = slug
	return res, nil
}

func currentBranch(g gitRunner) (string, error) {
	branch, err := g.run("branch", "--show-current")
	if err != nil {
		return "", err
	}
	if branch != "" {
		return branch, nil
	}
	sha, err := g.run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	if len(sha) < 7 {
		return "", errors.New("unexpected HEAD object name: " + sha)
	}
	return "detached-" + sha[:7], nil
}

// primaryBranch is origin/HEAD when it exists. Otherwise it is the first of
// init.defaultBranch, main and master that exists locally or on origin; if
// none exists, the first candidate.
func primaryBranch(g gitRunner) string {
	if head, err := g.run("symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil && head != "" {
		return strings.TrimPrefix(head, "origin/")
	}
	var candidates []string
	if def, err := g.run("config", "init.defaultBranch"); err == nil && def != "" {
		candidates = append(candidates, def)
	}
	candidates = append(candidates, "main", "master")
	for _, name := range candidates {
		for _, ref := range []string{"refs/heads/" + name, "refs/remotes/origin/" + name} {
			if _, err := g.run("show-ref", "--verify", "--quiet", ref); err == nil {
				return name
			}
		}
	}
	return candidates[0]
}

// readMarker returns the policy named by the marker, or "" when the marker is
// absent or ignored. A marker written on Windows (CRLF, UTF-8 BOM) is valid.
func readMarker(path string) (string, *Warning) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	var marker struct {
		MemoryPolicy string `json:"memory_policy"`
	}
	if err := json.Unmarshal(raw, &marker); err != nil {
		return "", &Warning{Code: WarnMarkerInvalid, Path: path}
	}
	if marker.MemoryPolicy != ModeCloudBranch {
		return "", &Warning{Code: WarnMarkerUnknownPolicy, Path: path, Policy: marker.MemoryPolicy}
	}
	return ModeCloudBranch, nil
}

// gitRunner runs read-only, local Git commands in dir.
type gitRunner struct{ dir string }

func (g gitRunner) run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", g.dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrGitMissing
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
