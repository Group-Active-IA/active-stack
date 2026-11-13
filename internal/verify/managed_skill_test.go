package verify_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

type managedVerifyAdapter struct{ fakeAdapter }

func (a managedVerifyAdapter) VariantKey() string { return "codex" }

func TestManagedSkillCheckRejectsMissingPolicy(t *testing.T) {
	home := t.TempDir()
	a := managedVerifyAdapter{newFakeAdapter(t, home, model.AgentCodex)}
	a.delivery = model.ConfigDeliveryManagedSkill
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if _, err := config.Install(h, []config.AgentAdapter{a}, home); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(a.SkillsDir(home), h.ID, "agents", "openai.yaml")); err != nil {
		t.Fatal(err)
	}
	checks := verify.ChecksForHarness(h, []verify.Adapter{a}, home)
	if len(checks) != 1 {
		t.Fatalf("got %d checks, want one structural check", len(checks))
	}
	if err := checks[0].Run(context.Background()); err == nil {
		t.Fatal("missing native policy must fail structural verification")
	}
}

func TestManagedSkillChecksStructureAndOwnership(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		path      string
		content   string
		remove    bool
		wantError bool
	}{
		{name: "installed"},
		{name: "user note", path: "notes.md", content: "my note"},
		{name: "missing module", path: "modules/tdd.md", remove: true, wantError: true},
		{name: "implicit policy", path: "agents/openai.yaml", content: "policy:\n  allow_implicit_invocation: true\n", wantError: true},
		{name: "missing ownership", path: ".active-stack-managed.json", remove: true, wantError: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			home := t.TempDir()
			a := managedVerifyAdapter{newFakeAdapter(t, home, model.AgentCodex)}
			a.delivery = model.ConfigDeliveryManagedSkill
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"tdd"}}
			if _, err := config.InstallManagedSkill(h, a, home); err != nil {
				t.Fatal(err)
			}
			if scenario.path != "" {
				path := filepath.Join(a.SkillsDir(home), h.ID, filepath.FromSlash(scenario.path))
				var err error
				if scenario.remove {
					err = os.Remove(path)
				} else {
					err = os.WriteFile(path, []byte(scenario.content), 0o600)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			checks := verify.ChecksForHarness(h, []verify.Adapter{a}, home)
			if len(checks) != 1 {
				t.Fatalf("checks = %d", len(checks))
			}
			if err := checks[0].Run(context.Background()); (err != nil) != scenario.wantError {
				t.Fatalf("structural check error = %v, want error %v", err, scenario.wantError)
			}
		})
	}
}
