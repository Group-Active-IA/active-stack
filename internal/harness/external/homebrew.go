package external

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// httpClient is used for GitHub API and asset downloads; replaceable in tests.
var httpClient = &http.Client{Timeout: 5 * time.Minute}

// githubBaseURL is replaceable in tests to point at a mock server.
var githubBaseURL = "https://github.com"

// binaryInstallDirFn resolves where binaries are installed; replaceable in tests.
var binaryInstallDirFn = binaryInstallDir

func installHomebrew(ctx context.Context, h model.Harness, profile system.PlatformProfile) (Result, error) {
	pkg := h.External.Pkg
	if pkg == "" {
		return Result{}, fmt.Errorf("harness %q: homebrew method requires External.Pkg", h.ID)
	}

	// Use brew when it is available (macOS native, or Linuxbrew).
	if profile.PackageManager == "brew" {
		return runBrew(ctx, pkg)
	}

	// Fallback: download binary from GitHub Releases.
	binaryPath, err := downloadBinary(ctx, h, profile)
	if err != nil {
		return Result{}, err
	}
	return Result{BinaryPath: binaryPath}, nil
}

func runBrew(ctx context.Context, pkg string) (Result, error) {
	if out, err := runner(ctx, "brew", "install", pkg); err != nil {
		return Result{}, fmt.Errorf("brew install %s: %w\n%s", pkg, err, out)
	}

	binaryName := filepath.Base(pkg)
	binaryPath, _ := lookPath(binaryName)
	return Result{BinaryPath: binaryPath}, nil
}

// downloadBinary fetches the latest GitHub Release binary for the harness.
// The download source is h.External.Repo (owner/repo) when set, otherwise it
// falls back to h.External.Pkg — this keeps the brew formula (Pkg) separate
// from the GitHub repo (Repo), e.g. brew "engram" vs repo
// "Gentleman-Programming/engram". h.External.URL may override the GitHub base.
func downloadBinary(ctx context.Context, h model.Harness, profile system.PlatformProfile) (string, error) {
	source := h.External.Repo
	if source == "" {
		source = h.External.Pkg
	}
	owner, repo := resolveOwnerRepo(source)
	binaryName := repo
	if profile.OS == "windows" {
		binaryName = repo + ".exe"
	}

	baseURL := githubBaseURL
	if h.External.URL != "" && !strings.HasPrefix(h.External.URL, "https://mcp.") {
		// URL field overrides GitHub base only for download harnesses, not mcp.
		baseURL = h.External.URL
	}

	version, err := fetchLatestVersion(owner, repo, baseURL)
	if err != nil {
		return "", fmt.Errorf("fetch latest version for %s/%s: %w", owner, repo, err)
	}

	goos := profile.OS
	goarch := normalizeArch(runtime.GOARCH)
	assetURL := buildAssetURL(baseURL, owner, repo, version, goos, goarch)

	installDir := binaryInstallDirFn(goos)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create install dir %q: %w", installDir, err)
	}

	outPath := filepath.Join(installDir, binaryName)

	if strings.HasSuffix(assetURL, ".zip") {
		if err := downloadAndExtractZip(assetURL, binaryName, outPath); err != nil {
			return "", fmt.Errorf("download %s zip: %w", repo, err)
		}
	} else {
		if err := downloadAndExtractTarGz(assetURL, repo, outPath); err != nil {
			return "", fmt.Errorf("download %s tar.gz: %w", repo, err)
		}
	}

	return outPath, nil
}

// resolveOwnerRepo parses "owner/repo" or returns "<pkg>/<pkg>" for bare names.
func resolveOwnerRepo(pkg string) (owner, repo string) {
	if strings.Contains(pkg, "/") {
		parts := strings.SplitN(pkg, "/", 2)
		return parts[0], parts[1]
	}
	return pkg, pkg
}

// normalizeArch maps GOARCH to the naming convention used in most GitHub release assets.
func normalizeArch(goarch string) string {
	switch goarch {
	case "386":
		return "amd64"
	case "arm":
		return "arm64"
	default:
		return goarch
	}
}

// buildAssetURL constructs the GitHub Releases download URL for a binary asset.
func buildAssetURL(baseURL, owner, repo, version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	filename := fmt.Sprintf("%s_%s_%s_%s%s", repo, version, goos, goarch, ext)
	return fmt.Sprintf("%s/%s/%s/releases/download/v%s/%s", baseURL, owner, repo, version, filename)
}

// apiBaseURL returns the GitHub API base URL. When baseURL is a local test
// server it is used directly; otherwise the real API host is returned.
func apiBaseURL(baseURL string) string {
	if strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "localhost") {
		return baseURL
	}
	return "https://api.github.com"
}

// fetchLatestVersion queries the GitHub Releases API for the latest tag.
func fetchLatestVersion(owner, repo, baseURL string) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBaseURL(baseURL), owner, repo)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if tok := githubToken(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode release JSON: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		return "", fmt.Errorf("empty tag_name in GitHub release response")
	}
	return version, nil
}

// githubToken returns a GitHub token from the environment if available.
func githubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

// binaryInstallDir returns the directory where extracted binaries should land.
func binaryInstallDir(goos string) string {
	if goos == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "jr-stack", "bin")
	}

	candidate := "/usr/local/bin"
	if isWritableDir(candidate) {
		return candidate
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin"
	}
	return filepath.Join(home, ".local", "bin")
}

func isWritableDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	tmp, err := os.CreateTemp(dir, ".jr-stack-write-test-*")
	if err != nil {
		return false
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return true
}

func downloadAndExtractTarGz(url, binaryName, outPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	return extractBinaryFromTarGz(resp.Body, binaryName, outPath)
}

func extractBinaryFromTarGz(r io.Reader, binaryName, outPath string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag != tar.TypeDir {
			return writeExecutable(tr, outPath)
		}
	}
	return fmt.Errorf("binary %q not found in archive", binaryName)
}

func downloadAndExtractZip(url, binaryName, outPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	zr, err := zip.NewReader(&byteReaderAt{data: data}, int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range zr.File {
		if filepath.Base(f.Name) == binaryName && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("open zip entry %q: %w", f.Name, err)
			}
			defer rc.Close()
			return writeExecutable(rc, outPath)
		}
	}
	return fmt.Errorf("binary %q not found in zip archive", binaryName)
}

type byteReaderAt struct{ data []byte }

func (b *byteReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || int(off) >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func writeExecutable(r io.Reader, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}
