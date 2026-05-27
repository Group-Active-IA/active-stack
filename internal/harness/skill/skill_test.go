package skill_test

// ──────────────────────────────────────────────────────────────────────────────
// Tests for internal/harness/skill
//
// TDD order: all tests written RED first, then implementation makes them GREEN.
// No real git/npx commands are executed — all exec is mocked via the Runner
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

	"github.com/JuanCruzRobledo/jr-stack/internal/harness/skill"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
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

func (f fakeAdapter) Agent() model.Agent           { return f.agent }
func (f fakeAdapter) SkillsDir(homeDir string) string { return f.skillsDir }

// stubRunner records the command that was run and optionally creates a file
// to simulate a successful installation side-effect.
type stubRunner struct {
	called   [][]string
	err      error
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

func TestInstaller_NPX_CallsRunner(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			// Simulate npx writing SKILL.md to the skills dir.
			// Find the --skills-dir flag value.
			dir := ""
			for i, a := range args {
				if a == "--skills-dir" && i+1 < len(args) {
					dir = args[i+1]
				}
			}
			if dir == "" {
				return
			}
			skillDir := filepath.Join(dir, "third-skill")
			_ = os.MkdirAll(skillDir, 0o755)
			_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# third"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("third-skill", "npx", "vercel/skills", true)

	ins := skill.NewInstaller(runner, nil)
	results, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Verify npx was called with --skills-dir
	if len(runner.called) == 0 {
		t.Fatal("expected runner to be called for npx")
	}
	args := runner.called[0]
	if args[0] != "npx" {
		t.Errorf("expected first arg %q, got %q", "npx", args[0])
	}
	hasFlag := false
	for _, a := range args {
		if a == "--skills-dir" {
			hasFlag = true
		}
	}
	if !hasFlag {
		t.Errorf("expected --skills-dir flag in args %v", args)
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

// ── Task 7.3: npx.go ─────────────────────────────────────────────────────────

func TestNPX_SkillsDirFlag(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")

	runner := &stubRunner{
		sideEffect: func(args []string) {
			dir := ""
			for i, a := range args {
				if a == "--skills-dir" && i+1 < len(args) {
					dir = args[i+1]
				}
			}
			if dir == "" {
				return
			}
			d := filepath.Join(dir, "npx-skill")
			_ = os.MkdirAll(d, 0o755)
			_ = os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("# npx"), 0o644)
		},
	}

	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("npx-skill", "npx", "vercel/skills", true)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}
	args := runner.called[0]
	// Verify: npx skills add <id> --skills-dir <dir>
	if args[0] != "npx" {
		t.Errorf("first arg: want npx, got %q", args[0])
	}
	// --skills-dir must be followed by the actual skills dir path
	for i, a := range args {
		if a == "--skills-dir" && i+1 < len(args) {
			if args[i+1] != skillsDir {
				t.Errorf("--skills-dir value: want %q, got %q", skillsDir, args[i+1])
			}
		}
	}
}

func TestNPX_NonZeroExit_ReturnsError(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	runner := &stubRunner{err: errors.New("exit status 1")}
	adapter := fakeAdapter{agent: model.AgentClaude, skillsDir: skillsDir}
	h := makeHarness("npx-skill", "npx", "vercel/skills", true)

	ins := skill.NewInstaller(runner, nil)
	_, err := ins.Install(context.Background(), h, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err == nil {
		t.Fatal("expected error on non-zero npx exit, got nil")
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
