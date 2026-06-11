package install_test

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// TestCollectWritePaths_ProjectTarget_ConfigHarness verifies that when the
// install target is Project, the snapshot for a config harness captures the
// instructions path under the project root (not homeDir).
func TestCollectWritePaths_ProjectTarget_ConfigHarness(t *testing.T) {
	projectRoot := filepath.FromSlash("/proj/myapp")
	homeDir := filepath.FromSlash("/home/testuser")
	var capturedPaths []string

	restore := install.SetSnapshotCreate(func(_ string, paths []string) (backup.Manifest, error) {
		capturedPaths = paths
		return backup.Manifest{}, nil
	})
	defer restore()

	restoreConfig := install.SetConfigInstallFn(func(_ model.Harness, _ interface{}, _ string) error {
		return nil
	})
	defer restoreConfig()

	h := model.Harness{
		ID:           "cfg-h",
		Name:         "Cfg H",
		Type:         model.HarnessConfig,
		Toggles:      []string{},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: projectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:       homeDir,
		ProjectRoot:   projectRoot,
		Target:        model.Project,
		Registry:      reg,
		// NoSelfInstall=true: this test checks harness paths only.
		// The self-install target is system-level (not under projectRoot).
		NoSelfInstall: true,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	for _, step := range plan.Prepare {
		_ = step.Run()
	}

	// All paths must be under projectRoot, not homeDir.
	homeDirSlash := filepath.ToSlash(homeDir)
	projectRootSlash := filepath.ToSlash(projectRoot)

	if len(capturedPaths) == 0 {
		t.Fatal("no paths captured for config harness")
	}
	for _, p := range capturedPaths {
		ps := filepath.ToSlash(p)
		if strings.HasPrefix(ps, homeDirSlash) {
			t.Errorf("path %q resolves under homeDir; want under projectRoot", p)
		}
		if !strings.HasPrefix(ps, projectRootSlash) {
			t.Errorf("path %q not under projectRoot %q", p, projectRoot)
		}
	}
}

// TestCollectWritePaths_ProjectTarget_SkillHarness_DirHint verifies that when
// the install target is Project, the skills dir is captured under the project
// root AND recorded as a DirHint (required for dir-aware rollback, D4).
func TestCollectWritePaths_ProjectTarget_SkillHarness_DirHint(t *testing.T) {
	projectRoot := filepath.FromSlash("/proj/myapp")
	homeDir := filepath.FromSlash("/home/testuser")
	var capturedPaths []string
	var capturedDirHints map[string]bool

	restoreHints := install.SetSnapshotCreateWithHints(func(_ string, paths []string, hints map[string]bool) (backup.Manifest, error) {
		capturedPaths = paths
		capturedDirHints = hints
		return backup.Manifest{}, nil
	})
	defer restoreHints()

	restoreSkill := install.SetSkillInstallFn(func(_ interface{}, _ fs.FS, _ context.Context, _ model.Harness, _ []skill.AgentAdapter, _, _ string) ([]skill.Result, error) {
		return nil, nil
	})
	defer restoreSkill()

	h := model.Harness{
		ID:           "skill-h",
		Name:         "Skill H",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/repo", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: projectAdapter{agent: model.AgentClaude},
	}}
	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
	}
	opts := install.Options{
		HomeDir:     homeDir,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	for _, step := range plan.Prepare {
		_ = step.Run()
	}

	projectRootSlash := filepath.ToSlash(projectRoot)

	// Skills dir must be under projectRoot.
	foundSkillsDir := false
	for _, p := range capturedPaths {
		ps := filepath.ToSlash(p)
		if strings.Contains(ps, "skills") {
			foundSkillsDir = true
			if !strings.HasPrefix(ps, projectRootSlash) {
				t.Errorf("skills dir %q not under projectRoot %q", p, projectRoot)
			}
		}
	}
	if !foundSkillsDir {
		t.Errorf("no skills dir found in captured paths: %v", capturedPaths)
	}

	// The skills dir must be in DirHints (for RemoveAll on rollback).
	if capturedDirHints == nil {
		t.Fatal("DirHints not captured — SetSnapshotCreateWithHints may not be wired")
	}
	foundHint := false
	for p := range capturedDirHints {
		if strings.Contains(filepath.ToSlash(p), "skills") {
			foundHint = true
		}
	}
	if !foundHint {
		t.Errorf("skills dir not in DirHints; DirHints = %v", capturedDirHints)
	}
}
