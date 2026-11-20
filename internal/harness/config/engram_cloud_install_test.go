package config_test

import (
	"errors"
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

// [EN] Fase 2 (R-13): the engram-cloud section is injected as its own
// idempotent section (active-stack:engram-cloud) for Claude and Codex only.

const (
	ecOpen  = "<!-- active-stack:engram-cloud -->"
	ecClose = "<!-- /active-stack:engram-cloud -->"
)

func engramCloudHarness() model.Harness {
	return model.Harness{
		ID:        "engram-cloud",
		Type:      model.HarnessConfig,
		Agents:    []model.Agent{model.AgentClaude, model.AgentCodex},
		DependsOn: []string{"engram", "sdd-orchestrator"},
	}
}

// T-F2-1: Claude gets the section in its CLAUDE.md.
func TestInstall_EngramCloud_Claude(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	res, err := config.Install(engramCloudHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0] != claudeMD {
		t.Fatalf("Files = %v, want [%s]", res.Files, claudeMD)
	}
	got := readFile(t, claudeMD)
	if strings.Count(got, ecOpen) != 1 || strings.Count(got, ecClose) != 1 {
		t.Errorf("want exactly one engram-cloud section, got:\n%s", got)
	}
	if strings.Contains(got, "active-stack:sdd-orchestrator") {
		t.Error("engram-cloud install must not compose the orchestrator block")
	}
}

// T-F2-2: Codex gets the section in its AGENTS.md.
func TestInstall_EngramCloud_Codex(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "codex", "AGENTS.md")
	adapter := fakeAdapter{agent: model.AgentCodex, path: path, variant: "codex"}
	res, err := config.Install(engramCloudHarness(), []config.AgentAdapter{adapter}, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0] != path {
		t.Fatalf("Files = %v, want [%s]", res.Files, path)
	}
	if got := readFile(t, path); strings.Count(got, ecOpen) != 1 {
		t.Errorf("codex AGENTS.md lacks exactly one engram-cloud section:\n%s", got)
	}
}

