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

func TestCodexContextModulesRejectMissingOrEmptyInstalledContent(t *testing.T) {
	for _, id := range []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"} {
		for _, scenario := range []string{"missing", "empty", "whitespace"} {
			t.Run(id+"/"+scenario, func(t *testing.T) {
				home := t.TempDir()
				a := managedVerifyAdapter{newFakeAdapter(t, home, model.AgentCodex)}
				a.delivery = model.ConfigDeliveryManagedSkill
				h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
				if _, err := config.InstallManagedSkill(h, a, home); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(a.SkillsDir(home), h.ID, "modules", id+".md")
				var err error
				if scenario == "missing" {
					err = os.Remove(path)
				} else {
					body := ""
					if scenario == "whitespace" {
						body = " \n\t\r\n"
					}
					err = os.WriteFile(path, []byte(body), 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
				checks := verify.ChecksForHarness(h, []verify.Adapter{a}, home)
				if len(checks) != 1 {
					t.Fatalf("got %d checks", len(checks))
				}
				if err = checks[0].Run(context.Background()); err == nil {
					t.Fatalf("%s module %s accepted", scenario, id)
				}
				if _, err = config.InstallManagedSkill(h, a, home); err == nil {
					t.Fatal("edited or removed module was silently adopted")
				}
			})
		}
	}
}
