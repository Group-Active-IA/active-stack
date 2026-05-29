package system

import (
	"context"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func harnessByType(t model.HarnessType, method string) model.Harness {
	h := model.Harness{Type: t, InstallModes: []model.InstallMode{model.ModeLite}}
	switch t {
	case model.HarnessSkill:
		h.Source = &model.Source{Method: method}
	case model.HarnessExternal:
		h.External = &model.External{Method: method}
	}
	return h
}

// ─── RequiredDependencies ─────────────────────────────────────────────────────

func TestRequiredDependencies_ConfigOnly_Empty(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessConfig, ""),
	}

	got := RequiredDependencies(harnesses, profile)
	if len(got) != 0 {
		t.Fatalf("config-only harness: want 0 deps, got %v", depNames(got))
	}
}

func TestRequiredDependencies_SkillEmbed_Empty(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessSkill, "embed"),
	}

	got := RequiredDependencies(harnesses, profile)
	if len(got) != 0 {
		t.Fatalf("embed-skill harness: want 0 deps, got %v", depNames(got))
	}
}

func TestRequiredDependencies_ExternalHomebrew_Empty(t *testing.T) {
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessExternal, "homebrew"),
		harnessByType(model.HarnessExternal, "download"),
		harnessByType(model.HarnessExternal, "mcp"),
		harnessByType(model.HarnessExternal, "go-install"),
	}

	got := RequiredDependencies(harnesses, profile)
	if len(got) != 0 {
		t.Fatalf("homebrew/download/mcp/go-install externals: want 0 deps, got %v", depNames(got))
	}
}

func TestRequiredDependencies_ExternalNpm_NodeAndNpm(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessExternal, "npm"),
	}

	got := RequiredDependencies(harnesses, profile)
	names := depNames(got)

	if !containsDep(names, "node") {
		t.Fatalf("npm external: want node in deps, got %v", names)
	}
	if !containsDep(names, "npm") {
		t.Fatalf("npm external: want npm in deps, got %v", names)
	}
	if containsDep(names, "npx") {
		t.Fatalf("npm external: must NOT include npx, got %v", names)
	}
}

func TestRequiredDependencies_SkillNpx_NodeNpmNpx(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessSkill, "npx"),
	}

	got := RequiredDependencies(harnesses, profile)
	names := depNames(got)

	for _, want := range []string{"node", "npm", "npx"} {
		if !containsDep(names, want) {
			t.Fatalf("npx-skill: want %q in deps, got %v", want, names)
		}
	}
}

func TestRequiredDependencies_SkillClone_Git(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessSkill, "clone"),
	}

	got := RequiredDependencies(harnesses, profile)
	names := depNames(got)

	if !containsDep(names, "git") {
		t.Fatalf("clone-skill: want git in deps, got %v", names)
	}
	if containsDep(names, "node") || containsDep(names, "npm") {
		t.Fatalf("clone-skill: must NOT include node/npm, got %v", names)
	}
}

func TestRequiredDependencies_MixedNpmAndClone_Deduped(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessExternal, "npm"),
		harnessByType(model.HarnessSkill, "clone"),
		// second npm harness — should not duplicate node/npm
		harnessByType(model.HarnessExternal, "npm"),
	}

	got := RequiredDependencies(harnesses, profile)
	names := depNames(got)

	// Expected: node, npm, git — no duplicates.
	for _, want := range []string{"node", "npm", "git"} {
		if !containsDep(names, want) {
			t.Fatalf("mixed npm+clone: want %q in deps, got %v", want, names)
		}
	}

	// Check no duplicates.
	seen := map[string]int{}
	for _, n := range names {
		seen[n]++
	}
	for name, count := range seen {
		if count > 1 {
			t.Fatalf("dependency %q appears %d times (want 1) — dedup failed", name, count)
		}
	}
}

func TestRequiredDependencies_MixedNpxAndConfig_NpxSet(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessSkill, "npx"),
		harnessByType(model.HarnessConfig, ""),
	}

	got := RequiredDependencies(harnesses, profile)
	names := depNames(got)

	for _, want := range []string{"node", "npm", "npx"} {
		if !containsDep(names, want) {
			t.Fatalf("npx+config: want %q in deps, got %v", want, names)
		}
	}
}

