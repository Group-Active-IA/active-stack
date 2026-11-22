package verify_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/autosync"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/harness/proxyreg"
	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

const (
	cloudMarker  = "<!-- active-stack:engram-cloud -->"
	claudeMCPRaw = "{\n  \"mcpServers\": {\n    \"engram\": {\"command\": \"engram\", \"args\": [\"mcp\"]}\n  }\n}\n"
	codexMCPRaw  = "[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]\n"
)

var cloudHarness = model.Harness{
	ID: "engram-cloud", Type: model.HarnessConfig,
	Agents: []model.Agent{model.AgentClaude, model.AgentCodex},
}

type cloudEnv struct {
	home    string
	adapter fakeAdapter
}

// newCloudEnv builds a fully installed engram-cloud for one agent, using the
// real installers of R-16/R-26/R-20 (backups stubbed out), and a PATH stub
// that finds active-stack.
func newCloudEnv(t *testing.T, agent model.Agent) cloudEnv {
	t.Helper()
	t.Cleanup(hooks.SetSnapshotterCreate(func(string, []string) error { return nil }))
	t.Cleanup(proxyreg.SetSnapshotterCreate(func(string, []string) error { return nil }))
	t.Cleanup(autosync.SetSnapshotterCreate(func(string, []string) error { return nil }))
	t.Cleanup(hooks.SetLookPath(func(string) (string, error) { return "active-stack", nil }))

	home := t.TempDir()
	a := fakeAdapter{
		agent:            agent,
		instructionsPath: filepath.Join(home, "instructions.md"),
	}
	mcp := claudeMCPRaw
	if agent == model.AgentCodex {
		a.settingsPath = filepath.Join(home, "config.toml")
		a.mcpConfigPath = a.settingsPath
		mcp = codexMCPRaw
	} else {
		a.settingsPath = filepath.Join(home, "settings.json")
		a.mcpConfigPath = filepath.Join(home, "mcp.json")
	}
	if err := os.WriteFile(a.instructionsPath, []byte("x\n"+cloudMarker+"\nbody\n<!-- /active-stack:engram-cloud -->\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(a.mcpConfigPath, []byte(mcp), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := hooks.Install(home, []hooks.Adapter{a}); err != nil {
		t.Fatal(err)
	}
	if _, err := proxyreg.Register(home, []proxyreg.Adapter{a}); err != nil {
		t.Fatal(err)
	}
	if _, err := autosync.Set(home, []autosync.Adapter{a}); err != nil {
		t.Fatal(err)
	}
	return cloudEnv{home: home, adapter: a}
}

func (e cloudEnv) run(t *testing.T, adapters ...verify.Adapter) map[string]verify.CheckResult {
	t.Helper()
	return e.runOpts(t, nil, adapters...)
}

func (e cloudEnv) runOpts(t *testing.T, opts []verify.Option, adapters ...verify.Adapter) map[string]verify.CheckResult {
	t.Helper()
	if len(adapters) == 0 {
		adapters = []verify.Adapter{e.adapter}
	}
	out := map[string]verify.CheckResult{}
	for _, r := range verify.RunChecks(context.Background(), verify.ChecksForHarness(cloudHarness, adapters, e.home, opts...)) {
		out[r.ID] = r
	}
	return out
}

func wantStatus(t *testing.T, got map[string]verify.CheckResult, id string, want verify.CheckStatus) verify.CheckResult {
	t.Helper()
	r, ok := got[id]
	if !ok {
		t.Fatalf("check %q missing; have %v", id, keys(got))
	}
	if r.Status != want {
		t.Errorf("check %q: status %q (%s), want %q", id, r.Status, r.Error, want)
	}
	return r
}

func keys(m map[string]verify.CheckResult) []string {
	var k []string
	for id := range m {
		k = append(k, id)
	}
	return k
}

func TestCloudAllGoodClaude(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	got := e.run(t)
	for _, r := range got {
		if r.Status != verify.CheckStatusPassed {
			t.Errorf("%s: %s %s", r.ID, r.Status, r.Error)
		}
	}
	for _, id := range []string{"config:engram-cloud:claude", "cloud:hooks:claude", "cloud:path:claude", "cloud:proxy:claude", "cloud:autosync:claude"} {
		wantStatus(t, got, id, verify.CheckStatusPassed)
	}
	if _, ok := got["cloud:trust:claude"]; ok {
		t.Error("trust reminder is for Codex only")
	}
}

// T-F10-1: the marker exactly once.
func TestCloudMarkerExactlyOnce(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	doubled := cloudMarker + "\n" + cloudMarker + "\n"
	if err := os.WriteFile(e.adapter.instructionsPath, []byte(doubled), 0o600); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, e.run(t), "config:engram-cloud:claude", verify.CheckStatusFailed)
	if err := os.WriteFile(e.adapter.instructionsPath, []byte("nothing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, e.run(t), "config:engram-cloud:claude", verify.CheckStatusFailed)
}

// T-F10-3: hooks installed (hard) and active-stack on PATH (hard).
func TestCloudHooksMissingAndPath(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if _, err := hooks.Remove(e.home, []hooks.Adapter{e.adapter}); err != nil {
		t.Fatal(err)
	}
	r := wantStatus(t, e.run(t), "cloud:hooks:claude", verify.CheckStatusFailed)
	if !strings.Contains(r.Error, "session-start") {
		t.Errorf("message does not name the missing hook: %s", r.Error)
	}

	e2 := newCloudEnv(t, model.AgentClaude)
	restore := hooks.SetLookPath(func(string) (string, error) { return "", errors.New("not found") })
	defer restore()
	got := e2.run(t)
	wantStatus(t, got, "cloud:hooks:claude", verify.CheckStatusPassed)
	wantStatus(t, got, "cloud:path:claude", verify.CheckStatusWarning)
}

// The proxy: not registered is a hard failure, no Engram entry too.
func TestCloudProxy(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if _, err := proxyreg.Unregister(e.home, []proxyreg.Adapter{e.adapter}); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, e.run(t), "cloud:proxy:claude", verify.CheckStatusFailed)
	if err := os.WriteFile(e.adapter.mcpConfigPath, []byte(`{"mcpServers":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, e.run(t), "cloud:proxy:claude", verify.CheckStatusFailed)
}

// T-F10-7: soft.
func TestCloudAutosyncSoft(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if _, err := autosync.Remove(e.home, []autosync.Adapter{e.adapter}); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, e.run(t), "cloud:autosync:claude", verify.CheckStatusWarning)
}

// T-F10-4: Codex always gets the trust reminder, as a warning.
func TestCloudCodex(t *testing.T) {
	e := newCloudEnv(t, model.AgentCodex)
	got := e.run(t)
	r := wantStatus(t, got, "cloud:trust:codex", verify.CheckStatusWarning)
	if !strings.Contains(r.Error, "/hooks") {
		t.Errorf("reminder: %s", r.Error)
	}
	for _, id := range []string{"cloud:hooks:codex", "cloud:path:codex", "cloud:proxy:codex", "cloud:autosync:codex", "config:engram-cloud:codex"} {
		wantStatus(t, got, id, verify.CheckStatusPassed)
	}
	if !verify.BuildReport(verify.RunChecks(context.Background(), nil)).Ready {
		t.Error("sanity")
	}
	var results []verify.CheckResult
	for _, r := range got {
		results = append(results, r)
	}
	if !verify.BuildReport(results).Ready {
		t.Error("the Codex warning must not block the install")
	}
}

// T-F10-2 and T-F10-5: OpenCode and Antigravity are never checked and never
// fail; with the mode in use they only warn.
func TestCloudOutOfScopeAgents(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	oc := fakeAdapter{agent: model.AgentOpenCode, instructionsPath: filepath.Join(e.home, "none.md")}
	ag := fakeAdapter{agent: model.AgentAntigravity, instructionsPath: filepath.Join(e.home, "none2.md")}
	got := e.run(t, e.adapter, oc, ag)
	for id := range got {
		if strings.Contains(id, "opencode") || strings.Contains(id, "antigravity") {
			t.Errorf("unexpected check %q without the project mode", id)
		}
	}

	if err := os.MkdirAll(filepath.Join(e.home, ".engram"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(e.home, ".engram", "config.json"), `{"project_name":"demo"}`)
	write(t, filepath.Join(e.home, ".engram", "active-stack.json"), `{"memory_policy":"cloud-branch"}`)
	got = e.run(t, e.adapter, oc, ag)
	wantStatus(t, got, "cloud:unsupported:opencode", verify.CheckStatusWarning)
	wantStatus(t, got, "cloud:unsupported:antigravity", verify.CheckStatusWarning)
	wantStatus(t, got, "cloud:project-files", verify.CheckStatusPassed)
	var results []verify.CheckResult
	for _, r := range got {
		results = append(results, r)
	}
	if !verify.BuildReport(results).Ready {
		t.Error("out-of-scope agents must never fail verify")
	}
}

func write(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// T-F10-6 and T-F10-8: project files are validated and no value is printed.
func TestCloudProjectFilesAndNoSecrets(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if err := os.MkdirAll(filepath.Join(e.home, ".engram"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(e.home, ".engram", "config.json")
	mk := filepath.Join(e.home, ".engram", "active-stack.json")
	const secret = "ENGRAM_CLOUD_TOKEN-s3cr3t-value"

	// BOM, valid.
	write(t, cfg, "\xEF\xBB\xBF"+`{"project_name":"demo","token":"`+secret+`"}`)
	write(t, mk, `{"memory_policy":"cloud-branch"}`)
	wantStatus(t, e.run(t), "cloud:project-files", verify.CheckStatusPassed)

	for name, c := range map[string][2]string{
		"bad json":     {`{"project_name":`, `{"memory_policy":"cloud-branch"}`},
		"no name":      {`{"token":"` + secret + `"}`, `{"memory_policy":"cloud-branch"}`},
		"wrong policy": {`{"project_name":"demo","token":"` + secret + `"}`, `{"memory_policy":"` + secret + `"}`},
	} {
		write(t, cfg, c[0])
		write(t, mk, c[1])
		r := wantStatus(t, e.run(t), "cloud:project-files", verify.CheckStatusFailed)
		if strings.Contains(r.Error, secret) {
			t.Errorf("%s: the message leaks a file value: %s", name, r.Error)
		}
	}
}

// Health checks never modify anything.
func TestCloudChecksAreReadOnly(t *testing.T) {
	e := newCloudEnv(t, model.AgentCodex)
	snap := func() map[string]string {
		m := map[string]string{}
		_ = filepath.Walk(e.home, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				b, _ := os.ReadFile(p)
				m[p] = string(b) + info.ModTime().String()
			}
			return nil
		})
		return m
	}
	before := snap()
	e.run(t)
	after := snap()
	if len(before) != len(after) {
		t.Fatalf("files created or removed: %d -> %d", len(before), len(after))
	}
	for p, v := range before {
		if after[p] != v {
			t.Errorf("%s changed", p)
		}
	}
}

func TestCloudMessagesSpanish(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if _, err := hooks.Remove(e.home, []hooks.Adapter{e.adapter}); err != nil {
		t.Fatal(err)
	}
	r := wantStatus(t, e.runOpts(t, []verify.Option{verify.WithLang(i18n.LangES)}), "cloud:hooks:claude", verify.CheckStatusFailed)
	if !strings.Contains(r.Error, "faltan los hooks") || !strings.Contains(r.Description, "instalados") {
		t.Errorf("not Spanish: %q / %q", r.Error, r.Description)
	}
	r = wantStatus(t, e.run(t), "cloud:hooks:claude", verify.CheckStatusFailed)
	if !strings.Contains(r.Error, "memory hooks missing") {
		t.Errorf("not English: %q", r.Error)
	}
}

// Windows: the PATH lookup finds active-stack.exe through PATHEXT, and paths
// with spaces and backslashes work. Runs the real exec.LookPath.
func TestCloudPathLookupWindows(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	restore := hooks.SetLookPath(exec.LookPath)
	defer restore()

	bin := filepath.Join(t.TempDir(), "with space")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "active-stack"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	write(t, filepath.Join(bin, name), "stub")
	if err := os.Chmod(filepath.Join(bin, name), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	wantStatus(t, e.run(t), "cloud:path:claude", verify.CheckStatusPassed)
	t.Setenv("PATH", t.TempDir())
	wantStatus(t, e.run(t), "cloud:path:claude", verify.CheckStatusWarning)
}

// Hooks, proxy and AUTOSYNC are user-level (D12): a Project install has none,
// so verify must not demand them there (it would roll the install back).
func TestCloudProjectTargetSkipsUserLevelChecks(t *testing.T) {
	e := newCloudEnv(t, model.AgentClaude)
	if _, err := hooks.Remove(e.home, []hooks.Adapter{e.adapter}); err != nil {
		t.Fatal(err)
	}
	got := e.runOpts(t, []verify.Option{verify.WithTarget(model.Project)})
	wantStatus(t, got, "config:engram-cloud:claude", verify.CheckStatusPassed)
	for id := range got {
		if strings.HasPrefix(id, "cloud:hooks") || strings.HasPrefix(id, "cloud:path") || strings.HasPrefix(id, "cloud:proxy") || strings.HasPrefix(id, "cloud:autosync") {
			t.Errorf("Project target must not build %q", id)
		}
	}
}
