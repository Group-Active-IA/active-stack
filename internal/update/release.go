package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	defaultAPIBase      = "https://api.github.com"
	defaultDownloadBase = "https://github.com"
	maxBinaryBytes      = 150 << 20
)

var releaseTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)

type Release struct{ Tag string }

// ReleaseClient reads only the official Active Stack release endpoint. Base
// overrides exist for isolated tests; production uses HTTPS GitHub endpoints.
type ReleaseClient struct {
	HTTP         *http.Client
	APIBase      string
	DownloadBase string
}

func (c ReleaseClient) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 2 * time.Minute}
}

func (c ReleaseClient) apiBase() string {
	if c.APIBase != "" {
		return strings.TrimRight(c.APIBase, "/")
	}
	return defaultAPIBase
}

func (c ReleaseClient) downloadBase() string {
	if c.DownloadBase != "" {
		return strings.TrimRight(c.DownloadBase, "/")
	}
	return defaultDownloadBase
}

func (c ReleaseClient) get(ctx context.Context, url string, max int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "active-stack-update")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("download exceeds %d bytes", max)
	}
	return data, nil
}

// Latest resolves the newest published release tag.
func (c ReleaseClient) Latest(ctx context.Context) (Release, error) {
	data, err := c.get(ctx, c.apiBase()+"/repos/Group-Active-IA/active-stack/releases/latest", 1<<20)
	if err != nil {
		return Release{}, err
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return Release{}, fmt.Errorf("parse release: %w", err)
	}
	if !releaseTag.MatchString(release.TagName) {
		return Release{}, fmt.Errorf("invalid release tag %q", release.TagName)
	}
	return Release{Tag: release.TagName}, nil
}

func assetName(goos, goarch string) (string, error) {
	if goarch != "amd64" || (goos != "windows" && goos != "linux") {
		return "", fmt.Errorf("no release binary for %s/%s", goos, goarch)
	}
	name := "active-stack_" + goos + "_" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

// ReleaseBinaryAvailable reports whether the release pipeline publishes a CLI
// executable for this platform. Other platforms can still refresh skills.
func ReleaseBinaryAvailable(goos, goarch string) bool {
	_, err := assetName(goos, goarch)
	return err == nil
}

// DownloadBinary verifies the published SHA256SUMS entry before making the
// payload available to the caller. The caller owns the temporary file.
func (c ReleaseClient) DownloadBinary(ctx context.Context, release Release, goos, goarch, dir string) (string, error) {
	if !releaseTag.MatchString(release.Tag) {
		return "", fmt.Errorf("invalid release tag %q", release.Tag)
	}
	name, err := assetName(goos, goarch)
	if err != nil {
		return "", err
	}
	base := c.downloadBase() + "/Group-Active-IA/active-stack/releases/download/" + release.Tag + "/"
	checksums, err := c.get(ctx, base+"SHA256SUMS.txt", 1<<20)
	if err != nil {
		return "", err
	}
	var expected string
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			expected = fields[0]
			break
		}
	}
	if len(expected) != 64 {
		return "", fmt.Errorf("missing SHA256 for %s", name)
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return "", fmt.Errorf("invalid SHA256 for %s: %w", name, err)
	}
	data, err := c.get(ctx, base+name, maxBinaryBytes)
	if err != nil {
		return "", err
	}
	actual := sha256.Sum256(data)
	if hex.EncodeToString(actual[:]) != strings.ToLower(expected) {
		return "", fmt.Errorf("SHA256 mismatch for %s", name)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	pattern := "active-stack-verified-*"
	if goos == "windows" {
		pattern += ".exe"
	}
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	if err := file.Chmod(0o755); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return filepath.Clean(path), nil
}
