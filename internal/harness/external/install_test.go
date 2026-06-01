package external

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// TestMain neutralizes addToUserPath for the whole package so any test that
// reaches downloadBinary never mutates the real user PATH. On Windows the real
// AddToUserPath runs PowerShell and writes the user-scoped PATH in the registry;
// letting that fire during `go test` would pollute the developer's PATH with
// throwaway TempDirs. Tests that need to assert the call override addToUserPath
// locally with a defer-restore.
func TestMain(m *testing.M) {
	addToUserPath = func(string) error { return nil }
	os.Exit(m.Run())
}

// ── helpers ────────────────────────────────────────────────────────────────

func harnessWithMethod(method, pkg, url string) model.Harness {
	return model.Harness{
		ID:   "test-harness",
		Name: "Test Harness",
		Type: model.HarnessExternal,
		External: &model.External{
			Method: method,
			Pkg:    pkg,
			URL:    url,
		},
	}
}

func linuxProfile(npmWritable bool) system.PlatformProfile {
	return system.PlatformProfile{OS: "linux", NpmWritable: npmWritable}
}

// fakeRunner captures the command name and args for assertion.
type fakeRunner struct {
	capturedName string
	capturedArgs []string
	output       []byte
	err          error
}

func (f *fakeRunner) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	f.capturedName = name
	f.capturedArgs = args
	return f.output, f.err
}

// withFakeRunner replaces the package-level runner and returns a cleanup func.
func withFakeRunner(f *fakeRunner) func() {
	orig := runner
	runner = f.run
	return func() { runner = orig }
}

// withFakeLookPath replaces lookPath and returns a cleanup func.
func withFakeLookPath(fn func(string) (string, error)) func() {
	orig := lookPath
	lookPath = fn
	return func() { lookPath = orig }
}

// ── dispatch ───────────────────────────────────────────────────────────────

func TestInstall_UnknownMethod(t *testing.T) {
	h := harnessWithMethod("ftp", "something", "")
	_, err := Install(context.Background(), h, linuxProfile(true), nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for unknown method, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported install method") {
		t.Errorf("error message should mention unsupported method, got: %v", err)
	}
}

func TestInstall_NilExternal(t *testing.T) {
	h := model.Harness{ID: "bad", Type: model.HarnessExternal}
	_, err := Install(context.Background(), h, linuxProfile(true), nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil External, got nil")
	}
}

// ── npm sudo logic ─────────────────────────────────────────────────────────

func TestUseSudo(t *testing.T) {
	tests := []struct {
		name     string
		profile  system.PlatformProfile
		wantSudo bool
	}{
		{
			name:     "linux writable no sudo",
			profile:  system.PlatformProfile{OS: "linux", NpmWritable: true},
			wantSudo: false,
		},
		{
			name:     "linux not writable uses sudo",
			profile:  system.PlatformProfile{OS: "linux", NpmWritable: false},
			wantSudo: true,
		},
		{
			name:     "windows never sudo",
			profile:  system.PlatformProfile{OS: "windows", NpmWritable: false},
			wantSudo: false,
		},
		{
			name:     "darwin writable no sudo",
			profile:  system.PlatformProfile{OS: "darwin", NpmWritable: true},
			wantSudo: false,
		},
		{
			name:     "darwin not writable uses sudo",
			profile:  system.PlatformProfile{OS: "darwin", NpmWritable: false},
			wantSudo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useSudo(tt.profile)
			if got != tt.wantSudo {
				t.Errorf("useSudo(%+v) = %v, want %v", tt.profile, got, tt.wantSudo)
			}
		})
	}
}

func TestNPM_CommandBuilt_NoSudo(t *testing.T) {
	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	h := harnessWithMethod("npm", "@fission-ai/openspec", "")
	profile := system.PlatformProfile{OS: "windows", NpmWritable: false}

	_, err := installNPM(context.Background(), h, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fr.capturedName != "npm" {
		t.Errorf("command = %q, want npm", fr.capturedName)
	}
}

func TestNPM_CommandBuilt_WithSudo(t *testing.T) {
	fr := &fakeRunner{output: []byte("ok")}
	defer withFakeRunner(fr)()
	defer withFakeLookPath(func(name string) (string, error) {
		return "/usr/local/bin/" + name, nil
	})()

	h := harnessWithMethod("npm", "@fission-ai/openspec", "")
	profile := system.PlatformProfile{OS: "linux", NpmWritable: false}

	_, err := installNPM(context.Background(), h, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fr.capturedName != "sudo" {
		t.Errorf("command = %q, want sudo", fr.capturedName)
	}

	found := false
	for _, a := range fr.capturedArgs {
		if a == "@fission-ai/openspec" {
			found = true
		}
	}
	if !found {
		t.Errorf("args %v do not contain package name", fr.capturedArgs)
	}
}

func TestNPM_Install_Error(t *testing.T) {
	fr := &fakeRunner{err: errors.New("npm not found")}
	defer withFakeRunner(fr)()

	h := harnessWithMethod("npm", "somepackage", "")
	_, err := installNPM(context.Background(), h, linuxProfile(true))
	if err == nil {
		t.Fatal("expected error when npm fails")
	}
}

func TestNPM_MissingPkg(t *testing.T) {
	h := harnessWithMethod("npm", "", "")
	_, err := installNPM(context.Background(), h, linuxProfile(true))
	if err == nil {
		t.Fatal("expected error for empty pkg")
	}
}

// ── pkgBinaryName ──────────────────────────────────────────────────────────

func TestPkgBinaryName(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{"@fission-ai/openspec", "openspec"},
		{"openspec", "openspec"},
		{"@scope/some-tool", "some-tool"},
		{"plain", "plain"},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := pkgBinaryName(tt.pkg)
			if got != tt.want {
				t.Errorf("pkgBinaryName(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}
