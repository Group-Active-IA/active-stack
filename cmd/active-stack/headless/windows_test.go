package headless_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/backup"
	extinstaller "github.com/Group-Active-IA/active-stack/internal/harness/external"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/system"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

func TestRunWindowsDetect_JSON(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.MkdirAll(homeDir+"/.claude", 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(homeDir+"/.codex", 0o755); err != nil {
		t.Fatalf("mkdir .codex: %v", err)
	}

	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
		model.AgentCodex:  fakeExecAdapter{agent: model.AgentCodex},
	}}

	var out bytes.Buffer
	if err := headless.RunWindowsDetect(homeDir, reg, &out); err != nil {
		t.Fatalf("RunWindowsDetect() error = %v", err)
	}

	var resp struct {
		DetectedAgents []string `json:"detected_agents"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal detect json: %v\nbody=%s", err, out.String())
	}

	if len(resp.DetectedAgents) == 0 {
		t.Fatal("expected at least one detected agent")
	}
	if !contains(resp.DetectedAgents, "claude") {
		t.Fatalf("detected_agents = %v, want claude", resp.DetectedAgents)
	}
}

func TestRunWindowsOptions_JSON(t *testing.T) {
	cat := &fakeExecCatalog{harnesses: []model.Harness{
		{ID: "openspec", Name: "OpenSpec", Description: "Specs CLI", InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}, Agents: []model.Agent{model.AgentClaude}},
		{ID: "permissions", Name: "Permissions", Description: "Safe defaults", InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}, Agents: []model.Agent{model.AgentClaude}},
		{ID: "kb-creator", Name: "KB Creator", Description: "Project knowledge base", InstallModes: []model.InstallMode{model.ModeFull}, Agents: []model.Agent{model.AgentClaude}},
	}}

	var out bytes.Buffer
	if err := headless.RunWindowsOptions(cat, model.AgentClaude, &out); err != nil {
		t.Fatalf("RunWindowsOptions() error = %v", err)
	}

	var resp struct {
		Modes []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"modes"`
		ForcedComponents []struct {
			ID string `json:"id"`
		} `json:"forced_components"`
		CustomComponents []struct {
			ID string `json:"id"`
		} `json:"custom_components"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal options json: %v\nbody=%s", err, out.String())
	}

	if len(resp.Modes) != 3 {
		t.Fatalf("modes len = %d, want 3", len(resp.Modes))
	}
	if resp.Modes[0].Label != "Quick" {
		t.Fatalf("first mode label = %q, want Quick", resp.Modes[0].Label)
	}
	if !containsComponent(resp.ForcedComponents, "permissions") {
		t.Fatalf("forced_components = %+v, want permissions", resp.ForcedComponents)
	}
	if containsComponent(resp.CustomComponents, "permissions") {
		t.Fatalf("custom_components = %+v, did not expect permissions duplicate", resp.CustomComponents)
	}
}

func TestRunWindowsInstall_JSONStream(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
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
		if downloadFn != nil {
			downloadFn(extinstaller.DownloadEvent{Type: extinstaller.DownloadStarted, URL: "https://example.test/ext-h.zip", Message: "Downloading ext-h."})
			downloadFn(extinstaller.DownloadEvent{Type: extinstaller.DownloadFinished, URL: "https://example.test/ext-h.zip", Message: "Finished downloading ext-h."})
		}
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	params := headless.ParsedFlags{
		HomeDir: t.TempDir(),
		Yes:     true,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsInstall(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsInstall() exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := splitJSONLines(out.String())
	if len(lines) < 2 {
		t.Fatalf("expected multiple json stream lines, got %d: %q", len(lines), out.String())
	}

	var first struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("first line is not valid json: %v\nline=%s", err, lines[0])
	}
	if first.Type == "" {
		t.Fatalf("first event missing type: %s", lines[0])
	}

	foundStepStarted := false
	foundStepSucceeded := false
	foundDownloadStarted := false
	foundDownloadFinished := false
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
		if evt.Type == "step_succeeded" && evt.StepID != "" {
			foundStepSucceeded = true
		}
		if evt.Type == "download_started" && evt.StepID != "" {
			foundDownloadStarted = true
		}
		if evt.Type == "download_finished" && evt.StepID != "" {
			foundDownloadFinished = true
		}
	}
	if !foundStepStarted {
		t.Fatal("expected at least one step_started event")
	}
	if !foundStepSucceeded {
		t.Fatal("expected at least one step_succeeded event")
	}
	if !foundDownloadStarted {
		t.Fatal("expected at least one download_started event")
	}
	if !foundDownloadFinished {
		t.Fatal("expected at least one download_finished event")
	}

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

func TestRunWindowsInstall_WritesJSONStreamToEventsFile(t *testing.T) {
	h := model.Harness{
		ID:           "ext-h",
		Type:         model.HarnessExternal,
		External:     &model.External{Method: "npm"},
		InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull},
	}
	cat := &fakeExecCatalog{harnesses: []model.Harness{h}}
	reg := &fakeExecRegistry{adapters: map[model.Agent]install.AgentAdapter{
		model.AgentClaude: fakeExecAdapter{agent: model.AgentClaude},
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
		if downloadFn != nil {
			downloadFn(extinstaller.DownloadEvent{Type: extinstaller.DownloadStarted, URL: "https://example.test/ext-h.zip", Message: "Downloading ext-h."})
		}
		return extinstaller.Result{}, nil
	})
	defer restoreExt()

	eventsFile := t.TempDir() + "/windows-events.jsonl"
	params := headless.ParsedFlags{
		HomeDir:           t.TempDir(),
		Yes:               true,
		WindowsEventsFile: eventsFile,
		Intent: install.Intent{
			Agents: []model.Agent{model.AgentClaude},
			Mode:   model.ModeLite,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsInstall(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsInstall() exit = %d; output:\n%s", exitCode, out.String())
	}

	raw, err := os.ReadFile(eventsFile)
	if err != nil {
		t.Fatalf("read events file: %v", err)
	}

	lines := splitJSONLines(string(raw))
	if len(lines) < 2 {
		t.Fatalf("expected multiple json lines in events file, got %d: %q", len(lines), string(raw))
	}

	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last file line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "install_finished" || !last.Success {
		t.Fatalf("last file event = %+v, want install_finished success=true", last)
	}
}

func TestRunWindowsUninstall_JSONStream(t *testing.T) {
	homeDir := t.TempDir()
	cat, reg := minimalCatalogAndRegistry(homeDir)

	restoreSnap := uninstall.SetSnapshotCreate(func(dir string, paths []string) (backup.Manifest, error) {
		return backup.Manifest{ID: "snap"}, nil
	})
	defer restoreSnap()

	restoreMarker := uninstall.SetMarkerRemovalFn(func(path, sectionID string) error {
		return nil
	})
	defer restoreMarker()

	restoreStale := uninstall.SetStalePurgeFn(func(path string) error {
		return nil
	})
	defer restoreStale()

	params := headless.ParsedUninstallFlags{
		HomeDir: homeDir,
		Yes:     true,
		Intent: uninstall.Intent{
			Agents:   []model.Agent{model.AgentClaude},
			Mode:     model.ModeLite,
			Strategy: uninstall.StrategyTargeted,
		},
	}

	var out bytes.Buffer
	exitCode := headless.RunWindowsUninstall(params, cat, reg, &out)
	if exitCode != 0 {
		t.Fatalf("RunWindowsUninstall() exit = %d; output:\n%s", exitCode, out.String())
	}

	lines := splitJSONLines(out.String())
	if len(lines) < 2 {
		t.Fatalf("expected multiple json stream lines, got %d: %q", len(lines), out.String())
	}

	foundStepStarted := false
	foundStepSucceeded := false
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
		if evt.Type == "step_succeeded" && evt.StepID != "" {
			foundStepSucceeded = true
		}
	}
	if !foundStepStarted {
		t.Fatal("expected at least one step_started event")
	}
	if !foundStepSucceeded {
		t.Fatal("expected at least one step_succeeded event")
	}

	var last struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line is not valid json: %v\nline=%s", err, lines[len(lines)-1])
	}
	if last.Type != "uninstall_finished" || !last.Success {
		t.Fatalf("last event = %+v, want uninstall_finished success=true", last)
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

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func containsComponent(cs []struct {
	ID string `json:"id"`
}, want string) bool {
	for _, c := range cs {
		if c.ID == want {
			return true
		}
	}
	return false
}
