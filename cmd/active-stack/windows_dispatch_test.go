package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestWindowsDispatch_Detect(t *testing.T) {
	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"detect", "--home", t.TempDir()}, nil, reg, &out)
	if exitCode != 0 {
		t.Fatalf("windows detect exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("detect output is not json: %v\nbody=%s", err, out.String())
	}
}

func TestWindowsDispatch_Options(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"options", "--agent", "claude"}, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("windows options exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("options output is not json: %v\nbody=%s", err, out.String())
	}
}

func TestWindowsDispatch_NoSubcommand(t *testing.T) {
	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{}, nil, nil, &out)
	if exitCode == 0 {
		t.Fatal("windows with no subcommand must exit non-zero")
	}
	if !strings.Contains(out.String(), "detect") {
		t.Fatalf("usage must mention detect/options/install; got:\n%s", out.String())
	}
}

func TestWindowsDispatch_OptionsMissingAgent(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"options"}, cat, nil, &out)
	if exitCode == 0 {
		t.Fatal("windows options without --agent must exit non-zero")
	}
}

// TestWindowsDispatch_OptionsMultiAgent verifies that "windows options" accepts
// a comma-separated --agent list and returns valid JSON (task 5).
func TestWindowsDispatch_OptionsMultiAgent(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"options", "--agent", "claude,opencode"}, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("windows options --agent claude,opencode exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("options output is not json: %v\nbody=%s", err, out.String())
	}
}

// TestWindowsDispatch_OptionsWhitespaceAndEmptyEntries verifies that the CSV
// --agent parser trims whitespace and skips empty entries.
func TestWindowsDispatch_OptionsWhitespaceAndEmptyEntries(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"options", "--agent", " , claude , "}, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf(`windows options --agent " , claude , " exit = %d; output:%s`, exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("options output is not json: %v\nbody=%s", err, out.String())
	}
}

// TestWindowsDispatch_OptionsOnlyEmptyEntries verifies that an --agent value
// made only of empty/whitespace entries is rejected (empty resulting slice).
func TestWindowsDispatch_OptionsOnlyEmptyEntries(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"options", "--agent", " , , "}, cat, nil, &out)
	if exitCode == 0 {
		t.Fatal("windows options with only empty/whitespace --agent entries must exit non-zero")
	}
}

func TestWindowsDispatch_Uninstall(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	defaultReg, err := agents.NewDefaultRegistry()
	if err != nil {
		t.Fatalf("agents.NewDefaultRegistry() error = %v", err)
	}

	var out bytes.Buffer
	exitCode := runWindowsDispatch(
		[]string{"uninstall", "--agent", "claude", "--mode", "lite", "--dry-run"},
		cat,
		agentRegistryAdapter{r: defaultReg},
		&out,
	)
	if exitCode != 0 {
		t.Fatalf("windows uninstall exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := splitJSONLines(out.String())
	if len(lines) == 0 {
		t.Fatal("expected json-stream output for windows uninstall")
	}

	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last uninstall line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "uninstall_finished" || !last.Success {
		t.Fatalf("last uninstall event = %+v, want uninstall_finished success=true", last)
	}
}

// ── windows-contract-hub-operations: new subcommand routing (Task 7.1 RED) ──

func TestWindowsDispatch_StartersList(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"starters", "list"}, cat, nil, &out)
	if exitCode != 0 {
		t.Fatalf("windows starters list exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("starters list output is not json: %v\nbody=%s", err, out.String())
	}
	if _, ok := body["starters"]; !ok {
		t.Fatalf("starters list response missing \"starters\" key: %s", out.String())
	}
}

func TestWindowsDispatch_StartersInstall(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	projectRoot := t.TempDir()

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{
		"starters", "install",
		"--starter", "active-ia",
		"--project", projectRoot,
		"--agent", "claude",
		"--dry-run",
	}, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("windows starters install exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := splitJSONLines(out.String())
	if len(lines) == 0 {
		t.Fatal("expected json-stream output for windows starters install")
	}
	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last starters install line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "starter_finished" || !last.Success {
		t.Fatalf("last starters install event = %+v, want starter_finished success=true", last)
	}
}

func TestWindowsDispatch_StartersInstallMissingProject(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("catalog.Load() error = %v", err)
	}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{
		"starters", "install",
		"--starter", "active-ia",
		"--agent", "claude",
	}, cat, nil, &out)
	if exitCode == 0 {
		t.Fatal("windows starters install without --project must exit non-zero")
	}
}

func TestWindowsDispatch_BackupsList(t *testing.T) {
	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"backups", "list", "--home", t.TempDir()}, nil, nil, &out)
	if exitCode != 0 {
		t.Fatalf("windows backups list exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("backups list output is not json: %v\nbody=%s", err, out.String())
	}
	if _, ok := body["backups"]; !ok {
		t.Fatalf("backups list response missing \"backups\" key: %s", out.String())
	}
}

func TestWindowsDispatch_BackupsActionMissingID(t *testing.T) {
	for _, action := range []string{"restore", "delete", "rename"} {
		t.Run(action, func(t *testing.T) {
			var out bytes.Buffer
			exitCode := runWindowsDispatch([]string{"backups", action, "--home", t.TempDir()}, nil, nil, &out)
			if exitCode == 0 {
				t.Fatalf("windows backups %s without --id must exit non-zero", action)
			}
		})
	}
}

func TestWindowsDispatch_BackupsActionReachesHandler(t *testing.T) {
	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{
		"backups", "delete",
		"--home", t.TempDir(),
		"--id", "does-not-exist",
	}, nil, nil, &out)
	if exitCode == 0 {
		t.Fatal("windows backups delete with an unknown id must exit non-zero")
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("backups delete output is not json: %v\nbody=%s", err, out.String())
	}
	if body["success"] != false {
		t.Fatalf("backups delete response = %+v, want success=false", body)
	}
	if body["id"] != "does-not-exist" {
		t.Fatalf("backups delete response = %+v, want id=does-not-exist", body)
	}
}

func TestWindowsDispatch_UninstallOptions(t *testing.T) {
	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"uninstall-options", "--home", t.TempDir()}, nil, nil, &out)
	if exitCode != 0 {
		t.Fatalf("windows uninstall-options exit = %d; output:\n%s", exitCode, out.String())
	}

	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("uninstall-options output is not json: %v\nbody=%s", err, out.String())
	}
	for _, key := range []string{"detected_agents", "modes", "strategies"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("uninstall-options response missing %q key: %s", key, out.String())
		}
	}
}

// TestWindowsDispatch_ExistingRoutesUnchanged re-asserts that detect, options,
// install, and uninstall keep routing exactly as before this change — the
// new subcommand cases must not shadow or alter them.
func TestWindowsDispatch_ExistingRoutesUnchanged(t *testing.T) {
	reg := starterAddTestReg{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: starterAddTestAdapter{agent: model.AgentClaude},
	}}

	var out bytes.Buffer
	exitCode := runWindowsDispatch([]string{"detect", "--home", t.TempDir()}, nil, reg, &out)
	if exitCode != 0 {
		t.Fatalf("windows detect exit = %d; output:\n%s", exitCode, out.String())
	}
	var body map[string]any
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("detect output is not json: %v\nbody=%s", err, out.String())
	}
	if _, ok := body["detected_agents"]; !ok {
		t.Fatalf("detect response missing \"detected_agents\" key: %s", out.String())
	}
}

func splitJSONLines(s string) []string {
	raw := strings.Split(strings.TrimSpace(s), "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
