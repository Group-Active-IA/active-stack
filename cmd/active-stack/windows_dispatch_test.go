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
