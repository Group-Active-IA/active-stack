package skill_test

// ──────────────────────────────────────────────────────────────────────────────
// Tests for internal/harness/skill
//
// TDD order: all tests written RED first, then implementation makes them GREEN.
// No real git commands are executed — all exec is mocked via the Runner
// interface.  File system operations use t.TempDir().
// ──────────────────────────────────────────────────────────────────────────────

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeHarness(id, method, repo string, thirdParty bool) model.Harness {
	return model.Harness{
		ID:           id,
		Name:         id,
		Type:         model.HarnessSkill,
		ThirdParty:   thirdParty,
		Source:       &model.Source{Repo: repo, Method: method},
		InstallModes: []model.InstallMode{model.ModeFull},
	}
}

type fakeAdapter struct {
	agent     model.Agent
	skillsDir string
}

func (f fakeAdapter) Agent() model.Agent              { return f.agent }
func (f fakeAdapter) SkillsDir(homeDir string) string { return f.skillsDir }

// stubRunner records the command that was run and optionally creates a file
// to simulate a successful installation side-effect.
type stubRunner struct {
	called [][]string
	err    error
	// sideEffect, if non-nil, is called before returning so the test can
	// create the expected SKILL.md to simulate a real install.
	sideEffect func(args []string)
}

func (r *stubRunner) Run(ctx context.Context, args []string) error {
	r.called = append(r.called, args)
	if r.sideEffect != nil {
		r.sideEffect(args)
	}
	return r.err
}

// ── Task 7.1: Installer dispatch ─────────────────────────────────────────────

func TestInstaller_Clone_CallsRunner(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			// Simulate git clone: create <tempDir>/my-skill/SKILL.md
			// The clone.go impl uses args[len-1] as destDir.
			destDir := args[len(args)-1]
			skillDir := filepath.Join(destDir, "my-skill")
			_ = os.MkdirAll(skillDir, 0o755)
			_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# my-skill"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("my-skill", "clone", "JuanCruz/my-skill", false)

	ins := skill.NewInstaller(runner, nil)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AlreadyInstalled {
		t.Error("expected not AlreadyInstalled on fresh install")
	}
	if len(runner.called) == 0 {
		t.Error("expected runner to be called for clone")
	}
	// First arg must be "git"
	if runner.called[0][0] != "git" {
		t.Errorf("expected first arg %q, got %q", "git", runner.called[0][0])
	}
}

func TestInstaller_Embed_WritesFile(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	// Build a test FS with an embedded SKILL.md.
	testFS := fstest.MapFS{
		"skills/embed-skill/SKILL.md": &fstest.MapFile{Data: []byte("# embed skill")},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("embed-skill", "embed", "", false)

	ins := skill.NewInstaller(nil, testFS)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	dest := filepath.Join(skillsDir, "embed-skill", "SKILL.md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("SKILL.md not written: %v", err)
	}
	if string(data) != "# embed skill" {
		t.Errorf("content mismatch: got %q", string(data))
	}
}

func TestInstaller_EmptyRepo_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	runner := &stubRunner{}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("bad-skill", "clone", "", false) // empty repo

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error for empty Source.Repo, got nil")
	}
	if !strings.Contains(err.Error(), "repo") {
		t.Errorf("expected error to mention 'repo', got %q", err.Error())
	}
}

func TestInstaller_EmptySkillsDir_Skips(t *testing.T) {
	home := t.TempDir()
	runner := &stubRunner{}
	// adapter with empty skillsDir → should be skipped
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: ""}
	h := makeHarness("my-skill", "clone", "some/repo", false)

	ins := skill.NewInstaller(runner, nil)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when adapter has empty skillsDir, got %d", len(results))
	}
}

func TestInstaller_UnknownMethod_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	runner := &stubRunner{}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("x", "ftp", "some/repo", false)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error for unknown method, got nil")
	}
}

// ── Task 7.2: clone.go ────────────────────────────────────────────────────────

