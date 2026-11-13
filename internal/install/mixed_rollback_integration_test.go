package install_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
)

// This triangulates the mixed-client dispatcher after its recorded RED cycle.
// A later verify failure must not discard Codex rollback guards just because
// Claude and Codex share one config step and one operation snapshot.
func TestSharedManagedDeliveryRollbackPreservesConcurrentCodexEdits(t *testing.T) {
	restore := install.SetClaudeDetectFn(func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318", Min: claude.MinVersion}
	})
	defer restore()
	for _, reverse := range []bool{false, true} {
		name := "claude-then-codex"
		if reverse {
			name = "codex-then-claude"
		}
		t.Run(name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), "mixed profile with spaces")
			ca, cx := claude.NewAdapter(), codex.NewAdapter()
			claudeOriginal := []byte("# Claude user rules\n")
			codexOriginal := []byte("# Codex user rules\n")
			for path, content := range map[string][]byte{
				ca.InstructionsPath(home): claudeOriginal,
				cx.InstructionsPath(home): codexOriginal,
			} {
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, content, 0600); err != nil {
					t.Fatal(err)
				}
			}
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeLite}, Toggles: []string{"tdd"}}
			order := []model.Agent{model.AgentClaude, model.AgentCodex}
			if reverse {
				order[0], order[1] = order[1], order[0]
			}
			reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentClaude: ca, model.AgentCodex: cx}}
			note := filepath.Join(cx.SkillsDir(home), h.ID, "user-note.md")
			var codexEdited []byte
			verifyReached := false
			hook := func() error {
				verifyReached = true
				if err := config.VerifyManagedSkill(h, cx, home); err != nil {
					t.Fatal(err)
				}
				claudeInstalled, err := os.ReadFile(ca.InstructionsPath(home))
				if err != nil || !strings.Contains(string(claudeInstalled), "<!-- active-stack:universal-rules -->") {
					t.Fatalf("Claude delivery did not reach verify: %s %v", claudeInstalled, err)
				}
				codexInstalled, err := os.ReadFile(cx.InstructionsPath(home))
				if err != nil {
					t.Fatal(err)
				}
				codexEdited = append(codexInstalled, []byte("\n# concurrent Codex user edit\n")...)
				if err := os.WriteFile(cx.InstructionsPath(home), codexEdited, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(note, []byte("user note after install\n"), 0600); err != nil {
					t.Fatal(err)
				}
				return errors.New("mixed late verify failure")
			}
			opts := buildOptions(home, reg, hook)
			opts.NoSelfInstall = true
			plan, err := install.BuildPlan(&fakeCatalog{harnesses: []model.Harness{h}}, install.Intent{Agents: order, Mode: model.ModeLite}, opts)
			if err != nil {
				t.Fatal(err)
			}
			result := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan)
			if !verifyReached || result.Err == nil || result.Apply.Err == nil || !strings.Contains(result.Apply.Err.Error(), "mixed late verify failure") {
				t.Fatalf("later failure was not reported: %+v", result)
			}
			if result.Rollback.Err == nil || !strings.Contains(result.Rollback.Err.Error(), "preserved") || !strings.Contains(result.Rollback.Err.Error(), cx.InstructionsPath(home)) {
				t.Fatalf("rollback did not report the preserved edit conflict: %v", result.Rollback.Err)
			}
			got, err := os.ReadFile(cx.InstructionsPath(home))
			if err != nil || !bytes.Equal(got, codexEdited) {
				t.Fatalf("concurrent Codex edit lost: %s %v", got, err)
			}
			got, err = os.ReadFile(note)
			if err != nil || string(got) != "user note after install\n" {
				t.Fatalf("concurrent Codex note lost: %s %v", got, err)
			}
			got, err = os.ReadFile(ca.InstructionsPath(home))
			if err != nil || !bytes.Equal(got, claudeOriginal) {
				t.Fatalf("unchanged Claude instructions were not restored: %s %v", got, err)
			}
			for _, a := range []install.AgentAdapter{ca, cx} {
				if _, err := os.Stat(filepath.Join(a.SkillsDir(home), h.ID, "SKILL.md")); !os.IsNotExist(err) {
					t.Fatalf("new managed skill was not rolled back for %s: %v", a.Agent(), err)
				}
			}
		})
	}
}
