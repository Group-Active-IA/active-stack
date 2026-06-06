package uninstall_test

import (
	"path/filepath"
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// ─────────────────────────────────────────────────────────────────
// BuildPlan tests
// ─────────────────────────────────────────────────────────────────

// TestBuildPlanLiteModeSelectsLiteHarnesses verifies that Lite mode selects
// only harnesses whose InstallModes includes "lite".
func TestBuildPlanLiteModeSelectsLiteHarnesses(t *testing.T) {
	liteH := model.Harness{
		ID:           "lite-h",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	fullH := model.Harness{
		ID:           "full-only",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{liteH, fullH}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan)
	if !containsStepID(ids, "lite-h") {
		t.Errorf("expected lite-h in plan, got %v", ids)
	}
	if containsStepID(ids, "full-only") {
		t.Errorf("full-only should not be in Lite plan, got %v", ids)
	}
}

// TestBuildPlanCustomModeUsesCustomList verifies that Custom mode uses the
// explicit list from Intent.Custom.
func TestBuildPlanCustomModeUsesCustomList(t *testing.T) {
	h1 := model.Harness{
		ID:           "h1",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeFull},
	}
	h2 := model.Harness{
		ID:           "h2",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h1, h2}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeCustom,
		Custom:   []string{"h1"}, // only h1
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan)
	if !containsStepID(ids, "h1") {
		t.Errorf("expected h1 in plan, got %v", ids)
	}
	if containsStepID(ids, "h2") {
		t.Errorf("h2 should not be in plan, got %v", ids)
	}
}

// TestBuildPlanAgentFiltering verifies that only harnesses supporting the
// selected agent are included in the plan.
func TestBuildPlanAgentFiltering(t *testing.T) {
	claudeOnly := model.Harness{
		ID:           "claude-cfg",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude},
	}
	allAgents := model.Harness{
		ID:           "universal-cfg",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		// empty Agents = all agents
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{claudeOnly, allAgents}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentOpenCode: fakeAdapter{agent: model.AgentOpenCode, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentOpenCode},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := applyStepIDs(plan)
	if containsStepID(ids, "claude-cfg") {
		t.Errorf("claude-cfg should not appear for opencode agent, got %v", ids)
	}
	if !containsStepID(ids, "universal-cfg") {
		t.Errorf("universal-cfg should appear for all agents, got %v", ids)
	}
}

// TestBuildPlanConfigHarnessUsesMarkerRemovalStep verifies that a config
// harness (non-permissions) maps to a marker removal step.
func TestBuildPlanConfigHarnessUsesMarkerRemovalStep(t *testing.T) {
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := rawStepIDs(plan)
	// Config harness should produce a "marker:" prefixed step.
	found := false
	for _, id := range ids {
		if len(id) > 7 && id[:7] == "marker:" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected marker: step for config harness, got %v", ids)
	}
}

