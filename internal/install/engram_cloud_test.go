package install_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// [EN] Fase 9 (R-20): engram-cloud en install.

func selectedIDs(t *testing.T, intent install.Intent) []string {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	hs, err := install.SelectHarnesses(cat, intent)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, h := range hs {
		ids = append(ids, h.ID)
	}
	return ids
}

func has(ids []string, id string) bool { return indexOf(ids, id) >= 0 }

// T-F9-1
func TestFullIncludesEngramCloudLiteDoesNot(t *testing.T) {
	agents := []model.Agent{model.AgentClaude, model.AgentCodex}
	if !has(selectedIDs(t, install.Intent{Agents: agents, Mode: model.ModeFull}), "engram-cloud") {
		t.Error("Full must include engram-cloud")
	}
	if has(selectedIDs(t, install.Intent{Agents: agents, Mode: model.ModeLite}), "engram-cloud") {
		t.Error("Lite must not include engram-cloud")
	}
}

// T-F9-2 and T-F9-3
func TestCustomCloudPullsDependenciesInOrderAndEngramAloneDoesNot(t *testing.T) {
	cat, _ := catalog.Load()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentClaude: fakeAdapter{agent: model.AgentClaude}}}
	plan, err := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeCustom, Custom: []string{"engram-cloud"}}, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatal(err)
	}
	ids := applyStepIDs(plan.StagePlan)
	e, o, c := indexOf(ids, "engram"), indexOf(ids, "sdd-orchestrator"), indexOf(ids, "engram-cloud")
	if e < 0 || o < 0 || c < 0 || !(e < c && o < c) {
		t.Fatalf("order %v", ids)
	}
	plan, err = install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeCustom, Custom: []string{"engram"}}, buildOptions(t.TempDir(), reg, nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range plan.Apply {
		if strings.Contains(s.ID(), "engram-cloud") {
			t.Fatalf("engram alone added %s", s.ID())
		}
	}
}

