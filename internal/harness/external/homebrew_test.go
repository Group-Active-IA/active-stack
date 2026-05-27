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
	got := buildAssetURL("https://github.com", "Owner", "repo", "1.0.0", "linux", "amd64")
	want := "https://github.com/Owner/repo/releases/download/v1.0.0/repo_1.0.0_linux_amd64.tar.gz"
	if got != want {
		t.Errorf("buildAssetURL = %q, want %q", got, want)
	}

	gotWin := buildAssetURL("https://github.com", "Owner", "repo", "1.0.0", "windows", "amd64")
	if !strings.HasSuffix(gotWin, ".zip") {
		t.Errorf("windows asset URL should end in .zip, got %q", gotWin)
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
