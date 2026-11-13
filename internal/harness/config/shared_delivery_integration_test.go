package config_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

func TestSharedManagedDeliveryInstallBothOrders(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "claude-then-codex"
		if reverse {
			name = "codex-then-claude"
		}
		t.Run(name, func(t *testing.T) {
			home := filepath.Join(t.TempDir(), "profile with spaces")
			ca, cx := claude.NewAdapter(), codex.NewAdapter()
			adapters := []config.AgentAdapter{ca, cx}
			if reverse {
				adapters[0], adapters[1] = adapters[1], adapters[0]
			}
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
			opts := config.InstallOptions{Home: home, DetectVersion: version(claude.VersionCompatible, "2.1.318")}
			res, err := config.InstallWith(h, adapters, home, opts)
			if err != nil {
				t.Fatal(err)
			}
			if res.Modes[model.AgentClaude] != "skill" || res.Modes[model.AgentCodex] != "skill" {
				t.Fatalf("delivery modes not preserved: %+v", res.Modes)
			}
			if len(res.ManagedWrites) == 0 {
				t.Fatal("Codex rollback states lost")
			}
			for _, tc := range []struct{ path, marker string }{
				{ca.InstructionsPath(home), "<!-- active-stack:universal-rules -->"},
				{cx.InstructionsPath(home), "<!-- active-stack:sdd-orchestrator -->"},
			} {
				raw, err := os.ReadFile(tc.path)
				if err != nil || strings.Count(string(raw), tc.marker) != 1 {
					t.Fatalf("client marker %s: %s, %v", tc.marker, raw, err)
				}
			}
			res, err = config.InstallWith(h, adapters, home, opts)
			if err != nil || !res.AllAlready {
				t.Fatalf("reinstall: %+v, %v", res, err)
			}
			h.Toggles = []string{"tdd"}
			if _, err = config.InstallWith(h, adapters, home, opts); err != nil {
				t.Fatal(err)
			}
			for _, a := range []config.AgentAdapter{ca, cx} {
				skills := a.(interface{ SkillsDir(string) string }).SkillsDir(home)
				if _, err = os.Stat(filepath.Join(skills, h.ID, "modules", "tdd.md")); err != nil {
					t.Fatalf("updated %s module: %v", a.Agent(), err)
				}
			}
		})
	}
}

func TestSharedManagedDeliveryVerifyUsesClientMarker(t *testing.T) {
	home := t.TempDir()
	ca, cx := claude.NewAdapter(), codex.NewAdapter()
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	layers, err := config.ComposeLayers(h.Toggles)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = config.InstallLayered(config.LayeredInstall{HomeDir: home, InstructionsPath: ca.InstructionsPath(home), SkillsDir: ca.SkillsDir(home), Skill: layers.Skill, UniversalContent: layers.Universal}); err != nil {
		t.Fatal(err)
	}
	if _, err = config.InstallManagedSkill(h, cx, home); err != nil {
		t.Fatal(err)
	}
	previous := verify.DetectClaudeVersion
	verify.DetectClaudeVersion = version(claude.VersionCompatible, "2.1.318")
	defer func() { verify.DetectClaudeVersion = previous }()
	for _, a := range []verify.Adapter{ca, cx} {
		for _, check := range verify.ChecksForHarness(h, []verify.Adapter{a}, home) {
			if err := check.Run(context.Background()); err != nil {
				t.Errorf("%s verify: %v", a.Agent(), err)
			}
		}
	}
}

type brokenSharedAdapter struct{ config.AgentAdapter }

func (a brokenSharedAdapter) Agent() model.Agent { return model.AgentOpenCode }
func (a brokenSharedAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a brokenSharedAdapter) InstructionsPath(home string) string {
	return filepath.Join(home, "blocked", "AGENTS.md")
}
func (a brokenSharedAdapter) VariantKey() string { return "opencode" }

func TestSharedManagedDeliveryKeepsPartialResultOnLaterFailure(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		name := "claude-then-codex"
		if reverse {
			name = "codex-then-claude"
		}
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			if err := os.WriteFile(filepath.Join(home, "blocked"), []byte("user file"), 0600); err != nil {
				t.Fatal(err)
			}
			ca, cx := claude.NewAdapter(), codex.NewAdapter()
			adapters := []config.AgentAdapter{ca, cx}
			if reverse {
				adapters[0], adapters[1] = adapters[1], adapters[0]
			}
			adapters = append(adapters, brokenSharedAdapter{AgentAdapter: cx})
			res, err := config.InstallWith(model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}, adapters, home, config.InstallOptions{Home: home, DetectVersion: version(claude.VersionCompatible, "2.1.318")})
			if err == nil || res.AllAlready {
				t.Fatalf("later failure not reported: %+v %v", res, err)
			}
			if len(res.Files) == 0 || len(res.ManagedWrites) == 0 || res.Modes[model.AgentClaude] != "skill" || res.Modes[model.AgentCodex] != "skill" {
				t.Fatalf("partial delivery result lost: %+v", res)
			}
		})
	}
}

func TestSharedManagedDeliveryClaudeFallbackDoesNotChangeCodex(t *testing.T) {
	for _, state := range []claude.VersionState{claude.VersionInsufficient, claude.VersionUnknown} {
		name := "insufficient"
		if state == claude.VersionUnknown {
			name = "unknown"
		}
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			ca, cx := claude.NewAdapter(), codex.NewAdapter()
			h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
			res, err := config.InstallWith(h, []config.AgentAdapter{ca, cx}, home, config.InstallOptions{Home: home, DetectVersion: version(state, "2.0.0")})
			if err != nil {
				t.Fatal(err)
			}
			if res.Modes[model.AgentClaude] != "legacy" || res.Modes[model.AgentCodex] != "skill" || len(res.Warnings) == 0 {
				t.Fatalf("per-client fallback lost: %+v", res)
			}
			if _, err := os.Stat(filepath.Join(ca.SkillsDir(home), h.ID, "SKILL.md")); !os.IsNotExist(err) {
				t.Fatalf("Claude fallback installed a native skill: %v", err)
			}
			if err := config.VerifyManagedSkill(h, cx, home); err != nil {
				t.Fatalf("Codex changed by Claude fallback: %v", err)
			}
		})
	}
}
