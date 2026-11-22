package update_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/update"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// [EQ] Fase 7 (R-20) T-F7-5: verify reconoce una skill ya actualizada por
// `update --skills-only`, y sigue reconociendola en ambos agentes que la reciben.
func TestVerifyRecognizesRefreshedSkills(t *testing.T) {
	home := t.TempDir()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"active-delivery", "active-roadmap"} {
		h, ok := cat.ByID(id)
		if !ok {
			t.Fatalf("%s missing", id)
		}
		var adapters []verify.Adapter
		for _, agent := range []model.Agent{model.AgentClaude, model.AgentCodex} {
			ad, _ := reg.Get(agent)
			dir := filepath.Join(ad.SkillsDir(home), id)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("old"), 0o644); err != nil {
				t.Fatal(err)
			}
			adapters = append(adapters, ad)
		}
		targets, err := update.ScanInstalledSkills(cat, reg, home)
		if err != nil {
			t.Fatal(err)
		}
		files := fstest.MapFS{"skills/" + id + "/SKILL.md": &fstest.MapFile{Data: []byte("# new content")}}
		update.RefreshInstalledSkills(context.Background(), targets, reg, home, filepath.Join(home, "b"), nil, files)

		checks := verify.ChecksForHarness(h, adapters, home)
		if len(checks) == 0 {
			t.Fatalf("%s: no checks", id)
		}
		for _, c := range checks {
			if err := c.Run(context.Background()); err != nil {
				t.Errorf("%s: %s failed after refresh: %v", id, c.ID, err)
			}
		}
		for _, a := range adapters {
			raw, _ := os.ReadFile(filepath.Join(a.SkillsDir(home), id, "SKILL.md"))
			if string(raw) != "# new content" {
				t.Errorf("%s/%s not refreshed: %q", a.Agent(), id, raw)
			}
		}
	}
}