func readOr(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

// T-F9-4, T-F9-9 and the hooks wiring: claude+opencode installs in Claude,
// writes nothing cloud-related in OpenCode and does not fail.
func TestCloudInstallClaudeAndOpenCode(t *testing.T) {
	stubNetwork(t)
	home := t.TempDir()
	reg := newRealRegistry(t)
	seedHome(t, home, reg)
	oc, _ := reg.Get(model.AgentOpenCode)
	before := map[string]string{}
	for _, p := range []string{oc.SettingsPath(home), oc.InstructionsPath(home)} {
		before[p] = readOr(t, p)
	}
	res := runRealInstall(t, home, install.Intent{
		Agents: []model.Agent{model.AgentClaude, model.AgentOpenCode},
		Mode:   model.ModeCustom, Custom: []string{"engram-cloud"}, Tier: model.TierBalanceado,
	})
	if res.Err != nil {
		t.Fatalf("install failed: %v", res.Err)
	}
	cl, _ := reg.Get(model.AgentClaude)
	settings := readOr(t, cl.SettingsPath(home))
	if !strings.Contains(settings, "memory") || !strings.Contains(settings, "session-start") || !strings.Contains(settings, "pre-read") {
		t.Errorf("hooks missing in Claude settings:\n%s", settings)
	}
	if !strings.Contains(readOr(t, cl.InstructionsPath(home)), "<!-- active-stack:engram-cloud -->") {
		t.Error("engram-cloud section missing in CLAUDE.md")
	}
	mcp := readOr(t, cl.MCPConfigPath(home, "engram"))
	if strings.Count(mcp, "ENGRAM_CLOUD_AUTOSYNC") != 1 {
		t.Errorf("AUTOSYNC missing in the engram MCP env:\n%s", mcp)
	}
	// DA-13: the save-tagging proxy wraps the existing Engram entry.
	if !strings.Contains(mcp, "proxy") {
		t.Errorf("save-tagging proxy missing in the engram MCP entry: %s", mcp)
	}
	for _, bad := range []string{"ENGRAM_CLOUD_TOKEN", "ENGRAM_CLOUD_SERVER", "ENGRAM_CLOUD_URL"} {
		if strings.Contains(mcp+settings, bad) {
			t.Errorf("%s must never be written", bad)
		}
	}
	for p, b := range before {
		if got := readOr(t, p); strings.Contains(got, "engram-cloud") || strings.Contains(got, "AUTOSYNC") || (strings.Contains(b, "theme") && !strings.Contains(got, "theme")) {
			t.Errorf("OpenCode file %s touched by the cloud harness:\n%s", p, got)
		}
	}
	cx, _ := reg.Get(model.AgentCodex)
	if c := readOr(t, cx.SettingsPath(home)); strings.Contains(c, "memory-hooks") || strings.Contains(c, "AUTOSYNC") {
		t.Errorf("Codex was not selected but got cloud content:\n%s", c)
	}
}

func sameJSON(t *testing.T, a, b string) bool {
	t.Helper()
	var x, y any
	if json.Unmarshal([]byte(a), &x) != nil || json.Unmarshal([]byte(b), &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}

// T-F9-8: reinstalling without the harness over an install that had it keeps
// hooks, AUTOSYNC and the section (nothing else is touched).
func TestReinstallWithoutCloudTouchesNothingOfCloud(t *testing.T) {
	stubNetwork(t)
	home := t.TempDir()
	reg := newRealRegistry(t)
	seedHome(t, home, reg)
	agents := []model.Agent{model.AgentClaude}
	if r := runRealInstall(t, home, install.Intent{Agents: agents, Mode: model.ModeCustom, Custom: []string{"engram-cloud"}, Tier: model.TierBalanceado}); r.Err != nil {
		t.Fatal(r.Err)
	}
	cl, _ := reg.Get(model.AgentClaude)
	settingsBefore := readOr(t, cl.SettingsPath(home))
	mcpBefore := readOr(t, cl.MCPConfigPath(home, "engram"))
	if r := runRealInstall(t, home, install.Intent{Agents: agents, Mode: model.ModeLite, Tier: model.TierBalanceado}); r.Err != nil {
		t.Fatal(r.Err)
	}
	// The permissions installer (shared, outside R-20) re-serializes
	// settings.json with sorted keys, so the comparison is semantic.
	if !sameJSON(t, readOr(t, cl.SettingsPath(home)), settingsBefore) {
		t.Error("settings.json changed by a reinstall without the harness")
	}
	if readOr(t, cl.MCPConfigPath(home, "engram")) != mcpBefore {
		t.Error("MCP file changed by a reinstall without the harness")
	}
}

// Reinstalling with the harness is idempotent.
func TestCloudInstallIsIdempotent(t *testing.T) {
	stubNetwork(t)
	home := t.TempDir()
	reg := newRealRegistry(t)
	seedHome(t, home, reg)
	intent := install.Intent{Agents: []model.Agent{model.AgentClaude, model.AgentCodex}, Mode: model.ModeCustom, Custom: []string{"engram-cloud"}, Tier: model.TierBalanceado}
	if r := runRealInstall(t, home, intent); r.Err != nil {
		t.Fatal(r.Err)
	}
	cl, _ := reg.Get(model.AgentClaude)
	cx, _ := reg.Get(model.AgentCodex)
	claudeFirst := readOr(t, cl.SettingsPath(home))
	if r := runRealInstall(t, home, intent); r.Err != nil {
		t.Fatal(r.Err)
	}
	if !sameJSON(t, readOr(t, cl.SettingsPath(home)), claudeFirst) {
		t.Error("claude settings.json changed semantically on reinstall")
	}
	if c := readOr(t, cl.SettingsPath(home)); strings.Count(c, "session-start") != 1 || strings.Count(c, "pre-read") != 1 {
		t.Errorf("claude hooks duplicated:\n%s", c)
	}
	if c := readOr(t, cl.MCPConfigPath(home, "engram")); strings.Count(c, "ENGRAM_CLOUD_AUTOSYNC") != 1 {
		t.Errorf("claude AUTOSYNC count:\n%s", c)
	}
	if c := readOr(t, cx.SettingsPath(home)); strings.Count(c, "memory-hooks begin") != 1 || strings.Count(c, "ENGRAM_CLOUD_AUTOSYNC") != 1 {
		t.Errorf("codex config.toml:\n%s", c)
	}
	if c := readOr(t, cl.InstructionsPath(home)); strings.Count(c, "<!-- active-stack:engram-cloud -->") != 1 {
		t.Error("engram-cloud section duplicated")
	}
}

// Project target never writes user-level hooks or AUTOSYNC.
func TestCloudStepsSkippedForProjectTarget(t *testing.T) {
	cat, _ := catalog.Load()
	reg := &fakeRegistry{adapters: map[model.Agent]install.AgentAdapter{model.AgentClaude: fakeAdapter{agent: model.AgentClaude}}}
	opts := buildOptions(t.TempDir(), reg, nil)
	opts.Target = model.Project
	opts.ProjectRoot = t.TempDir()
	plan, err := install.BuildPlan(cat, install.Intent{Agents: []model.Agent{model.AgentClaude}, Mode: model.ModeCustom, Custom: []string{"engram-cloud"}}, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range plan.Apply {
		if strings.HasPrefix(s.ID(), "hooks:") || strings.HasPrefix(s.ID(), "autosync:") {
			t.Fatalf("project target got %s", s.ID())
		}
	}
}
