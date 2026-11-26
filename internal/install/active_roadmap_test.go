package install_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestActiveRoadmapInstallsSkillAndHelper(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, agent := range []model.Agent{model.AgentCodex, model.AgentClaude} {
		t.Run(string(agent), func(t *testing.T) {
			for _, id := range []string{"active-roadmap", "active-delivery"} {
				harness, ok := cat.ByID(id)
				if !ok {
					t.Fatalf("missing %s harness", id)
				}
				adapter, err := agents.NewAdapter(agent)
				if err != nil {
					t.Fatal(err)
				}
				home := t.TempDir()
				_, err = skill.NewInstaller(nil, assets.SkillsFS).Install(context.Background(), harness, []skill.AgentAdapter{adapter}, home, t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				files := map[string][]string{"active-roadmap": {"SKILL.md", "roadmap.py", "references/clickup-actions.md"}, "active-delivery": {"SKILL.md", "references/handoff.md", "usage.py", "references/usage-accounting.md"}}[id]
				for _, name := range files {
					if _, err := os.Stat(filepath.Join(adapter.SkillsDir(home), harness.ID, filepath.FromSlash(name))); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}
