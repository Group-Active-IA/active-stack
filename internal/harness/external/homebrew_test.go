package external

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// ── resolveOwnerRepo ───────────────────────────────────────────────────────

func TestResolveOwnerRepo(t *testing.T) {
	tests := []struct {
		pkg       string
		wantOwner string
		wantRepo  string
	}{
		{"engram", "engram", "engram"},
		{"Gentleman-Programming/engram", "Gentleman-Programming", "engram"},
		{"owner/repo", "owner", "repo"},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			owner, repo := resolveOwnerRepo(tt.pkg)
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("resolveOwnerRepo(%q) = (%q, %q), want (%q, %q)",
					tt.pkg, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

// ── normalizeArch ─────────────────────────────────────────────────────────

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "amd64"},
		{"arm64", "arm64"},
		{"386", "amd64"},
		{"arm", "arm64"},
	}
	for _, tt := range tests {
		t.Run(tt.goarch, func(t *testing.T) {
			got := normalizeArch(tt.goarch)
			if got != tt.want {
				t.Errorf("normalizeArch(%q) = %q, want %q", tt.goarch, got, tt.want)
			}
		})
	}
}

// ── buildAssetURL ─────────────────────────────────────────────────────────

func TestBuildAssetURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		owner       string
		repo        string
		version     string
		goos        string
		goarch      string
		wantURL     string
		wantSuffix  string // checked when wantURL is empty
		wantNoDouble bool  // true: assert no double-underscore in filename
	}{
		{
			name:    "linux amd64 tar.gz",
			baseURL: "https://github.com",
			owner:   "Gentleman-Programming",
			repo:    "engram",
			version: "1.16.1",
			goos:    "linux",
			goarch:  "amd64",
			wantURL: "https://github.com/Gentleman-Programming/engram/releases/download/v1.16.1/engram_1.16.1_linux_amd64.tar.gz",
		},
		{
			name:    "windows amd64 zip",
			baseURL: "https://github.com",
			owner:   "Gentleman-Programming",
			repo:    "engram",
			version: "1.16.1",
			goos:    "windows",
			goarch:  "amd64",
			wantURL: "https://github.com/Gentleman-Programming/engram/releases/download/v1.16.1/engram_1.16.1_windows_amd64.zip",
		},
		{
			name:    "darwin arm64 tar.gz",
			baseURL: "https://github.com",
			owner:   "Owner",
			repo:    "repo",
			version: "1.0.0",
			goos:    "darwin",
			goarch:  "arm64",
			wantURL: "https://github.com/Owner/repo/releases/download/v1.0.0/repo_1.0.0_darwin_arm64.tar.gz",
		},
		{
			// Regression: empty goos must NOT produce a double-underscore filename
			// (e.g. "engram_1.16.1__amd64.tar.gz"). This reproduced the real HTTP 404.
			name:         "empty goos must not produce double underscore",
			baseURL:      "https://github.com",
			owner:        "Gentleman-Programming",
			repo:         "engram",
			version:      "1.16.1",
			goos:         "", // zero-value — the broken case
			goarch:       "amd64",
			wantNoDouble: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAssetURL(tt.baseURL, tt.owner, tt.repo, tt.version, tt.goos, tt.goarch)

			if tt.wantURL != "" && got != tt.wantURL {
				t.Errorf("buildAssetURL = %q\nwant           %q", got, tt.wantURL)
			}

			if tt.wantSuffix != "" && !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("URL %q should end in %q", got, tt.wantSuffix)
			}

			if tt.wantNoDouble && strings.Contains(got, "__") {
				t.Errorf("URL contains double-underscore (goos was empty): %q", got)
			}
		})
	}
}

