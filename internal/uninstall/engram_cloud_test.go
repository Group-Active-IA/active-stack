package uninstall_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

// [EN] Fase 9 (R-20) T-F9-6 y T-F9-7.

type realReg struct {
	m map[model.Agent]uninstall.AgentAdapter
}

func (r realReg) Get(a model.Agent) (uninstall.AgentAdapter, bool) { v, ok := r.m[a]; return v, ok }

func w(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func r(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const (
	cOpen   = "<!-- active-stack:engram-cloud -->\ncloud block\n<!-- /active-stack:engram-cloud -->\n"
	foreign = `{
  "theme": "dark",
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "^mcp__plugin_engram_engram__mem_search$",
        "hooks": [ { "type": "command", "command": "engram", "args": ["hook", "claude-pre-tool-use"] } ]
      }
    ]
  }
}
`
)

func TestUninstallCloudRemovesOwnHooksAndKeepsEverythingElse(t *testing.T) {
	home := t.TempDir()
	defer uninstall.SetSnapshotCreate(func(string, []string) (backup.Manifest, error) { return backup.Manifest{}, nil })()
	ca, cx := claude.NewAdapter(), codex.NewAdapter()
	reg := realReg{m: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: ca, model.AgentCodex: cx}}

	instr := ca.InstructionsPath(home)
	w(t, instr, "# Mine\n\n"+dmOrchSection+"\n"+dmSection+"\n"+cOpen+"\nfooter\n")
	w(t, ca.SettingsPath(home), foreign)
	mcp := ca.MCPConfigPath(home, "engram")
	const mcpSrc = "{\n  \"mcpServers\": {\n    \"engram\": {\n      \"command\": \"engram\",\n      \"args\": [\n        \"mcp\"\n      ]\n    }\n  }\n}\n"
	w(t, mcp, mcpSrc)
	codexCfg := cx.SettingsPath(home)
	const codexSrc = "model = \"gpt-5\"\n\n[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]\n"
	w(t, codexCfg, codexSrc)
	cxInstr := cx.InstructionsPath(home)
	w(t, cxInstr, "# codex mine\n\n"+cOpen)

	// What a project and the person own: never deleted.
	project := t.TempDir()
	keep := map[string]string{
		filepath.Join(project, ".engram", "config.json"):       `{"project_name":"x"}`,
		filepath.Join(project, ".engram", "active-stack.json"): `{"mode":"cloud-branch"}`,
		filepath.Join(home, ".engram", "engram.db"):            "memories",
	}
	for p, c := range keep {
		w(t, p, c)
	}

	// Install what engram-cloud installs.
	if _, err := hooks.Install(home, []hooks.Adapter{ca, cx}); err != nil {
		t.Fatal(err)
	}
	if _, err := autosync.Set(home, []autosync.Adapter{ca, cx}); err != nil {
		t.Fatal(err)
	}

	plan, err := uninstall.BuildPlan(
		&fakeCatalog{harnesses: []model.Harness{{ID: "engram-cloud", Type: model.HarnessConfig, InstallModes: []model.InstallMode{model.ModeFull}, Agents: []model.Agent{model.AgentClaude, model.AgentCodex}}}},
		uninstall.Intent{Agents: []model.Agent{model.AgentClaude, model.AgentCodex}, Mode: model.ModeFull, Strategy: uninstall.StrategyTargeted},
		uninstall.Options{HomeDir: home, Registry: reg},
	)
	if err != nil {
		t.Fatal(err)
	}
	if res := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(plan.StagePlan); res.Err != nil {
		t.Fatalf("uninstall: %v", res.Err)
	}

	got := r(t, instr)
	if strings.Contains(got, "engram-cloud") {
		t.Errorf("cloud section not removed:\n%s", got)
	}
	for _, k := range []string{"# Mine", "footer", "<!-- active-stack:sdd-orchestrator -->", "<!-- active-stack:design-mockups -->"} {
		if !strings.Contains(got, k) {
			t.Errorf("%q lost:\n%s", k, got)
		}
	}
	if s := r(t, ca.SettingsPath(home)); s != foreign {
		t.Errorf("settings.json is not back to the foreign content:\n%s", s)
	}
	if r(t, mcp) != mcpSrc {
		t.Error("claude MCP file not restored")
	}
	if c := r(t, codexCfg); strings.Contains(c, "memory-hooks") || strings.Contains(c, "AUTOSYNC") || !strings.Contains(c, `model = "gpt-5"`) || !strings.Contains(c, `command = "engram"`) {
		t.Errorf("codex config.toml:\n%s", c)
	}
	// T-F9-7
	for p, c := range keep {
		if r(t, p) != c {
			t.Errorf("%s was changed or deleted", p)
		}
	}
}
