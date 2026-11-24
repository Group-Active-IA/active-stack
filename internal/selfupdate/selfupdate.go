// Package selfupdate warns when a newer active-stack release exists and, only
// when the person asks for it, replaces the running binary (R-27).
//
// Governance ALTO: it downloads and replaces an executable. Nothing changes
// without an explicit `self-update`; the SHA256 of the download is checked
// against the release's SHA256SUMS.txt before the binary is touched, the
// current binary is backed up first, and a failed replacement restores it.
package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/backup"
)

// APIBase is the release API root. Tests point it at a fake server.
var APIBase = "https://api.github.com/repos/Group-Active-IA/active-stack"

// Now is the clock (tests).
var Now = time.Now

// CheckInterval is how often the network is consulted.
const CheckInterval = 24 * time.Hour

// Timeout bounds every network call of the startup check.
const Timeout = 2 * time.Second

const sumsAsset = "SHA256SUMS.txt"

// ErrChecksum means the download does not match the published SHA256.
var ErrChecksum = errors.New("the downloaded binary does not match the published SHA256; the current binary was not touched")

type release struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

type cacheFile struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// AssetName is the release asset for a platform ("" when none is published).
func AssetName(goos, goarch string) string {
	switch {
	case goos == "windows" && goarch == "amd64":
		return "active-stack_windows_amd64.exe"
	case goos == "linux" && goarch == "amd64":
		return "active-stack_linux_amd64"
	case goos == "darwin" && (goarch == "amd64" || goarch == "arm64"):
		return "active-stack_darwin_" + goarch
	}
	return ""
}

// parse reads "v1.2.3" (a "-suffix" is ignored). ok is false for "dev" or junk.
func parse(v string) (n [3]int, ok bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return n, false
	}
	for i, p := range parts {
		x, err := strconv.Atoi(p)
		if err != nil || x < 0 {
			return n, false
		}
		n[i] = x
	}
	return n, true
}

// Newer reports whether latest is a higher version than current. A current
// version that is not a release (a "dev" build) never counts as outdated.
func Newer(current, latest string) bool {
	c, ok1 := parse(current)
	l, ok2 := parse(latest)
	if !ok1 || !ok2 {
		return false
	}
	for i := range c {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func cachePath(home string) string { return filepath.Join(home, ".active-stack", "update-check.json") }

func get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "active-stack-self-update")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

func latestRelease(ctx context.Context) (release, error) {
	var r release
	b, err := get(ctx, APIBase+"/releases/latest", 1<<20)
	if err != nil {
		return r, err
	}
	return r, json.Unmarshal(b, &r)
}

// Check returns the newest release tag when it is higher than current, or "".
// It never fails loudly: without network, with a fresh cache or with a "dev"
// build it returns "" and nil. The network is consulted at most once per
// CheckInterval and only for the bounded Timeout.
func Check(ctx context.Context, home, current string) string {
	if _, ok := parse(current); !ok {
		return ""
	}
	var c cacheFile
	if b, err := os.ReadFile(cachePath(home)); err == nil && json.Unmarshal(b, &c) == nil {
		if Now().Sub(c.CheckedAt) < CheckInterval {
			if Newer(current, c.Latest) {
				return c.Latest
			}
			return ""
		}
	}
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	r, err := latestRelease(ctx)
	if err != nil {
		return ""
	}
	if b, err := json.Marshal(cacheFile{CheckedAt: Now(), Latest: r.Tag}); err == nil {
		_ = os.MkdirAll(filepath.Dir(cachePath(home)), 0o755)
		_ = os.WriteFile(cachePath(home), b, 0o644)
	}
	if Newer(current, r.Tag) {
		return r.Tag
	}
	return ""
}

var snapshotCreate = func(dir string, paths []string) error {
	_, err := backup.NewSnapshotter().Create(dir, paths)
	return err
}

// SetSnapshotCreate replaces the snapshot function (tests).
func SetSnapshotCreate(fn func(string, []string) error) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

// Apply downloads the release asset for goos/goarch, verifies its SHA256 and
// replaces exe with it. exe is left untouched on any failure.
func Apply(ctx context.Context, home, exe, goos, goarch string) (tag string, err error) {
	name := AssetName(goos, goarch)
	if name == "" {
		return "", fmt.Errorf("no release binary is published for %s/%s", goos, goarch)
	}
	r, err := latestRelease(ctx)
	if err != nil {
		return "", fmt.Errorf("query the latest release: %w", err)
	}
	var binURL, sumsURL string
	for _, a := range r.Assets {
		switch a.Name {
		case name:
			binURL = a.URL
		case sumsAsset:
			sumsURL = a.URL
		}
	}
	if binURL == "" || sumsURL == "" {
		return "", fmt.Errorf("release %s lacks %s or %s", r.Tag, name, sumsAsset)
	}
	sums, err := get(ctx, sumsURL, 1<<20)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", sumsAsset, err)
	}
	want := ""
	for _, line := range strings.Split(string(sums), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name {
			want = strings.ToLower(f[0])
		}
	}
	if want == "" {
		return "", fmt.Errorf("%s has no entry for %s", sumsAsset, name)
	}
	data, err := get(ctx, binURL, 200<<20)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", name, err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != want {
		return "", ErrChecksum
	}

	dir := filepath.Join(home, ".active-stack", "backups", "self-update")
	if err := snapshotCreate(dir, []string{exe}); err != nil {
		return "", fmt.Errorf("backup failed, the binary was not touched: %w", err)
	}
	if err := replace(exe, data); err != nil {
		return "", err
	}
	_ = os.Remove(cachePath(home))
	return r.Tag, nil
}

// replace writes data next to exe and swaps it in (rename-then-replace, safe
// for a running .exe on Windows), restoring the old file if the swap fails.
func replace(exe string, data []byte) error {
	tmp, old := exe+".new", exe+".old"
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("move the current binary aside: %w", err)
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Rename(old, exe)
		_ = os.Remove(tmp)
		return fmt.Errorf("put the new binary in place (the previous one was restored): %w", err)
	}
	return nil
}
