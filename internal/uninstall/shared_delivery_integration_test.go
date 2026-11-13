package uninstall_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

func TestSharedManagedDeliveryUninstallPreservesOtherClient(t *testing.T) {
	for _, first := range []model.Agent{model.AgentClaude, model.AgentCodex} {
		t.Run(string(first), func(t *testing.T) {
			home := t.TempDir()
			ca, cx := claude.NewAdapter(), codex.NewAdapter()
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
			layers, err := config.ComposeLayers(nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = config.InstallLayered(config.LayeredInstall{HomeDir: home, InstructionsPath: ca.InstructionsPath(home), SkillsDir: ca.SkillsDir(home), Skill: layers.Skill, UniversalContent: layers.Universal}); err != nil {
				t.Fatal(err)
			}
			if _, err = config.InstallManagedSkill(h, cx, home); err != nil {
				t.Fatal(err)
			}
			registry := &fakeRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: ca, model.AgentCodex: cx}}
			other := model.AgentClaude
			if first == other {
				other = model.AgentCodex
			}
			otherAdapter, _ := registry.Get(other)
			otherPaths := []string{otherAdapter.InstructionsPath(home), filepath.Join(otherAdapter.SkillsDir(home), h.ID, "SKILL.md")}
			before := map[string][]byte{}
			for _, path := range otherPaths {
				before[path], err = os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, a := range []uninstall.AgentAdapter{ca, cx} {
				if a.Agent() == model.AgentCodex {
					if err = os.WriteFile(filepath.Join(a.SkillsDir(home), h.ID, "user-note.md"), []byte("keep"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			plan, err := uninstall.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, uninstall.Intent{Agents: []model.Agent{first}, Mode: model.ModeCustom, Custom: []string{h.ID}, Strategy: uninstall.StrategyTargeted}, buildUninstallOptions(home, registry))
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range append(plan.Prepare, plan.Apply...) {
				if err = step.Run(); err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range otherPaths {
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before[path], after) {
					t.Fatalf("other client content changed at %s: %v", path, err)
				}
			}
			for _, a := range []uninstall.AgentAdapter{ca, cx} {
				_, err := os.Stat(filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md"))
				if a.Agent() == first && !os.IsNotExist(err) {
					t.Fatalf("selected client %s not removed: %v", a.Agent(), err)
				}
				if a.Agent() == other && err != nil {
					t.Fatalf("other client %s was changed: %v", a.Agent(), err)
				}
				if a.Agent() == model.AgentCodex {
					if raw, err := os.ReadFile(filepath.Join(a.SkillsDir(home), h.ID, "user-note.md")); err != nil || string(raw) != "keep" {
						t.Fatalf("user note %s: %s, %v", a.Agent(), raw, err)
					}
				}
			}
		})
	}
}
