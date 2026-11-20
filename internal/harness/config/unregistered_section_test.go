package config_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// R-07 "cierre sin huecos": a config harness that is neither the orchestrator
// nor registered in sectionFragments (engram-cloud before its phase 2 fragment
// exists) must not fall through to the orchestrator path. It must be an
// explicit no-op that writes and replaces nothing.

func dirDigest(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = fmt.Sprintf("%x", sha256.Sum256(b))
		return nil
	})
	if err != nil {
		t.Fatalf("digest %s: %v", root, err)
	}
	return out
}

func realOpts(home string) config.InstallOptions {
	return config.InstallOptions{
		Home: home,
		DetectVersion: func() claude.VersionCheck {
			return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.327", Min: "2.1.0"}
		},
	}
}

func installOrchestrator(t *testing.T, home string, adapters []config.AgentAdapter) {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load: %v", err)
	}
	orch, ok := cat.ByID("sdd-orchestrator")
	if !ok {
		t.Fatal("sdd-orchestrator missing from catalog")
	}
	if _, err := config.InstallWith(orch, adapters, home, realOpts(home)); err != nil {
		t.Fatalf("install sdd-orchestrator: %v", err)
	}
}

// The real catalog entry, real Claude and Codex adapters.
func TestInstall_EngramCloudWithoutFragment_LeavesOrchestratorIntact(t *testing.T) {
	home := t.TempDir()
	adapters := []config.AgentAdapter{claude.NewAdapter(), codex.NewAdapter()}
	installOrchestrator(t, home, adapters)
	before := dirDigest(t, home)

	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	ec, ok := cat.ByID("engram-cloud")
	if !ok {
		t.Fatal("engram-cloud missing from catalog")
	}
	res, err := config.InstallWith(ec, adapters, home, realOpts(home))
	if err != nil {
		t.Fatalf("engram-cloud install must not fail (codex used to): %v", err)
	}
	if len(res.Files) != 0 {
		t.Errorf("engram-cloud must write no files, wrote %v", res.Files)
	}
	if !res.AllAlready {
		t.Error("a no-op must report AllAlready")
	}
	if len(res.Warnings) == 0 {
		t.Error("a no-op must explain itself with a warning")
	}
	after := dirDigest(t, home)
	for k, v := range before {
		if after[k] != v {
			t.Errorf("%s changed after installing engram-cloud", k)
		}
	}
	for k := range after {
		if _, existed := before[k]; !existed {
			t.Errorf("%s was created by installing engram-cloud", k)
		}
	}
}

// Triangulation: an invented config harness without a fragment behaves the same.
func TestInstall_UnknownConfigHarnessWithoutFragment_IsNoOp(t *testing.T) {
	home := t.TempDir()
	adapters := []config.AgentAdapter{claude.NewAdapter(), codex.NewAdapter()}
	installOrchestrator(t, home, adapters)
	before := dirDigest(t, home)

	h := model.Harness{
		ID:     "made-up-config",
		Type:   model.HarnessConfig,
		Agents: []model.Agent{model.AgentClaude, model.AgentCodex},
	}
	res, err := config.InstallWith(h, adapters, home, realOpts(home))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Files) != 0 || !res.AllAlready {
		t.Errorf("want no-op, got files=%v allAlready=%v", res.Files, res.AllAlready)
	}
	if !strings.Contains(strings.Join(res.Warnings, " "), "made-up-config") {
		t.Errorf("warning must name the harness, got %v", res.Warnings)
	}
	after := dirDigest(t, home)
	if len(after) != len(before) {
		t.Errorf("file set changed: %d -> %d", len(before), len(after))
	}
	for k, v := range before {
		if after[k] != v {
			t.Errorf("%s changed", k)
		}
	}
}

// The no-op must also hold on a machine where the orchestrator was never
// installed: nothing is created for an unregistered harness.
func TestInstall_UnknownConfigHarness_CreatesNothingOnEmptyHome(t *testing.T) {
	home := t.TempDir()
	adapters := []config.AgentAdapter{claude.NewAdapter(), codex.NewAdapter()}
	h := model.Harness{ID: "made-up-config", Type: model.HarnessConfig,
		Agents: []model.Agent{model.AgentClaude, model.AgentCodex}}
	if _, err := config.InstallWith(h, adapters, home, realOpts(home)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := dirDigest(t, home); len(got) != 0 {
		t.Errorf("empty home must stay empty, got %d files", len(got))
	}
}