func TestClone_UsesDepth1AndHTTPS(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			destDir := args[len(args)-1]
			skillDir := filepath.Join(destDir, "s")
			_ = os.MkdirAll(skillDir, 0o755)
			_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("ok"), 0o644)
		},
	}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("s", "clone", "owner/s", false)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(runner.called) == 0 {
		t.Fatal("runner not called")
	}
	args := runner.called[0]
	hasDepth1 := false
	hasHTTPS := false
	for _, a := range args {
		if a == "--depth" {
			hasDepth1 = true
		}
		if strings.HasPrefix(a, "https://github.com/") {
			hasHTTPS = true
		}
	}
	if !hasDepth1 {
		t.Errorf("expected --depth flag; args: %v", args)
	}
	if !hasHTTPS {
		t.Errorf("expected https://github.com/ URL; args: %v", args)
	}
}

func TestClone_WithRef_UsesBranchFlag(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			destDir := args[len(args)-1]
			skillDir := filepath.Join(destDir, "s")
			_ = os.MkdirAll(skillDir, 0o755)
			_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("ok"), 0o644)
		},
	}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := model.Harness{
		ID: "s", Name: "s", Type: model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/s", Ref: "v1.2", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	args := runner.called[0]
	hasBranch := false
	for i, a := range args {
		if a == "--branch" && i+1 < len(args) && args[i+1] == "v1.2" {
			hasBranch = true
		}
	}
	if !hasBranch {
		t.Errorf("expected --branch v1.2; args: %v", args)
	}
}

func TestClone_RunnerError_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	runner := &stubRunner{err: errors.New("git not found")}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("s", "clone", "owner/s", false)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error on runner failure, got nil")
	}
}

// ── Task 7.4: embed.go ────────────────────────────────────────────────────────

func TestEmbed_AssetPresent_WritesFile(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	testFS := fstest.MapFS{
		"skills/e-skill/SKILL.md": &fstest.MapFile{Data: []byte("# embedded")},
	}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("e-skill", "embed", "", false)

	ins := skill.NewInstaller(nil, testFS)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 || results[0].AlreadyInstalled {
		t.Fatalf("unexpected results: %+v", results)
	}
	data, readErr := os.ReadFile(filepath.Join(skillsDir, "e-skill", "SKILL.md"))
	if readErr != nil {
		t.Fatalf("SKILL.md not found: %v", readErr)
	}
	if string(data) != "# embedded" {
		t.Errorf("content: want %q, got %q", "# embedded", string(data))
	}
}

func TestEmbed_AssetAbsent_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	// FS has no entry for "missing-skill"
	testFS := fstest.MapFS{}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("missing-skill", "embed", "", false)

	ins := skill.NewInstaller(nil, testFS)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing embed asset, got nil")
	}
}

// ── Task 7.5: idempotence ─────────────────────────────────────────────────────

func TestIdempotent_IdenticalContent_NoOp(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	// Pre-populate the destination with identical content.
	existingDir := filepath.Join(skillsDir, "e-skill")
	_ = os.MkdirAll(existingDir, 0o755)
	_ = os.WriteFile(filepath.Join(existingDir, "SKILL.md"), []byte("# embedded"), 0o644)

	testFS := fstest.MapFS{
		"skills/e-skill/SKILL.md": &fstest.MapFile{Data: []byte("# embedded")},
	}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("e-skill", "embed", "", false)

	ins := skill.NewInstaller(nil, testFS)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].AlreadyInstalled {
		t.Error("expected AlreadyInstalled=true for identical content")
	}
}

func TestIdempotent_DifferentContent_CallsBackup(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	backupDir := t.TempDir()

	// Pre-populate with different content.
	existingDir := filepath.Join(skillsDir, "e-skill")
	_ = os.MkdirAll(existingDir, 0o755)
	_ = os.WriteFile(filepath.Join(existingDir, "SKILL.md"), []byte("# old version"), 0o644)

	testFS := fstest.MapFS{
		"skills/e-skill/SKILL.md": &fstest.MapFile{Data: []byte("# new version")},
	}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("e-skill", "embed", "", false)

	ins := skill.NewInstaller(nil, testFS)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, backupDir)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AlreadyInstalled {
		t.Error("expected AlreadyInstalled=false when content differs")
	}
	// Verify backup was created (backup dir must be non-empty).
	entries, _ := os.ReadDir(backupDir)
	if len(entries) == 0 {
		t.Error("expected backup to be created for changed content")
	}
	// Verify new content was written.
	data, _ := os.ReadFile(filepath.Join(skillsDir, "e-skill", "SKILL.md"))
	if string(data) != "# new version" {
		t.Errorf("expected new content, got %q", string(data))
	}
}

