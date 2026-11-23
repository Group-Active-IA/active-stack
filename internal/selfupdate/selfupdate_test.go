package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fakeServer(t *testing.T, tag string, bin []byte, sumHex string, hits *int) *httptest.Server {
	t.Helper()
	asset := AssetName("windows", "amd64")
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			if hits != nil {
				*hits++
			}
			fmt.Fprintf(w, `{"tag_name":%q,"assets":[{"name":%q,"browser_download_url":%q},{"name":"SHA256SUMS.txt","browser_download_url":%q}]}`,
				tag, asset, srv.URL+"/bin", srv.URL+"/sums")
		case "/bin":
			w.Write(bin)
		case "/sums":
			fmt.Fprintf(w, "%s  %s\n", sumHex, asset)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func sumOf(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func useServer(t *testing.T, s *httptest.Server) {
	old := APIBase
	APIBase = s.URL
	t.Cleanup(func() { APIBase = old })
}

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		cur, lat string
		want     bool
	}{{"v0.8.1", "v0.9.0", true}, {"v0.8.1", "v0.8.1", false}, {"v0.9.0", "v0.8.1", false}, {"dev", "v9.9.9", false}, {"v0.8.1", "junk", false}, {"v0.8.9", "v0.8.10", true}} {
		if got := Newer(c.cur, c.lat); got != c.want {
			t.Errorf("Newer(%q,%q)=%v want %v", c.cur, c.lat, got, c.want)
		}
	}
}

func TestCheckOfflineIsSilent(t *testing.T) {
	old := APIBase
	APIBase = "http://127.0.0.1:1"
	defer func() { APIBase = old }()
	if got := Check(context.Background(), t.TempDir(), "v0.8.1"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestCheckWarnsOnceADay(t *testing.T) {
	hits := 0
	useServer(t, fakeServer(t, "v0.9.0", nil, "", &hits))
	home := t.TempDir()
	if got := Check(context.Background(), home, "v0.8.1"); got != "v0.9.0" {
		t.Fatalf("got %q", got)
	}
	if got := Check(context.Background(), home, "v0.8.1"); got != "v0.9.0" || hits != 1 {
		t.Fatalf("second check: got %q, hits %d (want the cache, 1 hit)", got, hits)
	}
	old := Now
	Now = func() time.Time { return time.Now().Add(25 * time.Hour) }
	defer func() { Now = old }()
	Check(context.Background(), home, "v0.8.1")
	if hits != 2 {
		t.Fatalf("after 24h the network must be consulted again, hits=%d", hits)
	}
}

func TestCheckEqualOrDevDoesNotWarn(t *testing.T) {
	useServer(t, fakeServer(t, "v0.8.1", nil, "", nil))
	if got := Check(context.Background(), t.TempDir(), "v0.8.1"); got != "" {
		t.Fatal(got)
	}
	if got := Check(context.Background(), t.TempDir(), "dev"); got != "" {
		t.Fatal(got)
	}
}

func setup(t *testing.T) (home, exe string) {
	t.Helper()
	home = t.TempDir()
	exe = filepath.Join(t.TempDir(), "active-stack.exe")
	if err := os.WriteFile(exe, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	return home, exe
}

func TestApplyReplacesWhenChecksumMatches(t *testing.T) {
	home, exe := setup(t)
	defer SetSnapshotCreate(func(string, []string) error { return nil })()
	bin := []byte("NEW-BINARY")
	useServer(t, fakeServer(t, "v0.9.0", bin, sumOf(bin), nil))
	tag, err := Apply(context.Background(), home, exe, "windows", "amd64")
	if err != nil || tag != "v0.9.0" {
		t.Fatalf("tag=%q err=%v", tag, err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "NEW-BINARY" {
		t.Fatalf("exe = %q", b)
	}
}

func TestApplyKeepsBinaryOnBadChecksum(t *testing.T) {
	home, exe := setup(t)
	defer SetSnapshotCreate(func(string, []string) error { return nil })()
	useServer(t, fakeServer(t, "v0.9.0", []byte("EVIL"), sumOf([]byte("other")), nil))
	if _, err := Apply(context.Background(), home, exe, "windows", "amd64"); err != ErrChecksum {
		t.Fatalf("err=%v want ErrChecksum", err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "OLD" {
		t.Fatalf("exe changed: %q", b)
	}
}

func TestApplyKeepsBinaryWhenBackupFails(t *testing.T) {
	home, exe := setup(t)
	defer SetSnapshotCreate(func(string, []string) error { return fmt.Errorf("disk full") })()
	bin := []byte("NEW")
	useServer(t, fakeServer(t, "v0.9.0", bin, sumOf(bin), nil))
	if _, err := Apply(context.Background(), home, exe, "windows", "amd64"); err == nil {
		t.Fatal("want an error")
	}
	if b, _ := os.ReadFile(exe); string(b) != "OLD" {
		t.Fatalf("exe changed: %q", b)
	}
}

func TestApplyUnsupportedPlatform(t *testing.T) {
	if _, err := Apply(context.Background(), t.TempDir(), "x", "darwin", "arm64"); err == nil {
		t.Fatal("want an error")
	}
}