// T-F2-3: OpenCode and Antigravity stay byte for byte intact even though
// their adapters are in the list.
func TestInstall_EngramCloud_OpenCodeAndAntigravityUntouched(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	ocInstr := filepath.Join(home, "oc", "AGENTS.md")
	ocSettings := filepath.Join(home, "oc", "opencode.json")
	agInstr := filepath.Join(home, "ag", "GEMINI.md")
	if err := os.MkdirAll(filepath.Dir(ocInstr), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(agInstr), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := map[string]string{
		ocInstr:    "# opencode rules\n",
		ocSettings: `{"agent":{"keep":{"mode":"primary"}}}`,
		agInstr:    "# antigravity rules\n",
	}
	for p, c := range seed {
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	adapters := []config.AgentAdapter{
		claudeAdapter{path: claudeMD},
		opencodeAdapter{instr: ocInstr, settings: ocSettings},
		fakeAdapter{agent: model.AgentAntigravity, path: agInstr, variant: "antigravity"},
	}
	res, err := config.Install(engramCloudHarness(), adapters, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0] != claudeMD {
		t.Errorf("Files = %v, want only %s", res.Files, claudeMD)
	}
	for p, want := range seed {
		if got := readFile(t, p); got != want {
			t.Errorf("%s changed:\n%s", p, got)
		}
	}
}

// T-F2-4: reinstalling neither duplicates the section nor rewrites the file.
func TestInstall_EngramCloud_Idempotent(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	adapters := []config.AgentAdapter{claudeAdapter{path: claudeMD}}
	if _, err := config.Install(engramCloudHarness(), adapters, home); err != nil {
		t.Fatalf("first Install: %v", err)
	}
	first := readFile(t, claudeMD)
	second, err := config.Install(engramCloudHarness(), adapters, home)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !second.AllAlready {
		t.Error("second install should report AllAlready")
	}
	if got := readFile(t, claudeMD); got != first {
		t.Error("second install changed the file")
	}
	if c := strings.Count(first, ecOpen); c != 1 {
		t.Errorf("section count = %d, want 1", c)
	}
}

// T-F2-5: the orchestrator, design-mockups and user text survive byte for byte.
func TestInstall_EngramCloud_PreservesOrchestratorDesignAndUserContent(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	orch := "<!-- active-stack:sdd-orchestrator -->\norchestrator body\n<!-- /active-stack:sdd-orchestrator -->"
	dm := dmOpen + "\ndesign body\n" + dmClose
	initial := "# My notes\n\nmine before\n\n" + orch + "\n\n" + dm + "\n\nmine after\n"
	if err := os.WriteFile(claudeMD, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Install(engramCloudHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err != nil {
		t.Fatalf("Install: %v", err)
	}
	got := readFile(t, claudeMD)
	for _, keep := range []string{"# My notes\n\nmine before\n\n" + orch + "\n\n" + dm, "mine after"} {
		if !strings.Contains(got, keep) {
			t.Errorf("content %q lost or altered", keep)
		}
	}
	if !strings.Contains(got, ecOpen) {
		t.Error("engram-cloud section missing")
	}
}

// T-F2-6: the snapshot runs before the file is written; a failed snapshot aborts.
func TestInstall_EngramCloud_BacksUpFirst(t *testing.T) {
	t.Run("snapshot runs before write", func(t *testing.T) {
		home := t.TempDir()
		claudeMD := filepath.Join(home, "CLAUDE.md")
		original := "# original\n"
		if err := os.WriteFile(claudeMD, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
		orig := config.SnapshotterCreate
		defer func() { config.SnapshotterCreate = orig }()
		var snapped []string
		var contentAtSnapshot string
		config.SnapshotterCreate = func(_ string, paths []string) error {
			snapped = append(snapped, paths...)
			contentAtSnapshot = readFile(t, claudeMD)
			return nil
		}
		if _, err := config.Install(engramCloudHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err != nil {
			t.Fatalf("Install: %v", err)
		}
		if len(snapped) != 1 || snapped[0] != claudeMD {
			t.Errorf("snapshot paths = %v, want [%s]", snapped, claudeMD)
		}
		if contentAtSnapshot != original {
			t.Error("snapshot must run before the file is modified")
		}
	})
	t.Run("failed snapshot aborts without touching the file", func(t *testing.T) {
		home := t.TempDir()
		claudeMD := filepath.Join(home, "CLAUDE.md")
		original := "# original\n"
		if err := os.WriteFile(claudeMD, []byte(original), 0o644); err != nil {
			t.Fatal(err)
		}
		orig := config.SnapshotterCreate
		defer func() { config.SnapshotterCreate = orig }()
		config.SnapshotterCreate = func(string, []string) error { return errors.New("disk full") }
		if _, err := config.Install(engramCloudHarness(), []config.AgentAdapter{claudeAdapter{path: claudeMD}}, home); err == nil {
			t.Fatal("Install must fail when the backup fails")
		}
		if got := readFile(t, claudeMD); got != original {
			t.Errorf("file modified despite backup failure:\n%s", got)
		}
	})
}

// T-F2-7: neither the orchestrator install nor PurgeStaleSections removes it.
func TestInstall_EngramCloud_NotPurged(t *testing.T) {
	t.Run("orchestrator reinstall keeps it", func(t *testing.T) {
		home := t.TempDir()
		claudeMD := filepath.Join(home, "CLAUDE.md")
		adapters := []config.AgentAdapter{claudeAdapter{path: claudeMD}}
		if _, err := config.Install(engramCloudHarness(), adapters, home); err != nil {
			t.Fatalf("engram-cloud install: %v", err)
		}
		orch := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"delegation"}}
		if _, err := config.Install(orch, adapters, home); err != nil {
			t.Fatalf("orchestrator install: %v", err)
		}
		got := readFile(t, claudeMD)
		if c := strings.Count(got, ecOpen); c != 1 {
			t.Errorf("engram-cloud section count = %d, want 1 after orchestrator install", c)
		}
	})
	t.Run("PurgeStaleSections keeps it", func(t *testing.T) {
		in := "# mine\n\n" + ecOpen + "\nbody\n" + ecClose + "\n"
		out := config.PurgeStaleSections(in)
		if !strings.Contains(out, ecOpen) || !strings.Contains(out, ecClose) {
			t.Errorf("PurgeStaleSections removed engram-cloud:\n%s", out)
		}
	})
}

// T-F2-8: a new version replaces the previous one in place.
func TestInstall_EngramCloud_UpgradeReplacesInPlace(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	orch := "<!-- active-stack:sdd-orchestrator -->\norchestrator body\n<!-- /active-stack:sdd-orchestrator -->"
	old := ecOpen + "\nold content of an earlier version\n" + ecClose
	before := "# My notes\n\nmine before\n\n" + orch + "\n\n" + old + "\n\nmine after\n"
	if err := os.WriteFile(claudeMD, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	adapters := []config.AgentAdapter{claudeAdapter{path: claudeMD}}
	res, err := config.Install(engramCloudHarness(), adapters, home)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.AllAlready {
		t.Error("upgrade must report a change")
	}
	got := readFile(t, claudeMD)
	if strings.Contains(got, "old content of an earlier version") {
		t.Error("old section content must be replaced")
	}
	if strings.Count(got, ecOpen) != 1 || strings.Count(got, ecClose) != 1 {
		t.Errorf("want exactly one section after upgrade:\n%s", got)
	}
	for _, keep := range []string{"# My notes\n\nmine before\n\n" + orch + "\n\n", "\n\nmine after\n"} {
		if !strings.Contains(got, keep) {
			t.Errorf("surrounding content %q not preserved byte for byte", keep)
		}
	}
	if c := strings.Count(got, "<!-- active-stack:"); c != 2 {
		t.Errorf("open marker count = %d, want 2 (orchestrator + engram-cloud)", c)
	}
	again, err := config.Install(engramCloudHarness(), adapters, home)
	if err != nil || !again.AllAlready {
		t.Errorf("second install after upgrade: %+v, err=%v", again, err)
	}
}

// T-F2-12 (install side): Claude and Codex receive the same section body.
func TestInstall_EngramCloud_SameBodyForClaudeAndCodex(t *testing.T) {
	home := t.TempDir()
	claudeMD := filepath.Join(home, "CLAUDE.md")
	codexMD := filepath.Join(home, "codex", "AGENTS.md")
	adapters := []config.AgentAdapter{
		claudeAdapter{path: claudeMD},
		fakeAdapter{agent: model.AgentCodex, path: codexMD, variant: "codex"},
	}
	if _, err := config.Install(engramCloudHarness(), adapters, home); err != nil {
		t.Fatalf("Install: %v", err)
	}
	body := func(s string) string {
		i, j := strings.Index(s, ecOpen), strings.Index(s, ecClose)
		if i < 0 || j < i {
			t.Fatalf("section missing in:\n%s", s)
		}
		return s[i : j+len(ecClose)]
	}
	if body(readFile(t, claudeMD)) != body(readFile(t, codexMD)) {
		t.Error("Claude and Codex must receive the same section text")
	}
}

// T-F2-10 with the REAL Claude and Codex adapters and the real catalog: after
// installing the orchestrator and then engram-cloud, the section sits after the
// orchestrator's universal block, the orchestrator skill and every other file
// are byte for byte identical, and nothing fails.
func TestInstall_EngramCloud_RealAdapters_AfterOrchestratorAndNothingDegrades(t *testing.T) {
	home := t.TempDir()
	adapters := []config.AgentAdapter{claude.NewAdapter(), codex.NewAdapter()}
	installOrchestrator(t, home, adapters)

	claudeMD := claude.NewAdapter().InstructionsPath(home)
	codexMD := codex.NewAdapter().InstructionsPath(home)
	before := dirDigest(t, home)
	beforeText := map[string]string{claudeMD: readFile(t, claudeMD)}
	if _, err := os.Stat(codexMD); err == nil {
		beforeText[codexMD] = readFile(t, codexMD)
	}

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
		t.Fatalf("engram-cloud install must not fail: %v", err)
	}
	if len(res.Files) != 2 {
		t.Errorf("Files = %v, want the Claude and Codex instruction files", res.Files)
	}

	// Order (T-F2-10): after the orchestrator's universal block.
	got := readFile(t, claudeMD)
	universalClose := "<!-- /active-stack:" + config.UniversalSectionID + " -->"
	u, e := strings.Index(got, universalClose), strings.Index(got, ecOpen)
	if u < 0 || e < 0 || e < u {
		t.Errorf("engram-cloud must come after the orchestrator universal block (universal=%d, engram-cloud=%d)", u, e)
	}
	// The orchestrator text is untouched: the new content is only appended.
	for path, prev := range beforeText {
		if !strings.HasPrefix(readFile(t, path), strings.TrimRight(prev, "\n")) {
			t.Errorf("%s: existing content was altered, not just appended to", path)
		}
		if strings.Count(readFile(t, path), ecOpen) != 1 {
			t.Errorf("%s: want exactly one engram-cloud section", path)
		}
	}

	// Nothing else changed: skills and other files identical; only backups added.
	after := dirDigest(t, home)
	for k, v := range before {
		if k == relTo(t, home, claudeMD) || k == relTo(t, home, codexMD) {
			continue
		}
		if after[k] != v {
			t.Errorf("%s changed after installing engram-cloud", k)
		}
	}
	for k := range after {
		if _, existed := before[k]; !existed && !strings.HasPrefix(k, ".active-stack") {
			t.Errorf("%s was created outside the backups dir", k)
		}
	}

	// Reinstalling the orchestrator keeps the section exactly once and in order.
	installOrchestrator(t, home, adapters)
	again := readFile(t, claudeMD)
	if strings.Count(again, ecOpen) != 1 {
		t.Errorf("engram-cloud section count after orchestrator reinstall = %d, want 1", strings.Count(again, ecOpen))
	}
	if u2, e2 := strings.Index(again, universalClose), strings.Index(again, ecOpen); e2 < u2 {
		t.Error("order lost after reinstalling the orchestrator")
	}
}

func relTo(t *testing.T, root, p string) string {
	t.Helper()
	r, err := filepath.Rel(root, p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
