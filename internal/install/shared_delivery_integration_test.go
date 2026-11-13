package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

func TestSharedManagedDeliveryPlanRoutesMachineAndProject(t *testing.T) {
	restore := install.SetClaudeDetectFn(func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
	})
	defer restore()
	for _, target := range []model.InstallTarget{model.Machine, model.Project} {
		targetName := "machine"
		if target == model.Project {
			targetName = "project"
		}
		for _, reverse := range []bool{false, true} {
			name := targetName + "/claude-then-codex"
			if reverse {
				name = targetName + "/codex-then-claude"
			}
			t.Run(name, func(t *testing.T) {
				home := filepath.Join(t.TempDir(), "home with spaces")
				project := filepath.Join(t.TempDir(), "project with spaces")
				if err := os.MkdirAll(project, 0700); err != nil {
					t.Fatal(err)
				}
				ca, cx := claude.NewAdapter(), codex.NewAdapter()
				order := []model.Agent{model.AgentClaude, model.AgentCodex}
				if reverse {
					order[0], order[1] = order[1], order[0]
				}
				h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite}, Toggles: []string{"tdd"}}
				reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentClaude: ca, model.AgentCodex: cx}}
				opts := buildOptions(home, reg, nil)
				opts.Target, opts.ProjectRoot, opts.NoSelfInstall = target, project, true
				plan, err := install.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, install.Intent{Agents: order, Mode: model.ModeLite}, opts)
				if err != nil {
					t.Fatal(err)
				}
				res := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
				if res.Err != nil {
					t.Fatal(res.Err)
				}
				base := home
				if target == model.Project {
					base = project
				}
				for _, a := range []install.AgentAdapter{ca, cx} {
					paths := a.PathsFor(base, target)
					raw, err := os.ReadFile(paths.InstructionsPath)
					if err != nil {
						t.Fatal(err)
					}
					marker := "<!-- active-stack:sdd-orchestrator -->"
					if a.Agent() == model.AgentClaude {
						marker = "<!-- active-stack:universal-rules -->"
					}
					if strings.Count(string(raw), marker) != 1 {
						t.Fatalf("%s marker missing in %s", a.Agent(), paths.InstructionsPath)
					}
					_, err = os.Stat(filepath.Join(paths.SkillsDir, h.ID, "SKILL.md"))
					if a.Agent() == model.AgentCodex && target == model.Project {
						if !os.IsNotExist(err) {
							t.Fatalf("Codex Project unexpectedly installed native skill: %v", err)
						}
					} else if err != nil {
						t.Fatalf("%s native skill missing: %v", a.Agent(), err)
					}
				}
			})
		}
	}
}
