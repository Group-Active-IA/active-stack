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

func TestActiveRoadmapInstallsCodexSkillAndHelper(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	harness, ok := cat.ByID("active-roadmap")
	if !ok {
		t.Fatal("missing active-roadmap harness")
	}
	adapter, err := agents.NewAdapter(model.AgentCodex)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	_, err = skill.NewInstaller(nil, assets.SkillsFS).Install(context.Background(), harness, []skill.AgentAdapter{adapter}, home, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SKILL.md", "roadmap.py"} {
		if _, err := os.Stat(filepath.Join(adapter.SkillsDir(home), harness.ID, name)); err != nil {
			t.Error(err)
		}
	}
}
