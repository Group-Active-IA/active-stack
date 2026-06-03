package install_test

import (
	"testing"

	"github.com/JuanCruzRobledo/jr-stack/internal/backup"
	perminstaller "github.com/JuanCruzRobledo/jr-stack/internal/harness/config/permissions"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
)

// tierTestPermissionsHarness returns a minimal permissions harness for tier wiring tests.
func tierTestPermissionsHarness() model.Harness {
	return model.Harness{
		ID:           "permissions",
		Name:         "Permissions",
		Type:         model.HarnessConfig,
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
		Agents:       []model.Agent{model.AgentClaude, model.AgentOpenCode},
	}
}

// runPlanWithTier builds and executes a plan with the given intent tier and
// returns the tier that was received by the permissions installer.
func runPlanWithTier(t *testing.T, intentTier model.PermissionTier) model.PermissionTier {
	t.Helper()

	var capturedTier model.PermissionTier

	restorePerm := install.SetPermissionsInstallFn(func(
		_ string,
		_ []perminstaller.PermissionsAdapter,
		tier model.PermissionTier,
	) (perminstaller.Result, error) {
		capturedTier = tier
		return perminstaller.Result{}, nil
	})
	defer restorePerm()

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	cat := &fakeCatalog{harnesses: []model.Harness{tierTestPermissionsHarness()}}
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeAdapter{agent: model.AgentClaude},
	}}
	homeDir := t.TempDir()

	intent := install.Intent{
		Agents: []model.Agent{model.AgentClaude},
		Mode:   model.ModeLite,
		Tier:   intentTier,
	}

	plan, err := install.BuildPlan(cat, intent, buildOptions(homeDir, reg, nil))
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	return capturedTier
}

// TestIntentTierZeroValueNormalizesToBalanceado verifies that an Intent with
// empty Tier (zero-value) resolves to TierBalanceado at the installer.
// Spec: tui-install — "El intent usa balanceado por defecto".
func TestIntentTierZeroValueNormalizesToBalanceado(t *testing.T) {
	got := runPlanWithTier(t, "") // zero-value

	if got == model.TierBypass {
		t.Fatal("zero-value tier must NEVER produce bypass — got bypassPermissions")
	}
	if got != model.TierBalanceado {
		t.Errorf("zero-value tier: capturedTier = %q, want %q", got, model.TierBalanceado)
	}
}

// TestIntentTierBypassPropagates verifies that an explicit bypass tier flows
// from Intent through BuildPlan to the permissions installer.
func TestIntentTierBypassPropagates(t *testing.T) {
	got := runPlanWithTier(t, model.TierBypass)

	if got != model.TierBypass {
		t.Errorf("explicit bypass tier: capturedTier = %q, want %q", got, model.TierBypass)
	}
}

// TestIntentTierEstrictoPropagates verifies that an explicit estricto tier
// flows through to the permissions installer.
func TestIntentTierEstrictoPropagates(t *testing.T) {
	got := runPlanWithTier(t, model.TierEstricto)

	if got != model.TierEstricto {
		t.Errorf("estricto tier: capturedTier = %q, want %q", got, model.TierEstricto)
	}
}