// TestDownloadBinary_EmptyProfileOS is a regression test for the real-world HTTP 404:
//
//	https://github.com/…/engram_1.16.1__amd64.tar.gz  ← double underscore, no goos
//
// Root cause: internal/install/steps.go passed system.PlatformProfile{} (zero-value),
// so profile.OS == "" in downloadBinary, which caused buildAssetURL to build a
// malformed filename. After the fix, goos must fall back to runtime.GOOS so the
// URL is always well-formed.
func TestDownloadBinary_EmptyProfileOS(t *testing.T) {
	const binaryContent = "fake-engram-binary"
	const version = "1.16.1"

	// The mock server must serve the archive format that matches the runtime OS
	// because after the fix, an empty profile.OS falls back to runtime.GOOS.
	// On Windows this produces a .zip URL; on other platforms it produces .tar.gz.
	var archiveData []byte
	binaryFilename := "engram"
	if runtime.GOOS == "windows" {
		binaryFilename = "engram.exe"
		archiveData = buildZip(t, binaryFilename, []byte(binaryContent))
	} else {
		archiveData = buildTarGz(t, binaryFilename, []byte(binaryContent))
	}

	var gotAssetPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v" + version})
			return
		}
		gotAssetPath = r.URL.Path
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(archiveData)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	installDir := t.TempDir()
	origFn := binaryInstallDirFn
	binaryInstallDirFn = func(string) string { return installDir }
	defer func() { binaryInstallDirFn = origFn }()

	h := harnessWithMethod("homebrew", "engram", "")
	h.External.Repo = "Gentleman-Programming/engram"

	// Pass a zero-value profile (OS == "") — this is exactly what the install
	// pipeline was doing in externalStep.Run() before the fix.
	profile := system.PlatformProfile{} // OS intentionally empty

	_, err := downloadBinary(nil, h, profile)
	if err != nil {
		t.Fatalf("downloadBinary failed: %v", err)
	}

	// The asset URL must NOT contain a double underscore in the filename segment.
	if strings.Contains(gotAssetPath, "__") {
		t.Errorf("asset path contains double-underscore (goos was empty/not resolved): %q", gotAssetPath)
	}

	// The filename segment must contain a non-empty goos token between the version and arch.
	// Expected pattern: /engram_1.16.1_<goos>_<arch>.<ext>
	parts := strings.Split(gotAssetPath, "/")
	filename := parts[len(parts)-1] // e.g. "engram_1.16.1_linux_amd64.tar.gz"
	// Strip known suffixes before splitting on "_".
	baseName := strings.TrimSuffix(strings.TrimSuffix(filename, ".tar.gz"), ".zip")
	segments := strings.Split(baseName, "_")
	// segments: ["engram", "1.16.1", "<goos>", "<arch>"]
	if len(segments) < 4 || segments[2] == "" {
		t.Errorf("filename %q is missing the goos segment; got segments: %v", filename, segments)
	}
}

// ── downloadBinary via mock HTTP server (tar.gz) ───────────────────────────

func TestDownloadBinary_TarGz(t *testing.T) {
	const binaryContent = "fake-engram-binary"
	const version = "1.2.3"

	tarGzData := buildTarGz(t, "engram", []byte(binaryContent))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v" + version})
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(tarGzData)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	installDir := t.TempDir()
	origFn := binaryInstallDirFn
	binaryInstallDirFn = func(string) string { return installDir }
	defer func() { binaryInstallDirFn = origFn }()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	outPath, err := downloadBinary(nil, h, profile)
	if err != nil {
		t.Fatalf("downloadBinary failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != binaryContent {
		t.Errorf("binary content = %q, want %q", data, binaryContent)
	}
}

func TestDownloadBinary_Zip(t *testing.T) {
	const binaryContent = "fake-tool-binary"
	const version = "0.9.1"

	zipData := buildZip(t, "engram.exe", []byte(binaryContent))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v" + version})
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(zipData)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	installDir := t.TempDir()
	origFn := binaryInstallDirFn
	binaryInstallDirFn = func(string) string { return installDir }
	defer func() { binaryInstallDirFn = origFn }()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := system.PlatformProfile{OS: "windows", PackageManager: "winget"}

	outPath, err := downloadBinary(nil, h, profile)
	if err != nil {
		t.Fatalf("downloadBinary failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if string(data) != binaryContent {
		t.Errorf("binary content = %q, want %q", data, binaryContent)
	}
}

func TestDownloadBinary_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	_, err := downloadBinary(nil, h, profile)
	if err == nil {
		t.Fatal("expected error for API 404, got nil")
	}
}

// ── downloadBinary prefers External.Repo over External.Pkg ─────────────────

func TestDownloadBinary_UsesRepoOverPkg(t *testing.T) {
	const binaryContent = "fake-engram-binary"
	var gotAPIPath string

	tarGzData := buildTarGz(t, "engram", []byte(binaryContent))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			gotAPIPath = r.URL.Path
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v1.0.0"})
			return
		}
		w.Write(tarGzData)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	installDir := t.TempDir()
	origFn := binaryInstallDirFn
	binaryInstallDirFn = func(string) string { return installDir }
	defer func() { binaryInstallDirFn = origFn }()

	// Pkg is the bare brew formula; Repo is the GitHub owner/repo for download.
	h := harnessWithMethod("homebrew", "engram", "")
	h.External.Repo = "Gentleman-Programming/engram"
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	if _, err := downloadBinary(nil, h, profile); err != nil {
		t.Fatalf("downloadBinary failed: %v", err)
	}
	if !strings.Contains(gotAPIPath, "Gentleman-Programming/engram") {
		t.Errorf("download should use External.Repo owner/repo; API path was %q, want it to contain %q",
			gotAPIPath, "Gentleman-Programming/engram")
	}
}