// ── Task 7.6: verify.go ───────────────────────────────────────────────────────

func TestVerify_SkillMDPresent_NoError(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "s")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# ok"), 0o644)

	if err := skill.Verify(dir, "s"); err != nil {
		t.Errorf("Verify() error on present SKILL.md: %v", err)
	}
}

func TestVerify_SkillMDAbsent_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	// No SKILL.md written.
	if err := skill.Verify(dir, "s"); err == nil {
		t.Error("Verify() expected error for absent SKILL.md, got nil")
	}
}

func TestVerify_SkillMDEmpty_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "s")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(""), 0o644)

	if err := skill.Verify(dir, "s"); err == nil {
		t.Error("Verify() expected error for empty SKILL.md, got nil")
	}
}

// ── C-16: root-layout + .git exclusion ───────────────────────────────────────

// TestClone_RootLayout_InstallsFromRoot verifies that a cloned repo whose
// SKILL.md lives at the clone root (no <skillID>/ subdir) is installed
// correctly. This is the layout used by our own skill repos.
func TestClone_RootLayout_InstallsFromRoot(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			// Simulate git clone: SKILL.md at clone root (root layout).
			destDir := args[len(args)-1]
			_ = os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte("# root-skill"), 0o644)
			assetsDir := filepath.Join(destDir, "assets")
			_ = os.MkdirAll(assetsDir, 0o755)
			_ = os.WriteFile(filepath.Join(assetsDir, "config.yaml"), []byte("key: val"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("root-skill", "clone", "JuanCruz/root-skill", false)

	ins := skill.NewInstaller(runner, nil)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].AlreadyInstalled {
		t.Error("expected not AlreadyInstalled on fresh install")
	}

	// SKILL.md must be installed at <skillsDir>/<id>/SKILL.md.
	dest := filepath.Join(skillsDir, "root-skill", "SKILL.md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("SKILL.md not found in skills dir: %v", err)
	}
	if string(data) != "# root-skill" {
		t.Errorf("SKILL.md content: want %q, got %q", "# root-skill", string(data))
	}

	// assets/config.yaml must also be installed.
	assetDest := filepath.Join(skillsDir, "root-skill", "assets", "config.yaml")
	if _, err := os.Stat(assetDest); err != nil {
		t.Errorf("assets/config.yaml not found in installed skill: %v", err)
	}
}

// TestClone_RootLayout_ExcludesGitDir verifies that when the clone root is the
// skill source, the .git/ directory is excluded from the installed skill but
// non-.git dotfiles like .gitignore are preserved.
func TestClone_RootLayout_ExcludesGitDir(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			destDir := args[len(args)-1]
			// Root SKILL.md.
			_ = os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte("# dotfile-skill"), 0o644)
			// .gitignore at root (must be preserved).
			_ = os.WriteFile(filepath.Join(destDir, ".gitignore"), []byte("*.tmp\n"), 0o644)
			// .git/ directory (must be excluded).
			gitDir := filepath.Join(destDir, ".git")
			_ = os.MkdirAll(gitDir, 0o755)
			_ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("dotfile-skill", "clone", "JuanCruz/dotfile-skill", false)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// .git/ must NOT be in the installed skill.
	gitInDest := filepath.Join(skillsDir, "dotfile-skill", ".git")
	if _, err := os.Stat(gitInDest); err == nil {
		t.Error(".git/ directory was copied into the installed skill — it must be excluded")
	}

	// .gitignore must be preserved.
	gitignoreDest := filepath.Join(skillsDir, "dotfile-skill", ".gitignore")
	data, err := os.ReadFile(gitignoreDest)
	if err != nil {
		t.Fatalf(".gitignore not found in installed skill — it must be preserved: %v", err)
	}
	if string(data) != "*.tmp\n" {
		t.Errorf(".gitignore content: want %q, got %q", "*.tmp\n", string(data))
	}
}

