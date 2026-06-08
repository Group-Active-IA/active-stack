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
// The default delegates to system.BinaryInstallDir (D1 — shared home).
var binaryInstallDirFn = system.BinaryInstallDir

// addToUserPath persists the install dir onto the user PATH; replaceable in tests.
var addToUserPath = system.AddToUserPath

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

	// Resolve goos early so it is used consistently for binaryName, the asset
	// URL, and the install directory. Falling back to runtime.GOOS when
	// profile.OS is empty avoids the double-underscore filename bug
	// ("engram_1.16.1__amd64.tar.gz") that produced HTTP 404 on GitHub Releases.
	goos := profile.OS
	if goos == "" {
		goos = runtime.GOOS
	}

	binaryName := repo
	if goos == "windows" {
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

	// El binario quedó en installDir — en Windows eso es %LOCALAPPDATA%\jr-stack\bin,
	// que NO está en el PATH del usuario por defecto. Lo agregamos al PATH de usuario
	// (persistente vía registro, sin admin; fuera de Windows solo toca el PATH del
	// proceso). Best-effort a propósito: si falla (p. ej. ExecutionPolicy de
	// PowerShell) NO abortamos — el binario ya está en disco y el health check de
	// PATH es Soft (verify.checkBinaryInPATH), que reportará el warning si quedó
	// inaccesible.
	_ = addToUserPath(installDir)

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
// If goos is empty it falls back to runtime.GOOS to guarantee a well-formed
// filename (an empty goos would produce a double-underscore like
// "engram_1.16.1__amd64.tar.gz", causing an HTTP 404 on GitHub Releases).
func buildAssetURL(baseURL, owner, repo, version, goos, goarch string) string {
	if goos == "" {
		goos = runtime.GOOS
	}
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
	f, err := openTruncExecutable(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// openTruncExecutable opens outPath for writing, truncating any prior content.
// If the target already exists but cannot be truncated in place — the classic
// Windows reinstall case where the old binary is still running and therefore
// locked ("being used by another process") — it moves the stale binary aside to
// "<path>.old" and retries. Windows forbids truncating or deleting a running
// .exe but DOES allow renaming it, so the old process keeps running under the
// renamed file while the new binary lands in place. This is the standard
// self-update pattern; on Unix the rename path is a harmless fallback.
func openTruncExecutable(outPath string) (*os.File, error) {
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err == nil {
		return f, nil
	}
	// Only recoverable when the file actually exists; otherwise the failure is
	// something else (e.g. an unwritable parent dir) and we must not mask it.
	if _, statErr := os.Stat(outPath); statErr != nil {
		return nil, fmt.Errorf("create %s: %w", outPath, err)
	}
	aside := outPath + ".old"
	_ = os.Remove(aside) // clear a leftover from a prior replace; ignore if absent/locked
	if mvErr := os.Rename(outPath, aside); mvErr != nil {
		return nil, fmt.Errorf("create %s: %w", outPath, err)
	}
	f, retryErr := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if retryErr != nil {
		return nil, fmt.Errorf("create %s: %w", outPath, retryErr)
	}
	return f, nil
}