// TestDownloadBinary_AddsInstallDirToPath verifies that after a successful
// download the install dir is added to the user PATH — otherwise the binary
// (e.g. engram on Windows, landing in %LOCALAPPDATA%\jr-stack\bin) is on disk
// but not invocable as a command.
func TestDownloadBinary_AddsInstallDirToPath(t *testing.T) {
	tarGzData := buildTarGz(t, "engram", []byte("fake-engram-binary"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v1.0.0"})
			return
		}
		w.Write(tarGzData)
	}))
	defer srv.Close()

	origBase := githubBaseURL
	githubBaseURL = srv.URL
	defer func() { githubBaseURL = origBase }()

	origClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = origClient }()

	installDir := t.TempDir()
	origFn := binaryInstallDirFn
	binaryInstallDirFn = func(string) string { return installDir }
	defer func() { binaryInstallDirFn = origFn }()

	// Override the package-level (TestMain no-op) hook to capture the dir.
	var gotDir string
	origAdd := addToUserPath
	addToUserPath = func(dir string) error { gotDir = dir; return nil }
	defer func() { addToUserPath = origAdd }()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := system.PlatformProfile{OS: "linux", PackageManager: "apt"}

	if _, err := downloadBinary(nil, h, profile); err != nil {
		t.Fatalf("downloadBinary failed: %v", err)
	}
	if gotDir != installDir {
		t.Errorf("AddToUserPath called with %q, want installDir %q", gotDir, installDir)
	}
}

// ── installHomebrew with brew available ────────────────────────────────────