func TestRequiredDependencies_MetadataFromDefineDeps(t *testing.T) {
	// Verify MinVersion and InstallHint come from defineDependencies, not hard-coded.
	profile := PlatformProfile{OS: "darwin", PackageManager: "brew"}
	harnesses := []model.Harness{
		harnessByType(model.HarnessExternal, "npm"),
	}

	got := RequiredDependencies(harnesses, profile)

	for _, dep := range got {
		if dep.Name == "node" {
			if dep.MinVersion != "18.0.0" {
				t.Fatalf("node MinVersion = %q, want 18.0.0", dep.MinVersion)
			}
			if dep.InstallHint == "" {
				t.Fatalf("node InstallHint is empty — must come from defineDependencies")
			}
		}
	}
}

func TestRequiredDependencies_EmptyHarnesses_Empty(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	got := RequiredDependencies(nil, profile)
	if len(got) != 0 {
		t.Fatalf("nil harnesses: want 0 deps, got %v", depNames(got))
	}
}

// ─── npx added to defineDependencies ─────────────────────────────────────────

func TestDefineDependencies_IncludesNpx(t *testing.T) {
	profile := PlatformProfile{OS: "linux", PackageManager: "apt"}
	deps := defineDependencies(profile)

	for _, dep := range deps {
		if dep.Name == "npx" {
			if len(dep.DetectCmd) == 0 || dep.DetectCmd[0] != "npx" {
				t.Fatalf("npx DetectCmd = %v, want [npx ...]", dep.DetectCmd)
			}
			// npx should be required (since failing to detect npx halts the install)
			if !dep.Required {
				t.Fatalf("npx should be Required=true")
			}
			return
		}
	}
	t.Fatalf("npx not found in defineDependencies")
}

// ─── DetectDepsFor ────────────────────────────────────────────────────────────

// TestDetectDepsFor_AllPresent verifies that DetectDepsFor marks AllPresent when
// all deps are already installed (use pre-set Installed=true to avoid exec).
func TestDetectDepsFor_AllPresent(t *testing.T) {
	// We pass deps already known to be installed so the function just re-checks them.
	// Use "echo" as a real binary that is guaranteed present.
	deps := []Dependency{
		{Name: "echo", Required: true, Installed: false, DetectCmd: []string{"echo", "1.0.0"}},
	}

	report := DetectDepsFor(context.Background(), deps)

	if !report.AllPresent {
		t.Fatalf("DetectDepsFor: all deps present, want AllPresent=true, got MissingRequired=%v", report.MissingRequired)
	}
	if len(report.MissingRequired) != 0 {
		t.Fatalf("DetectDepsFor: want no MissingRequired, got %v", report.MissingRequired)
	}
}

// TestDetectDepsFor_MissingRequired verifies that a non-existent binary populates MissingRequired.
func TestDetectDepsFor_MissingRequired(t *testing.T) {
	deps := []Dependency{
		{Name: "nonexistent_binary_abc123", Required: true, DetectCmd: []string{"nonexistent_binary_abc123", "--version"}},
	}

	report := DetectDepsFor(context.Background(), deps)

	if report.AllPresent {
		t.Fatalf("DetectDepsFor: missing dep, want AllPresent=false")
	}
	if len(report.MissingRequired) == 0 {
		t.Fatalf("DetectDepsFor: want MissingRequired to contain the missing dep")
	}
	if report.MissingRequired[0] != "nonexistent_binary_abc123" {
		t.Fatalf("DetectDepsFor: MissingRequired[0] = %q, want nonexistent_binary_abc123", report.MissingRequired[0])
	}
}

// TestDetectDepsFor_EmptyDeps verifies that an empty slice gives AllPresent=true.
func TestDetectDepsFor_EmptyDeps(t *testing.T) {
	report := DetectDepsFor(context.Background(), nil)

	if !report.AllPresent {
		t.Fatalf("DetectDepsFor(nil): want AllPresent=true (nothing required)")
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func depNames(deps []Dependency) []string {
	names := make([]string, len(deps))
	for i, d := range deps {
		names[i] = d.Name
	}
	return names
}

func containsDep(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}