// TestBuildPlanSkillHarnessUsesSkillRemovalStep verifies that a skill harness
// maps to a skill-removal step.
func TestBuildPlanSkillHarnessUsesSkillRemovalStep(t *testing.T) {
	h := model.Harness{
		ID:           "my-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/my-skill", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	ids := rawStepIDs(plan)
	found := false
	const skillPrefix = "skill-removal:"
	for _, id := range ids {
		if len(id) >= len(skillPrefix) && id[:len(skillPrefix)] == skillPrefix {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected skill-removal: step for skill harness, got %v", ids)
	}
}

// TestBuildPlanExternalHarnessIsSkippedNoop verifies that an external harness
// maps to an explicit no-op skip step.
func TestBuildPlanExternalHarnessIsSkippedNoop(t *testing.T) {
	h := model.Harness{
		ID:           "context7",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "mcp"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// External step should be present (recorded no-op), not absent.
	ids := rawStepIDs(plan)
	found := false
	for _, id := range ids {
		if len(id) > 14 && id[:14] == "external-skip:" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected external-skip: step for external harness, got %v", ids)
	}

	// Executing it should not error.
	for _, step := range plan.Apply {
		if err := step.Run(); err != nil {
			t.Errorf("external skip step Run() error = %v", err)
		}
	}
}

// TestBuildPlanSnapshotStepInPrepare verifies that the uninstall-time snapshot
// step is present in the Prepare stage and comes before all Apply steps.
func TestBuildPlanSnapshotStepInPrepare(t *testing.T) {
	h := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap"}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	if len(plan.Prepare) == 0 {
		t.Fatal("Prepare stage must have at least one step (snapshot)")
	}
	snapshotID := plan.Prepare[0].ID()
	if snapshotID != "uninstall-snapshot" {
		t.Errorf("first Prepare step ID = %q, want %q", snapshotID, "uninstall-snapshot")
	}
}

// TestBuildPlanSnapshotCapturesApplyPaths verifies that collectUninstallPaths
// covers exactly the paths the Apply steps will touch (parity with install's
// collectWritePaths logic).
func TestBuildPlanSnapshotCapturesApplyPaths(t *testing.T) {
	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	expectedInstrPath := adapter.InstructionsPath(homeDir)
	expectedSkillPath := filepath.Join(adapter.SkillsDir(homeDir), "my-skill")
	expectedSettingsPath := adapter.SettingsPath(homeDir)

	var gotPaths []string
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		gotPaths = paths
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	configH := model.Harness{
		ID:           "sdd-orchestrator",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	skillH := model.Harness{
		ID:           "my-skill",
		Type:         model.HarnessSkill,
		Source:       &model.Source{Repo: "owner/my-skill", Method: "clone"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	permH := model.Harness{
		ID:           "permissions",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	cat := &fakeCatalog{harnesses: []model.Harness{configH, skillH, permH}}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: adapter,
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	// Execute Prepare to trigger snapshot.
	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatalf("prepare step error = %v", err)
		}
	}

	pathSet := make(map[string]bool, len(gotPaths))
	for _, p := range gotPaths {
		pathSet[p] = true
	}

	if !pathSet[expectedInstrPath] {
		t.Errorf("instructions path %q not in snapshot paths %v", expectedInstrPath, gotPaths)
	}
	if !pathSet[expectedSkillPath] {
		t.Errorf("skill path %q not in snapshot paths %v", expectedSkillPath, gotPaths)
	}
	if !pathSet[expectedSettingsPath] {
		t.Errorf("settings path %q not in snapshot paths %v", expectedSettingsPath, gotPaths)
	}
}

// TestBuildPlanEmptySelectionYieldsEmptyPlan verifies that when no harnesses
// are selected, an empty plan (no Apply steps) is returned without error.
func TestBuildPlanEmptySelectionYieldsEmptyPlan(t *testing.T) {
	// No harnesses in catalog.
	cat := &fakeCatalog{harnesses: nil}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	if len(plan.Apply) != 0 {
		t.Errorf("expected empty Apply steps for empty catalog, got %d steps", len(plan.Apply))
	}
}

// TestBuildPlanUnknownAgentReturnsError verifies that requesting an agent not
// in the registry returns an error.
func TestBuildPlanUnknownAgentReturnsError(t *testing.T) {
	h := model.Harness{
		ID:           "h1",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	// Registry has claude, but we request opencode.
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	_, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentOpenCode},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err == nil {
		t.Error("expected error for unregistered agent")
	}
}

// TestBuildPlanCommandHarnessUsesCommandRemovalStep verifies that a command
// harness maps to a command-removal step (not an error) and is included in
// the plan. This is the regression test for the HarnessCommand bug: the engine
// previously crashed with "unknown harness type" when BuildPlan encountered a
// command harness (e.g. starter-add-command in lite mode).
//
// RED: fails until uninstall.AgentAdapter gets CommandsDir+VariantKey and
// buildUninstallStep handles model.HarnessCommand.
func TestBuildPlanCommandHarnessUsesCommandRemovalStep(t *testing.T) {
	h := model.Harness{
		ID:           "starter-add-command",
		Type:         model.HarnessCommand,
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	homeDir := t.TempDir()
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude, homeDir: homeDir},
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() must not error for command harness, got: %v", err)
	}

	ids := rawStepIDs(plan)
	const cmdPrefix = "command-removal:"
	found := false
	for _, id := range ids {
		if len(id) >= len(cmdPrefix) && id[:len(cmdPrefix)] == cmdPrefix {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected command-removal: step for command harness, got %v", ids)
	}
}

// TestBuildPlanCommandHarnessSnapshotsCaptureCommandPaths verifies that
// collectUninstallPaths includes the command file paths when a command harness
// is in the plan. The snapshot must cover the files that commandRemovalStep
// will delete.
func TestBuildPlanCommandHarnessSnapshotsCaptureCommandPaths(t *testing.T) {
	h := model.Harness{
		ID:           "starter-add-command",
		Type:         model.HarnessCommand,
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}

	homeDir := t.TempDir()
	adapter := fakeAdapter{agent: model.AgentClaude, homeDir: homeDir}

	var gotPaths []string
	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		gotPaths = paths
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{h}}
	reg := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{
		model.AgentClaude: adapter,
	}}

	plan, err := uninstall.BuildPlan(cat, uninstall.Intent{
		Agents:   []model.Agent{model.AgentClaude},
		Mode:     model.ModeLite,
		Strategy: uninstall.StrategyTargeted,
	}, buildUninstallOptions(homeDir, reg))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	for _, step := range plan.Prepare {
		if err := step.Run(); err != nil {
			t.Fatalf("prepare step error = %v", err)
		}
	}

	// The adapter's CommandsDir + RelPathForVariant("claude") = expected path.
	expectedPath := filepath.Join(adapter.CommandsDir(homeDir), "jr", "starter-add.md")
	found := false
	for _, p := range gotPaths {
		if p == expectedPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected command path %q in snapshot paths %v", expectedPath, gotPaths)
	}
}

// ─────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────

// applyStepIDs returns the harness IDs stripped of their prefix.
func applyStepIDs(plan uninstall.Plan) []string {
	var ids []string
	for _, s := range plan.Apply {
		id := s.ID()
		for _, prefix := range []string{"marker:", "skill-removal:", "permissions-removal:", "external-skip:", "restore-from-backup", "command-removal:"} {
			if id == prefix {
				id = prefix
				break
			}
			if len(id) > len(prefix) && id[:len(prefix)] == prefix {
				id = id[len(prefix):]
				break
			}
		}
		ids = append(ids, id)
	}
	return ids
}

// rawStepIDs returns all Apply step IDs as-is.
func rawStepIDs(plan uninstall.Plan) []string {
	var ids []string
	for _, s := range plan.Apply {
		ids = append(ids, s.ID())
	}
	return ids
}

func containsStepID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
