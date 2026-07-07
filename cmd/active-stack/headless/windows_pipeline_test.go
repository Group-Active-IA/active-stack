// Package headless (internal test) — tests for the shared runWindowsPipeline
// helper extracted from RunWindowsInstall (windows-contract-hub-operations,
// Task 2.1 RED). This file is package headless (not headless_test) because
// runWindowsPipeline is unexported.
package headless

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	extinstaller "github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
)

// pipelineFakeCatalog and pipelineFakeRegistry are minimal fakes satisfying
// install.Catalog / install.Registry for the pipeline test, mirroring the
// fakeExecCatalog/fakeExecRegistry pattern used by executor_test.go.
type pipelineFakeCatalog struct {
	harnesses []model.Harness
}

func (f *pipelineFakeCatalog) ByID(id string) (model.Harness, bool) {
	for _, h := range f.harnesses {
		if h.ID == id {
			return h, true
		}
	}
	return model.Harness{}, false
}

func (f *pipelineFakeCatalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

func (f *pipelineFakeCatalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range f.harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

func (f *pipelineFakeCatalog) AllHarnesses() []model.Harness {
	return f.harnesses
}

type pipelineFakeAdapter struct{ agent model.Agent }

func (a pipelineFakeAdapter) Agent() model.Agent               { return a.agent }
func (a pipelineFakeAdapter) InstructionsPath(d string) string { return d + "/CLAUDE.md" }
func (a pipelineFakeAdapter) SkillsDir(d string) string        { return d + "/skills" }
func (a pipelineFakeAdapter) CommandsDir(d string) string      { return d + "/commands" }
func (a pipelineFakeAdapter) SettingsPath(d string) string     { return d + "/settings.json" }
func (a pipelineFakeAdapter) MCPConfigPath(d, s string) string { return d + "/mcp/" + s + ".json" }
func (a pipelineFakeAdapter) MCPStrategy() extinstaller.MCPStrategy {
	return extinstaller.StrategySeparateFile
}
func (a pipelineFakeAdapter) VariantKey() string { return string(a.agent) }
func (a pipelineFakeAdapter) ConfigDelivery() model.ConfigDelivery {
	return model.ConfigDeliveryInstructions
}
func (a pipelineFakeAdapter) PathsFor(base string, _ model.InstallTarget) model.AgentPaths {
	return model.AgentPaths{
		InstructionsPath: base + "/CLAUDE.md",
		SkillsDir:        base + "/skills",
		SettingsPath:     base + "/settings.json",
		CommandsDir:      base + "/commands",
	}.WithMCPConfigFn(func(s string) string { return base + "/mcp/" + s + ".json" })
}

type pipelineFakeRegistry struct {
	adapters map[model.Agent]install.AgentAdapter
}

func (r *pipelineFakeRegistry) Get(agent model.Agent) (install.AgentAdapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// TestRunWindowsPipeline_ParameterizedTerminalEventType asserts that the
// extracted runWindowsPipeline helper emits the install event contract and
// that its terminal event's type is whatever finishedEventType the caller
// passes in (e.g. "starter_finished" for the starters-install path), not
// hardcoded to "install_finished".
func TestRunWindowsPipeline_ParameterizedTerminalEventType(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &pipelineFakeCatalog{harnesses: []model.Harness{h}}
	reg := &pipelineFakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: pipelineFakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFnWithDownload(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
		downloadFn extinstaller.DownloadEventFunc,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	params := ParsedFlags{
		HomeDir: t.TempDir(),
		Yes:     true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := runWindowsPipeline(params, cat, reg, &out, "starter_finished")
	if exitCode != 0 {
		t.Fatalf("runWindowsPipeline() exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected multiple json stream lines, got %d: %q", len(lines), out.String())
	}

	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "starter_finished" || !last.Success {
		t.Fatalf("last event = %+v, want starter_finished success=true", last)
	}

	foundStepStarted := false
	for _, line := range lines {
		var evt struct {
			Type   string `json:"type"`
			StepID string `json:"step_id"`
		}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("stream line is not valid json: %v\nline=%s", err, line)
		}
		if evt.Type == "step_started" && evt.StepID != "" {
			foundStepStarted = true
		}
	}
	if !foundStepStarted {
		t.Fatal("expected at least one step_started event")
	}
}

// TestRunWindowsPipeline_DefaultInstallFinished asserts the install path's
// default terminal event type still works as "install_finished" through the
// shared helper (regression guard alongside the existing windows_test.go
// RunWindowsInstall assertions).
func TestRunWindowsPipeline_DefaultInstallFinished(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &pipelineFakeCatalog{harnesses: []model.Harness{h}}
	reg := &pipelineFakeRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: pipelineFakeAdapter{agent: model.AgentClaude},
	}}

	restoreSnap := install.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{}, nil
	})
	defer restoreSnap()

	restoreExt := install.SetExternalInstallFnWithDownload(func(
		_ context.Context,
		_ model.Harness,
		_ system.PlatformProfile,
		_ []extinstaller.AgentAdapter,
		_ string,
		downloadFn extinstaller.DownloadEventFunc,
	) (extinstaller.Result, error) {
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	params := ParsedFlags{
		HomeDir: t.TempDir(),
		Yes:     true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := runWindowsPipeline(params, cat, reg, &out, "install_finished")
	if exitCode != 0 {
		t.Fatalf("runWindowsPipeline() exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "install_finished" || !last.Success {
		t.Fatalf("last event = %+v, want install_finished success=true", last)
	}
}
