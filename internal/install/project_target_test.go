package install_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/harness/external"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// projectAdapter is a test double that implements install.AgentAdapter with
// full target-aware path resolution (mimics the Claude project layout).
type projectAdapter struct {
	agent model.Agent
}

func (a projectAdapter) Agent() model.Agent { return a.agent }
func (a projectAdapter) InstructionsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "CLAUDE.md")
}
func (a projectAdapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "skills")
}
func (a projectAdapter) CommandsDir(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "commands")
}
func (a projectAdapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "settings.json")
}
func (a projectAdapter) MCPConfigPath(homeDir, serverName string) string {
	return filepath.Join(homeDir, ".claude", "mcp", serverName+".json")
}
func (a projectAdapter) MCPStrategy() external.MCPStrategy { return external.StrategySeparateFile }
func (a projectAdapter) VariantKey() string                { return string(a.agent) }
func (a projectAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a projectAdapter) PathsFor(base string, t model.InstallTarget) model.AgentPaths {
	// Claude: same .claude/ layout for both machine and project.
	dir := filepath.Join(base, ".claude")
	return model.AgentPaths{
		InstructionsPath: filepath.Join(dir, "CLAUDE.md"),
		SkillsDir:        filepath.Join(dir, "skills"),
		SettingsPath:     filepath.Join(dir, "settings.json"),
		CommandsDir:      filepath.Join(dir, "commands"),
	}.WithMCPConfigFn(func(serverName string) string {
		return filepath.Join(dir, "mcp", serverName+".json")
	})
}

// TestBuildPlanProjectTarget_SnapshotDirUnderProjectRoot verifies that when
// Target=Project, the snapshot dir is <projectRoot>/.jr-stack/backups/install (D4).
func TestBuildPlanProjectTarget_SnapshotDirUnderProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()
	capturedSnapDir := ""

	// Override snapshot to capture the dir passed to it.
	restore := install.SetSnapshotCreate(func(dir string, _ []string) (backup.Manifest, error) {
		capturedSnapDir = dir
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
		HomeDir:     homeDir,
		ProjectRoot: projectRoot,
		Target:      model.Project,
		Registry:    reg,
	}

	plan, err := install.BuildPlan(cat, intent, opts)
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Run Prepare to trigger snapshot.
	for _, step := range plan.Prepare {
		_ = step.Run()
	}

	wantSnapDir := filepath.Join(projectRoot, ".jr-stack", "backups", "install")
	if capturedSnapDir != wantSnapDir {
		t.Errorf("snapshot dir = %q, want %q", capturedSnapDir, wantSnapDir)
	}
}

// TestBuildPlanProjectTarget_PathsUnderProjectRoot verifies that when
// Target=Project, write paths resolve under projectRoot, NOT homeDir.
func TestBuildPlanProjectTarget_PathsUnderProjectRoot(t *testing.T) {
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

	if len(capturedPaths) == 0 {
		t.Fatal("no paths captured; expected at least one write path for config harness")
	}

	// Every write path must be under projectRoot, not homeDir.
	homeDirSlash := filepath.ToSlash(homeDir)
	projectRootSlash := filepath.ToSlash(projectRoot)
	for _, p := range capturedPaths {
		ps := filepath.ToSlash(p)
		if strings.HasPrefix(ps, homeDirSlash) {
			t.Errorf("write path %q resolves under homeDir %q; want under projectRoot %q",
				p, homeDir, projectRoot)
		}
		if !strings.HasPrefix(ps, projectRootSlash) {
			t.Errorf("write path %q does not resolve under projectRoot %q", p, projectRoot)
		}
	}
}
