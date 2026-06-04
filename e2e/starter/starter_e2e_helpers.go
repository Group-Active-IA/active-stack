// Package starter provides a hybrid E2E test suite for `jr-stack starter add`.
//
// # Architecture
//
// The suite has two arms:
//
//   - Hermetic (starter_hermetic_test.go): no build tag, always runs in `go test ./...`.
//     Uses local mock git repos served via file:// URL — no network, fully deterministic.
//
//   - Network (starter_network_test.go): `//go:build e2e_network` tag, opt-in only.
//     Clones real upstream repos (e.g. JuanCruzRobledo/jr-skills). Never runs in default CI.
//
// This file holds shared assertion helpers and fixture builders used by both arms.
// It has NO build tag so it is always compiled.
package starter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/catalog"
)

// ─────────────────────────────────────────────────────────────────
// Git availability guard
// ─────────────────────────────────────────────────────────────────

// requireGit skips the test if git is not on PATH.
// The hermetic and network arms both shell out to git (via cloneInstaller);
// if git is absent, t.Skip is the correct response (not failure).
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH — skipping E2E test (git is the installer's hard dependency)")
	}
}

// ─────────────────────────────────────────────────────────────────
// Fixture git-repo builder
// ─────────────────────────────────────────────────────────────────

// FixtureRepoResult describes a created fixture git repository.
type FixtureRepoResult struct {
	// Dir is the absolute path to the repository root.
	Dir string
	// FileURL is a file:// URL pointing to the repository. Pass this as
	// source.Repo in a test harness so cloneInstaller uses the local repo
	// without making any network request.
	FileURL string
}

// buildFixtureRepo creates a minimal git repository in t.TempDir() containing
// a SKILL.md at the root (and optionally inside skills/<name>/ if pathInRepo is
// set), commits everything, and returns the repo path + file:// URL.
//
// Parameters:
//   - skillContent: content written to SKILL.md.
//   - pathInRepo: when non-empty, also writes SKILL.md at that subdirectory
//     (e.g. "skills/my-skill") to simulate a monorepo layout.
//
// The caller should use FixtureRepoResult.FileURL as source.Repo in a Harness.
func buildFixtureRepo(t *testing.T, skillContent, pathInRepo string) FixtureRepoResult {
	t.Helper()
	requireGit(t)

	dir := t.TempDir()
	ctx := context.Background()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fixture repo: %v\n%s", err, string(out))
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	// Write SKILL.md at the root (D2: root layout — our convention).
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("fixture repo: write root SKILL.md: %v", err)
	}
	run("git", "add", "SKILL.md")

	// Also write SKILL.md inside pathInRepo/ if requested (source.path cases).
	if pathInRepo != "" {
		subDir := filepath.Join(dir, filepath.FromSlash(pathInRepo))
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("fixture repo: mkdir %s: %v", pathInRepo, err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
			t.Fatalf("fixture repo: write %s/SKILL.md: %v", pathInRepo, err)
		}
		run("git", "add", ".")
	}

	run("git", "commit", "-m", "initial")

	// Produce a file:// URL. On Windows the path uses backslashes; git requires
	// forward slashes in the URL even on Windows.
	urlPath := filepath.ToSlash(dir)
	if !strings.HasPrefix(urlPath, "/") {
		// Windows absolute path: "C:/..." → "/C:/..." (three-slash form file:///C:/...)
		urlPath = "/" + urlPath
	}
	fileURL := "file://" + urlPath

	return FixtureRepoResult{Dir: dir, FileURL: fileURL}
}

// buildBrokenFixtureRepo creates a fixture git repo that has a SKILL.md at the
// root but is MISSING the expected source.path subdir. This simulates the
// upstream-path-moved failure that the real 6 stale third-party skills exhibit.
func buildBrokenFixtureRepo(t *testing.T) FixtureRepoResult {
	t.Helper()
	// Write something at root but NOT at the expected subdir path.
	return buildFixtureRepo(t, "# root only — no subdir", "" /* no pathInRepo */)
}

// ─────────────────────────────────────────────────────────────────
// Test-only catalog builder
// ─────────────────────────────────────────────────────────────────

// FixtureHarness describes a test harness for the fixture catalog.
type FixtureHarness struct {
	// ID is the harness identifier.
	ID string
	// FileURL is the file:// URL pointing at the fixture git repo.
	FileURL string
	// PathInRepo is the optional source.path (mirrors source.path in harnesses.yaml).
	// When non-empty the installer looks for SKILL.md at this subdir.
	PathInRepo string
	// BestEffort when true marks the harness as best-effort (soft failure on install error).
	BestEffort bool
}

