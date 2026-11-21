package hooks

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

// The two hooks and nothing else.
func TestSpecs_OnlyTheTwoReadHooks(t *testing.T) {
	specs := Specs()
	if len(specs) != 2 {
		t.Fatalf("specs = %+v, want session-start and pre-read only", specs)
	}
	got := map[string]string{}
	for _, s := range specs {
		got[s.Subcommand] = s.Event
		if s.TimeoutSeconds != 10 {
			t.Errorf("%s timeout = %d", s.Subcommand, s.TimeoutSeconds)
		}
	}
	if got["session-start"] != "SessionStart" || got["pre-read"] != "PreToolUse" {
		t.Errorf("events = %v", got)
	}
	for _, s := range specs {
		if strings.Contains(s.Subcommand, "save") {
			t.Errorf("a save hook must not be installed: %+v", s)
		}
	}
}

// T-F6-9: each agent gets only its own format; other agents are not touched.
func TestInstall_AgentsAreIndependentAndOthersUntouched(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	opencode := fakeAdapter{model.AgentOpenCode, ".config/opencode/opencode.json"}
	antigravity := fakeAdapter{model.AgentAntigravity, ".gemini/antigravity/settings.json"}
	noFile := fakeAdapter{model.AgentClaude, ""}

	res, err := Install(home, []Adapter{claudeFake, codexFake, opencode, antigravity})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Agents) != 4 {
		t.Fatalf("agents = %+v", res.Agents)
	}
	if exists(opencode.SettingsPath(home)) || exists(antigravity.SettingsPath(home)) {
		t.Error("OpenCode and Antigravity must not be written")
	}
	for _, r := range res.Agents {
		if (r.Agent == model.AgentOpenCode || r.Agent == model.AgentAntigravity) && (!r.Skipped || r.Changed) {
			t.Errorf("%s must be skipped: %+v", r.Agent, r)
		}
	}
	claude := readFile(t, settingsPath(home))
	codex := readFile(t, configPath(home))
	if strings.Contains(claude, EntryMarkerBegin) || strings.Contains(claude, "[[hooks") {
		t.Error("Claude received the Codex block")
	}
	if strings.Contains(codex, "\"hooks\"") || strings.Contains(codex, "\"args\"") {
		t.Error("Codex received the Claude JSON")
	}

	// an adapter without a settings file is a no-op, not an error
	res, err = Install(t.TempDir(), []Adapter{noFile})
	if err != nil || !res.Agents[0].Skipped {
		t.Errorf("no settings path: %+v %v", res, err)
	}
}

func TestRemove_OnlyTouchesClaudeAndCodex(t *testing.T) {
	snaps := stubSnapshots(t)
	home := t.TempDir()
	opencode := fakeAdapter{model.AgentOpenCode, ".config/opencode/opencode.json"}
	writeFile(t, opencode.SettingsPath(home), "{\"a\": 1}\n")
	if _, err := Remove(home, []Adapter{opencode}); err != nil {
		t.Fatal(err)
	}
	if readFile(t, opencode.SettingsPath(home)) != "{\"a\": 1}\n" || len(snaps.calls) != 0 {
		t.Error("OpenCode must not be touched")
	}
}

// If the second agent fails, the first one stays installed and the error says which.
func TestInstall_ErrorNamesTheAgentAndKeepsTheOther(t *testing.T) {
	stubSnapshots(t)
	stubLookPath(t, true)
	home := t.TempDir()
	writeFile(t, configPath(home), "model = \n[[hooks\n")
	res, err := Install(home, []Adapter{claudeFake, codexFake})
	if err == nil || !strings.Contains(err.Error(), "codex") {
		t.Fatalf("err = %v", err)
	}
	if len(res.Agents) < 1 || !res.Agents[0].Changed {
		t.Errorf("the Claude install already done must be reported: %+v", res)
	}
	if readFile(t, configPath(home)) != "model = \n[[hooks\n" {
		t.Error("the broken file changed")
	}
}
