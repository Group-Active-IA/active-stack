package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/update"
)

func TestParseUpdateFlags(t *testing.T) {
	flags, err := parseUpdateFlags([]string{"--skills-only", "--project", "C:/work/app"})
	if err != nil {
		t.Fatal(err)
	}
	if flags.Check || !flags.SkillsOnly || flags.Project != "C:/work/app" {
		t.Fatalf("flags=%+v", flags)
	}
	if _, err := parseUpdateFlags([]string{"--unknown"}); err == nil {
		t.Fatal("unknown flag accepted")
	}
	if _, err := parseUpdateFlags([]string{"--check", "--skills-only"}); err == nil {
		t.Fatal("conflicting flags accepted")
	}
}

func TestInstallArgsPreserveCustomSelection(t *testing.T) {
	got := installArgs(model.AgentCodex, update.AgentSelection{
		Mode: model.ModeCustom, Custom: []string{"design-mockups", "design-system"}, Tier: model.TierEstricto,
	}, "C:/Users/example")
	want := []string{"install", "--headless", "--yes", "--no-self-install", "--home", "C:/Users/example", "--agent", "codex", "--mode", "custom", "--tier", "estricto", "--custom", "design-mockups,design-system"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args=%v, want %v", got, want)
	}
}

func TestShouldUpdateBinaryNeverDowngrades(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v0.3.0", "v0.4.0", true},
		{"v0.4.0", "v0.4.0", false},
		{"v0.5.0", "v0.4.0", false},
		{"v0.3.0-dirty", "v0.4.0", false},
		{"dev", "v0.4.0", false},
	}
	for _, tc := range cases {
		got := shouldUpdateBinary(tc.current, tc.latest)
		if got != tc.want {
			t.Errorf("shouldUpdateBinary(%q,%q)=%v", tc.current, tc.latest, got)
		}
	}
}

func TestInstallRunsReapplyPresetAndExtraCustomHarnesses(t *testing.T) {
	runs := installRuns(model.AgentCodex, update.AgentSelection{Mode: model.ModeFull, Custom: []string{"extra"}}, "C:/home")
	if len(runs) != 2 || runs[0][9] != "full" || runs[1][9] != "custom" {
		t.Fatalf("runs=%v", runs)
	}
}

func TestUpdateHelpExplainsSafeModes(t *testing.T) {
	var out bytes.Buffer
	if err := runUpdate(context.Background(), []string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("--skills-only")) || !bytes.Contains(out.Bytes(), []byte("--check")) {
		t.Fatalf("help=%q", out.String())
	}
}

func TestUpdateSkipsUnownedSkillWithCatalogName(t *testing.T) {
	home := t.TempDir()
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := reg.Get(model.AgentCodex)
	path := filepath.Join(adapter.SkillsDir(home), "design-system")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(path, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("personal content"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := refreshSkills(context.Background(), home, "", false, &out); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "personal content" || !bytes.Contains(out.Bytes(), []byte("omitidas: 1")) {
		t.Fatalf("skill=%q, output=%s", got, out.String())
	}
	if err := refreshSkills(context.Background(), home, "", true, &out); err != nil {
		t.Fatal(err)
	}
	state, err := update.LoadState(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ManagedSkills[model.AgentCodex]) != 1 {
		t.Fatalf("adopted paths=%v", state.ManagedSkills)
	}
}

func TestFinishFlagsRequireParentAndTarget(t *testing.T) {
	if _, err := parseUpdateFlags([]string{"--finish"}); err == nil {
		t.Fatal("unsafe finish accepted")
	}
	if _, err := parseUpdateFlags([]string{"--cleanup-bin", "C:/Temp/active-stack.exe"}); err == nil {
		t.Fatal("unsafe helper cleanup accepted")
	}
	flags, err := parseUpdateFlags([]string{"--finish", "--parent-pid", "123", "--target", "C:/active-stack.exe"})
	if err != nil || !flags.Finish {
		t.Fatalf("flags=%+v, err=%v", flags, err)
	}
}

func TestSyncRecordedInstallRejectsUnownedCollisionBeforeReplay(t *testing.T) {
	home := t.TempDir()
	if err := update.SaveIntent(home, install.Intent{Agents: []model.Agent{model.AgentCodex}, Mode: model.ModeFull}); err != nil {
		t.Fatal(err)
	}
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := reg.Get(model.AgentCodex)
	path := filepath.Join(adapter.SkillsDir(home), "design-system")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(path, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("personal content"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = syncRecordedInstall(ctx, home, "", false, &out)
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("--adopt-legacy-skills")) {
		t.Fatalf("expected actionable collision error, got %v; output=%s", err, out.String())
	}
	if bytes.Contains(out.Bytes(), []byte("Reaplicando")) {
		t.Fatalf("replay started before collision check: %s", out.String())
	}
	got, err := os.ReadFile(skillPath)
	if err != nil || string(got) != "personal content" {
		t.Fatalf("personal skill changed: %q, %v", got, err)
	}
}

func TestReplayPreflightAllowsOwnedAndExplicitlyAdoptedSkills(t *testing.T) {
	for _, adopted := range []bool{false, true} {
		t.Run(map[bool]string{false: "owned", true: "explicit adoption"}[adopted], func(t *testing.T) {
			home := t.TempDir()
			reg, err := agents.NewDefaultRegistry()
			if err != nil {
				t.Fatal(err)
			}
			adapter, _ := reg.Get(model.AgentCodex)
			path := filepath.Join(adapter.SkillsDir(home), "design-system")
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			state := update.State{Agents: map[model.Agent]update.AgentSelection{model.AgentCodex: {Mode: model.ModeFull}}, ManagedSkills: make(map[model.Agent][]string)}
			if !adopted {
				state.ManagedSkills[model.AgentCodex] = []string{path}
			}
			if err := preflightRecordedSkills(home, state, adopted); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUpdateLogCreatesDirectoryAndPreservesPreviousRun(t *testing.T) {
	home := t.TempDir()
	for _, content := range []string{"first failure\n", "next result\n"} {
		log, err := openUpdateLog(home)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := log.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(home, ".active-stack", "logs", "update.log"))
	if err != nil || string(data) != "first failure\nnext result\n" {
		t.Fatalf("log=%q err=%v", data, err)
	}
}