func TestInstallHomebrew_UsesBrewWhenAvailable(t *testing.T) {
	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	h := harnessWithMethod("homebrew", "engram", "")
	profile := system.PlatformProfile{OS: "darwin", PackageManager: "brew"}

	result, err := installHomebrew(nil, h, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.capturedName != "brew" {
		t.Errorf("expected brew command, got %q", fr.capturedName)
	}
	if !strings.Contains(result.BinaryPath, "engram") {
		t.Errorf("BinaryPath = %q should contain 'engram'", result.BinaryPath)
	}
}

// ── TestRunBrew_EngamTapFormula: brew install uses full tap path, binary name is "engram" ─
//
// Engram must be installed via the tap formula "gentleman-programming/tap/engram",
// NOT the bare "engram" (which does not exist in homebrew-core).
// runBrew must call `brew install gentleman-programming/tap/engram` and resolve
// the binary name as filepath.Base("gentleman-programming/tap/engram") == "engram".
func TestRunBrew_EngramTapFormula(t *testing.T) {
	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	const tapFormula = "gentleman-programming/tap/engram"
	result, err := runBrew(nil, tapFormula)
	if err != nil {
		t.Fatalf("runBrew failed: %v", err)
	}

	// 1. brew must be called with the full tap formula (not the bare "engram").
	if len(fr.capturedArgs) == 0 || fr.capturedArgs[len(fr.capturedArgs)-1] != tapFormula {
		t.Errorf("brew args = %v, want last arg %q", fr.capturedArgs, tapFormula)
	}

	// 2. BinaryPath must contain "engram" (filepath.Base of the tap path).
	if !strings.Contains(result.BinaryPath, "engram") {
		t.Errorf("BinaryPath = %q should contain 'engram' (Base of tap formula)", result.BinaryPath)
	}
}

// ── archive helpers ───────────────────────────────────────────────────────

func buildTarGz(t *testing.T, filename string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name:     filename,
		Mode:     0o755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write tar content: %v", err)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func buildZip(t *testing.T, filename string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create(filename)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	io.Copy(f, bytes.NewReader(content))
	zw.Close()
	return buf.Bytes()
}

// ── writeExecutable replaces an in-use (locked) binary ─────────────────────

// TestWriteExecutable_ReplacesInUseBinary covers the reinstall case: the target
// already holds a binary that cannot be truncated in place. On Windows a running
// .exe is locked, so os.OpenFile with O_TRUNC fails ("being used by another
// process"). We reproduce that portably with a read-only (0o400) file: opening
// it O_WRONLY|O_TRUNC fails the same way on every OS, yet it can still be
// renamed. writeExecutable must move the stale binary aside to "<path>.old" and
// write the new content fresh — the self-update pattern for in-use executables.
func TestWriteExecutable_ReplacesInUseBinary(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "engram.exe")

	if err := os.WriteFile(outPath, []byte("old-binary"), 0o400); err != nil {
		t.Fatalf("seed read-only binary: %v", err)
	}
	t.Cleanup(func() { os.Chmod(outPath, 0o600); os.Chmod(outPath+".old", 0o600) })

	if err := writeExecutable(strings.NewReader("new-binary"), outPath); err != nil {
		t.Fatalf("writeExecutable failed: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read new binary: %v", err)
	}
	if string(got) != "new-binary" {
		t.Errorf("new binary content = %q, want %q", got, "new-binary")
	}

	aside, err := os.ReadFile(outPath + ".old")
	if err != nil {
		t.Fatalf("stale binary not moved aside: %v", err)
	}
	if string(aside) != "old-binary" {
		t.Errorf(".old content = %q, want %q", aside, "old-binary")
	}
}

// TestWriteExecutable_FreshWrite is the happy path: no pre-existing target, so
// the binary is written in place and no ".old" sidecar is left behind. Guards
// the rename-aside branch from firing (and littering) on a normal first install.
func TestWriteExecutable_FreshWrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "engram.exe")

	if err := writeExecutable(strings.NewReader("fresh"), outPath); err != nil {
		t.Fatalf("writeExecutable failed: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil || string(got) != "fresh" {
		t.Fatalf("content = %q (err %v), want %q", got, err, "fresh")
	}
	if _, err := os.Stat(outPath + ".old"); !os.IsNotExist(err) {
		t.Errorf("unexpected .old sidecar created on a fresh write (stat err: %v)", err)
	}
}

// TestWriteExecutable_ClearsStaleAside verifies a leftover ".old" from a prior
// replace does not block a new replace: it is cleared before the rename, so the
// freshly displaced binary takes its place.
func TestWriteExecutable_ClearsStaleAside(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "engram.exe")
	if err := os.WriteFile(outPath, []byte("v1"), 0o400); err != nil {
		t.Fatalf("seed read-only binary: %v", err)
	}
	if err := os.WriteFile(outPath+".old", []byte("ancient"), 0o600); err != nil {
		t.Fatalf("seed stale aside: %v", err)
	}
	t.Cleanup(func() { os.Chmod(outPath, 0o600); os.Chmod(outPath+".old", 0o600) })

	if err := writeExecutable(strings.NewReader("v2"), outPath); err != nil {
		t.Fatalf("writeExecutable failed: %v", err)
	}

	got, _ := os.ReadFile(outPath)
	if string(got) != "v2" {
		t.Errorf("new binary = %q, want %q", got, "v2")
	}
	aside, _ := os.ReadFile(outPath + ".old")
	if string(aside) != "v1" {
		t.Errorf(".old = %q, want the displaced v1 (ancient leftover should be cleared)", aside)
	}
}