// TestClone_NeitherLayout_ReturnsError verifies that when the cloned repo has
// neither a root SKILL.md nor a <id>/SKILL.md, the installer returns a
// descriptive error naming the repo and the skill ID.
func TestClone_NeitherLayout_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			// Clone produces a dir with no SKILL.md in root or subdir.
			destDir := args[len(args)-1]
			_ = os.WriteFile(filepath.Join(destDir, "README.md"), []byte("# just a readme"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("some-skill", "clone", "JuanCruz/some-skill", false)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error when SKILL.md not found in either layout, got nil")
	}

	// Error must mention the repo and the skill ID.
	if !strings.Contains(err.Error(), "some-skill") {
		t.Errorf("error should mention skill ID %q; got: %v", "some-skill", err)
	}
	if !strings.Contains(err.Error(), "github.com/JuanCruz/some-skill") {
		t.Errorf("error should mention the repo URL; got: %v", err)
	}

	// Skills dir must not have been written.
	if _, statErr := os.Stat(filepath.Join(skillsDir, "some-skill")); statErr == nil {
		t.Error("skills dir was written despite layout error — it must not be touched")
	}
}

// ── C-22: clone from repo subdir (Source.Path) ───────────────────────────────

// TestClone_WithPath_InstallsFromSubdir verifies that when Source.Path is set,
// the SKILL.md is copied from <tempDir>/<path>/ (not the clone root) into
// <skillsDir>/<id>/. This is how third-party skills (find-skill, skill-creator)
// live in monorepos like vercel-labs/skills and anthropics/skills.
func TestClone_WithPath_InstallsFromSubdir(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			// Simulate git clone of a monorepo: SKILL.md lives in
			// <tempDir>/skills/find-skills/ (PLURAL upstream name).
			destDir := args[len(args)-1]
			subDir := filepath.Join(destDir, "skills", "find-skills")
			_ = os.MkdirAll(subDir, 0o755)
			_ = os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte("# find-skill content"), 0o644)
			// Also write a root SKILL.md to prove we do NOT pick it up when path is set.
			_ = os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte("# WRONG root"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	// Harness ID stays "find-skill"; upstream subdir is "skills/find-skills".
	h := model.Harness{
		ID: "find-skill", Name: "find-skill", Type: model.HarnessSkill,
		ThirdParty:   true,
		Source:       &model.Source{Repo: "vercel-labs/skills", Method: "clone", Path: "skills/find-skills"},
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	ins := skill.NewInstaller(runner, nil)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Content must come from the subdir, installed at <skillsDir>/<id>/SKILL.md.
	dest := filepath.Join(skillsDir, "find-skill", "SKILL.md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("SKILL.md not found in skills dir: %v", err)
	}
	if string(data) != "# find-skill content" {
		t.Errorf("SKILL.md content: want %q, got %q", "# find-skill content", string(data))
	}
}

// TestClone_EmptyPath_KeepsRootBehavior is a C-16 regression guard: with an
// empty Source.Path, the root-first / <id>-subdir fallback resolution must be
// unchanged.
func TestClone_EmptyPath_KeepsRootBehavior(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			destDir := args[len(args)-1]
			_ = os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte("# root layout"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("own-skill", "clone", "JuanCruz/own-skill", false) // Path == ""

	ins := skill.NewInstaller(runner, nil)
	if _, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir()); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skillsDir, "own-skill", "SKILL.md"))
	if err != nil {
		t.Fatalf("SKILL.md not found in skills dir: %v", err)
	}
	if string(data) != "# root layout" {
		t.Errorf("SKILL.md content: want %q, got %q", "# root layout", string(data))
	}
}

// ── Types / interface checks ──────────────────────────────────────────────────

// Compile-time check: fakeAdapter must implement skill.AgentAdapter.
var _ skill.AgentAdapter = fakeAdapter{}

// Ensure Result fields are accessible.
func TestResult_Fields(t *testing.T) {
	r := skill.Result{SkillPath: "/a/b", AlreadyInstalled: true}
	if r.SkillPath != "/a/b" {
		t.Error("SkillPath")
	}
	if !r.AlreadyInstalled {
		t.Error("AlreadyInstalled")
	}
}

// Ensure Installer can be constructed (build check).
func TestNewInstaller_NilRunnerAndFS_Panics(t *testing.T) {
	// NewInstaller with nil runner and nil fs should NOT panic at construction.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewInstaller panicked: %v", r)
		}
	}()
	_ = skill.NewInstaller(nil, nil)
}

// Ensure the embed FS type is exported so callers can pass fstest.MapFS.
var _ fs.FS = fstest.MapFS{}
