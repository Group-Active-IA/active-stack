package update_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/update"
)

func TestDownloadReleaseBinaryChecksChecksum(t *testing.T) {
	payload := []byte("new executable")
	hash := sha256.Sum256(payload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/Group-Active-IA/active-stack/releases/latest":
			fmt.Fprint(w, `{"tag_name":"v1.2.3"}`)
		case "/Group-Active-IA/active-stack/releases/download/v1.2.3/SHA256SUMS.txt":
			fmt.Fprintf(w, "%x  active-stack_windows_amd64.exe\n", hash)
		case "/Group-Active-IA/active-stack/releases/download/v1.2.3/active-stack_windows_amd64.exe":
			w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := update.ReleaseClient{HTTP: server.Client(), APIBase: server.URL, DownloadBase: server.URL}
	release, err := client.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if release.Tag != "v1.2.3" {
		t.Fatalf("tag = %q", release.Tag)
	}
	path, err := client.DownloadBinary(context.Background(), release, "windows", "amd64", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	if filepath.Ext(path) != ".exe" {
		t.Fatalf("Windows helper has no executable extension: %s", path)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(payload) {
		t.Fatalf("binary = %q, %v", got, err)
	}
}

func TestDownloadedWindowsBinaryCanStart(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows executable launch regression")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(payload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "SHA256SUMS.txt") {
			fmt.Fprintf(w, "%x  active-stack_windows_amd64.exe\n", hash)
			return
		}
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	client := update.ReleaseClient{HTTP: server.Client(), DownloadBase: server.URL}
	path, err := client.DownloadBinary(context.Background(), update.Release{Tag: "v1.2.3"}, "windows", "amd64", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(path, "-test.run=^TestDownloadedBinaryHelperProcess$")
	cmd.Env = append(os.Environ(), "ACTIVE_STACK_DOWNLOADED_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("downloaded executable did not launch: %v; output=%s", err, output)
	}
	if !strings.Contains(string(output), "downloaded helper started") {
		t.Fatalf("unexpected helper output: %s", output)
	}
}

func TestDownloadedBinaryHelperProcess(t *testing.T) {
	if os.Getenv("ACTIVE_STACK_DOWNLOADED_HELPER") != "1" {
		return
	}
	fmt.Println("downloaded helper started")
	os.Exit(0)
}

// Every platform the release workflow publishes must be updatable.
func TestDownloadReleaseBinaryCoversPublishedPlatforms(t *testing.T) {
	cases := []struct{ goos, goarch, asset string }{
		{"windows", "amd64", "active-stack_windows_amd64.exe"},
		{"linux", "amd64", "active-stack_linux_amd64"},
		{"darwin", "amd64", "active-stack_darwin_amd64"},
		{"darwin", "arm64", "active-stack_darwin_arm64"},
	}
	for _, tc := range cases {
		t.Run(tc.goos+"/"+tc.goarch, func(t *testing.T) {
			payload := []byte("binary for " + tc.asset)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/Group-Active-IA/active-stack/releases/download/v1.2.3/SHA256SUMS.txt":
					fmt.Fprintf(w, "%x  %s\n", sha256.Sum256(payload), tc.asset)
				case "/Group-Active-IA/active-stack/releases/download/v1.2.3/" + tc.asset:
					w.Write(payload)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			if !update.ReleaseBinaryAvailable(tc.goos, tc.goarch) {
				t.Fatalf("no release binary reported for %s/%s", tc.goos, tc.goarch)
			}
			client := update.ReleaseClient{HTTP: server.Client(), DownloadBase: server.URL}
			path, err := client.DownloadBinary(context.Background(), update.Release{Tag: "v1.2.3"}, tc.goos, tc.goarch, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != string(payload) {
				t.Fatalf("binary = %q, %v", got, err)
			}
		})
	}
}

func TestReleaseBinaryUnavailableForUnpublishedPlatforms(t *testing.T) {
	for _, p := range [][2]string{{"windows", "arm64"}, {"linux", "arm64"}, {"darwin", "386"}, {"freebsd", "amd64"}} {
		if update.ReleaseBinaryAvailable(p[0], p[1]) {
			t.Errorf("%s/%s reported as published", p[0], p[1])
		}
	}
}

func TestDownloadReleaseBinaryRejectsBadChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "SHA256SUMS.txt") {
			fmt.Fprint(w, strings.Repeat("0", 64)+"  active-stack_linux_amd64\n")
			return
		}
		fmt.Fprint(w, "tampered")
	}))
	defer server.Close()
	client := update.ReleaseClient{HTTP: server.Client(), APIBase: server.URL, DownloadBase: server.URL}
	dir := t.TempDir()
	_, err := client.DownloadBinary(context.Background(), update.Release{Tag: "v1.2.3"}, "linux", "amd64", dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".new" {
			t.Fatalf("unverified payload left behind: %s", e.Name())
		}
	}
}
