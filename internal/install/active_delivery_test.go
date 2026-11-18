package install_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/harness/skill"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestActiveDeliveryPackagesSkillAndHandoffForCodexAndClaude(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}

	modes := []struct {
		name       string
		mode       model.InstallMode
		custom     []string
		wantOnHost bool
	}{
		{"lite", model.ModeLite, nil, true},
		{"full", model.ModeFull, nil, true},
		{"custom/delivery-only", model.ModeCustom, []string{"active-delivery"}, true},
		{"custom/sdd-only", model.ModeCustom, []string{"sdd-orchestrator"}, false},
	}
	for _, host := range []model.Agent{model.AgentCodex, model.AgentClaude, model.AgentOpenCode, model.AgentAntigravity} {
		for _, mode := range modes {
			t.Run(string(host)+"/"+mode.name, func(t *testing.T) {
				selected, err := install.SelectHarnesses(cat, install.Intent{Agents: []model.Agent{host}, Mode: mode.mode, Custom: mode.custom})
				if err != nil {
					t.Fatal(err)
				}
				var delivery *model.Harness
				for _, candidate := range selected {
					if candidate.ID == "active-delivery" {
						delivery = &candidate
						break
					}
				}
				if (host != model.AgentCodex && host != model.AgentClaude) || !mode.wantOnHost {
					if delivery != nil {
						t.Fatal("active-delivery was selected outside the requested mode and host")
					}
					return
				}
				if delivery == nil {
					t.Fatal("active-delivery missing from requested mode")
				}
				adapter, err := agents.NewAdapter(host)
				if err != nil {
					t.Fatal(err)
				}
				home := t.TempDir()
				results, err := skill.NewInstaller(nil, assets.SkillsFS).Install(context.Background(), *delivery, []skill.AgentAdapter{adapter}, home, t.TempDir())
				if err != nil || len(results) != 1 {
					t.Fatalf("install active-delivery results = %+v, err = %v", results, err)
				}
				for _, file := range []string{"SKILL.md", "references/handoff.md"} {
					installed, err := os.ReadFile(filepath.Join(adapter.SkillsDir(home), "active-delivery", file))
					if err != nil {
						t.Fatal(err)
					}
					bundled, err := fs.ReadFile(assets.SkillsFS, "skills/active-delivery/"+file)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(installed, bundled) {
						t.Errorf("installed %s differs from packaged file", file)
					}
				}
			})
		}
	}
}