// buildFixtureCatalog constructs a *catalog.Catalog in-memory from a YAML
// description derived from the provided FixtureHarness entries plus a starter
// whose id is starterID and whose harnesses are the IDs in harnesses.
// The catalog is built via catalog.ParseForTest (no harnesses.yaml edits).
func buildFixtureCatalog(t *testing.T, starterID string, harnesses []FixtureHarness) *catalog.Catalog {
	t.Helper()

	// Build YAML for harnesses.
	var sb strings.Builder
	sb.WriteString("harnesses:\n")
	for _, h := range harnesses {
		sb.WriteString(fmt.Sprintf("  - id: %q\n", h.ID))
		sb.WriteString("    name: " + fmt.Sprintf("%q\n", h.ID))
		sb.WriteString("    type: skill\n")
		sb.WriteString("    install_modes: [custom]\n")
		sb.WriteString("    source:\n")
		sb.WriteString("      method: clone\n")
		sb.WriteString(fmt.Sprintf("      repo: %q\n", h.FileURL))
		if h.PathInRepo != "" {
			sb.WriteString(fmt.Sprintf("      path: %q\n", h.PathInRepo))
		}
		if h.BestEffort {
			sb.WriteString("    best_effort: true\n")
		}
		sb.WriteString("    third_party: false\n")
	}

	// Build starter.
	sb.WriteString("starters:\n")
	sb.WriteString(fmt.Sprintf("  - id: %q\n", starterID))
	sb.WriteString(fmt.Sprintf("    name: %q\n", "Fixture: "+starterID))
	sb.WriteString("    description: \"E2E fixture starter\"\n")
	sb.WriteString("    harnesses:\n")
	for _, h := range harnesses {
		sb.WriteString(fmt.Sprintf("      - %q\n", h.ID))
	}

	cat, err := catalog.ParseForTest([]byte(sb.String()))
	if err != nil {
		t.Fatalf("buildFixtureCatalog: parse YAML: %v\nYAML:\n%s", err, sb.String())
	}
	return cat
}

// ─────────────────────────────────────────────────────────────────
// FS post-condition helpers (D4)
// ─────────────────────────────────────────────────────────────────

// tbHelper is the minimal subset of *testing.T needed by the assertion helpers.
// Using this interface makes the helpers usable from self-tests (e.g. a mockTB).
type tbHelper interface {
	Helper()
	Errorf(format string, args ...any)
	Logf(format string, args ...any)
}

// assertSKILLmdExists asserts that SKILL.md is present and non-empty at the
// expected path under skillsDir/<harnessID>/SKILL.md.
// The skillsDir is resolved via the real adapter — never hardcoded (CLAUDE.md §3).
func assertSKILLmdExists(t tbHelper, skillsDir, harnessID string) {
	t.Helper()
	skillPath := filepath.Join(skillsDir, harnessID, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Errorf("SKILL.md not found at %s: %v", skillPath, err)
		return
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		t.Errorf("SKILL.md at %s is empty", skillPath)
	}
}

// assertSKILLmdAbsent asserts that SKILL.md does NOT exist at
// skillsDir/<harnessID>/SKILL.md. Used for degraded harnesses: the installer
// must not leave a half-written file on failure.
func assertSKILLmdAbsent(t tbHelper, skillsDir, harnessID string) {
	t.Helper()
	skillPath := filepath.Join(skillsDir, harnessID, "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		t.Errorf("SKILL.md should NOT exist for degraded harness %s, but found at %s", harnessID, skillPath)
	}
}

// watchForMissingCompanion logs a watch finding (t.Log) when an installed
// skill is missing a referenced companion file. This is detect-only; the test
// does NOT fail and does NOT modify the installed content (C-32 spec).
func watchForMissingCompanion(t tbHelper, skillsDir, harnessID, companionFile string) {
	t.Helper()
	companionPath := filepath.Join(skillsDir, harnessID, companionFile)
	if _, err := os.Stat(companionPath); os.IsNotExist(err) {
		t.Logf("WATCH: harness %q is missing companion file %q at path %s — content incomplete (detect-only, not repaired)",
			harnessID, companionFile, companionPath)
	}
}

// ─────────────────────────────────────────────────────────────────
// Agent adapter helper (for path resolution — never hardcode paths)
// ─────────────────────────────────────────────────────────────────

// claudeSkillsDir returns the Claude agent's skills directory for the given
// homeDir, resolved via the real claude adapter. Never hardcoded (CLAUDE.md §3).
func claudeSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
}
