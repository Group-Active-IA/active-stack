package projectinit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Git is used only for local reads: rev-parse, config, status, ls-files and
// check-ignore. It never runs add, commit, push, fetch or anything that
// changes the repository or the network (T-F7-8, T-F7-10).

const gitTimeout = 10 * time.Second

// gitOut runs git in dir and returns stdout and the exit code (-1 when git
// could not run at all).
func gitOut(dir string, args ...string) (stdout string, code int, err error) {
	if _, lerr := exec.LookPath("git"); lerr != nil {
		return "", -1, ErrGitMissing
	}
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	full := append([]string{"-C", dir, "-c", "core.quotepath=off"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	// read-only commands must not take the index lock or ask for anything
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	runErr := cmd.Run()
	if runErr == nil {
		return out.String(), 0, nil
	}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		return out.String(), ee.ExitCode(), nil
	}
	return "", -1, fmt.Errorf("git %s: %w", args[0], runErr)
}

// gitOK runs git and requires exit code 0.
func gitOK(dir string, args ...string) (string, error) {
	out, code, err := gitOut(dir, args...)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return out, fmt.Errorf("git %s exited with %d", args[0], code)
	}
	return out, nil
}

// resolveRoot returns the work tree root of the repository that holds dir.
func resolveRoot(dir string) (string, error) {
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return "", fmt.Errorf("the directory %s does not exist", dir)
	}
	out, code, err := gitOut(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	if code != 0 || strings.TrimSpace(out) == "" {
		return "", ErrNotGitRepo
	}
	return filepath.Clean(filepath.FromSlash(strings.TrimSpace(out))), nil
}

// splitNUL splits git -z output.
func splitNUL(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// statusEntries returns, for the given paths, the porcelain status code and
// path of every entry git reports (untracked files are listed one by one).
func statusEntries(root string, paths []string) (map[string]string, error) {
	args := append([]string{"status", "--porcelain", "-z", "-uall", "--"}, paths...)
	out, err := gitOK(root, args...)
	if err != nil {
		return nil, err
	}
	res := map[string]string{}
	for _, e := range splitNUL(out) {
		if len(e) < 4 {
			continue
		}
		res[e[3:]] = e[:2]
	}
	return res, nil
}

// dirtyFiles are the given paths that are tracked and have uncommitted changes.
func dirtyFiles(root string, paths []string) ([]string, error) {
	st, err := statusEntries(root, paths)
	if err != nil {
		return nil, err
	}
	var out []string
	for p, code := range st {
		if code != "??" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}

// pendingFiles are the given paths git still shows as changed or untracked.
func pendingFiles(root string, paths []string) ([]string, error) {
	st, err := statusEntries(root, paths)
	if err != nil {
		return nil, err
	}
	var out []string
	for p := range st {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// trackedOther lists versioned files under .engram/ that are not ours.
func trackedOther(root string) ([]string, error) {
	out, err := gitOK(root, "ls-files", "-z", "--", ".engram")
	if err != nil {
		return nil, err
	}
	var other []string
	for _, p := range splitNUL(out) {
		if p != ConfigPath && p != MarkerPath {
			other = append(other, p)
		}
	}
	sort.Strings(other)
	return other, nil
}

// isIgnored asks git, ignoring the index, whether the rules ignore path.
func isIgnored(root, path string) (bool, error) {
	_, code, err := gitOut(root, "check-ignore", "-q", "--no-index", "--", path)
	if err != nil {
		return false, err
	}
	switch code {
	case 0:
		return true, nil
	case 1:
		return false, nil
	default:
		return false, fmt.Errorf("git check-ignore exited with %d", code)
	}
}
