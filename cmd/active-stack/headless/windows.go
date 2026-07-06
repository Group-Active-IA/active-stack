package headless

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/pipeline"
	"github.com/Group-Active-IA/active-stack/internal/system"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

type windowsDetectResponse struct {
	DetectedAgents []string `json:"detected_agents"`
}

type windowsOptionsResponse struct {
	Modes            []windowsModeOption      `json:"modes"`
	ForcedComponents []windowsComponentOption `json:"forced_components"`
	CustomComponents []windowsComponentOption `json:"custom_components"`
}

type windowsModeOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type windowsComponentOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Recommended bool   `json:"recommended,omitempty"`
}

type windowsInstallEvent struct {
	Type      string `json:"type"`
	Phase     string `json:"phase,omitempty"`
	StepID    string `json:"step_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Details   string `json:"details,omitempty"`
	Success   bool   `json:"success,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func RunWindowsDetect(homeDir string, _ install.Registry, w io.Writer) error {
	states := system.ScanConfigs(homeDir)
	detected := make([]string, 0, len(states))
	for _, state := range states {
		if !state.Exists || !state.IsDirectory {
			continue
		}
		if agent, ok := configStateAgent(state.Agent); ok {
			detected = append(detected, string(agent))
		}
	}
	return json.NewEncoder(w).Encode(windowsDetectResponse{DetectedAgents: detected})
}

func RunWindowsOptions(cat install.Catalog, agent model.Agent, w io.Writer) error {
	customHarnesses, err := install.SelectHarnesses(cat, install.Intent{
		Agents: []model.Agent{agent},
		Mode:   model.ModeCustom,
	})
	if err != nil {
		return err
	}

	resp := windowsOptionsResponse{
		Modes: []windowsModeOption{
			{ID: string(model.ModeLite), Label: "Quick", Description: "Fast setup to start working right away."},
			{ID: string(model.ModeFull), Label: "Complete", Description: "Full recommended setup with all key tools."},
			{ID: string(model.ModeCustom), Label: "Custom", Description: "Choose exactly what to install."},
		},
		ForcedComponents: make([]windowsComponentOption, 0, 1),
		CustomComponents: make([]windowsComponentOption, 0, len(customHarnesses)),
	}

	seen := make(map[string]bool, len(customHarnesses))
	for _, h := range customHarnesses {
		if seen[h.ID] {
			continue
		}
		seen[h.ID] = true

		item := windowsComponentOption{
			ID:          h.ID,
			Label:       h.Name,
			Description: windowsComponentDescription(h),
			Recommended: h.InMode(model.ModeLite) || h.InMode(model.ModeFull),
		}
		if h.ID == install.SecurityFirstHarnessID {
			resp.ForcedComponents = append(resp.ForcedComponents, item)
			continue
		}
		resp.CustomComponents = append(resp.CustomComponents, item)
	}

	return json.NewEncoder(w).Encode(resp)
}

func RunWindowsInstall(params ParsedFlags, cat install.Catalog, reg install.Registry, w io.Writer) int {
	eventWriter := w
	var file *os.File
	if params.WindowsEventsFile != "" {
		if err := os.MkdirAll(filepath.Dir(params.WindowsEventsFile), 0o755); err == nil {
			if f, err := os.Create(params.WindowsEventsFile); err == nil {
				file = f
				eventWriter = io.MultiWriter(w, f)
			}
		}
	}
	if file != nil {
		defer file.Close()
	}

	enc := json.NewEncoder(eventWriter)
	emit := func(event windowsInstallEvent) {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
		_ = enc.Encode(event)
	}

	var lastStage pipeline.Stage

	emit(windowsInstallEvent{
		Type:    "phase_started",
		Phase:   "install",
		Message: "Starting installation.",
	})

	params.ProgressEventFn = func(e pipeline.ProgressEvent) {
		if e.Stage != lastStage {
			lastStage = e.Stage
			emit(windowsInstallEvent{
				Type:    "phase_started",
				Phase:   string(e.Stage),
				Message: windowsPhaseMessage(e.Stage),
			})
		}
		emit(mapProgressEvent(e))
	}
	params.DownloadEventFn = func(e install.DownloadEvent) {
		emit(windowsInstallEvent{
			Type:    e.Type,
			Phase:   "apply",
			StepID:  e.StepID,
			Message: e.Message,
			Details: e.URL,
		})
	}

	var plain bytes.Buffer
	exitCode := RunHeadless(params, cat, reg, &plain)
	for _, line := range strings.Split(strings.ReplaceAll(plain.String(), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		emit(windowsInstallEvent{
			Type:    "step_output",
			Phase:   "install",
			Message: line,
			Details: line,
		})
	}

	emit(windowsInstallEvent{
		Type:    "install_finished",
		Phase:   "install",
		Success: exitCode == 0,
		Message: installFinishedMessage(exitCode),
	})

	return exitCode
}

func RunWindowsUninstall(params ParsedUninstallFlags, cat uninstall.Catalog, reg uninstall.Registry, w io.Writer) int {
	enc := json.NewEncoder(w)
	emit := func(event windowsInstallEvent) {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
		_ = enc.Encode(event)
	}

	var lastStage pipeline.Stage

	emit(windowsInstallEvent{
		Type:    "phase_started",
		Phase:   "uninstall",
		Message: "Starting uninstall.",
	})

	progressFn := func(e pipeline.ProgressEvent) {
		if e.Stage != lastStage {
			lastStage = e.Stage
			emit(windowsInstallEvent{
				Type:    "phase_started",
				Phase:   string(e.Stage),
				Message: windowsUninstallPhaseMessage(e.Stage),
			})
		}
		emit(mapProgressEvent(e))
	}

	params2 := params
	var plain bytes.Buffer
	if params2.Intent.Strategy == "" {
		params2.Intent.Strategy = uninstall.StrategyTargeted
	}

	// Mirror RunHeadlessUninstall, but with JSON-stream output instead of text.
	opts := uninstall.Options{
		HomeDir:    params2.HomeDir,
		Registry:   reg,
		OnProgress: progressFn,
	}

	plan, err := uninstall.BuildPlan(cat, params2.Intent, opts)
	if err != nil {
		emit(windowsInstallEvent{
			Type:    "step_failed",
			Phase:   "uninstall",
			Message: "Uninstall plan failed.",
			Details: err.Error(),
		})
		emit(windowsInstallEvent{
			Type:    "uninstall_finished",
			Phase:   "uninstall",
			Success: false,
			Message: "Uninstall failed.",
		})
		return 1
	}

	if params2.DryRun {
		for _, s := range plan.Prepare {
			emit(windowsInstallEvent{
				Type:    "step_started",
				Phase:   string(pipeline.StagePrepare),
				StepID:  s.ID(),
				Message: "Dry-run step planned.",
			})
			emit(windowsInstallEvent{
				Type:    "step_succeeded",
				Phase:   string(pipeline.StagePrepare),
				StepID:  s.ID(),
				Message: "Dry-run step listed.",
			})
		}
		for _, s := range plan.Apply {
			emit(windowsInstallEvent{
				Type:    "step_started",
				Phase:   string(pipeline.StageApply),
				StepID:  s.ID(),
				Message: "Dry-run step planned.",
			})
			emit(windowsInstallEvent{
				Type:    "step_succeeded",
				Phase:   string(pipeline.StageApply),
				StepID:  s.ID(),
				Message: "Dry-run step listed.",
			})
		}
		emit(windowsInstallEvent{
			Type:    "uninstall_finished",
			Phase:   "uninstall",
			Success: true,
			Message: "Uninstall finished successfully.",
		})
		return 0
	}

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithProgressFunc(progressFn),
	)
	result := orch.Execute(plan.StagePlan)
	if result.Err != nil {
		for _, line := range strings.Split(strings.ReplaceAll(plain.String(), "\r\n", "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			emit(windowsInstallEvent{
				Type:    "step_output",
				Phase:   "uninstall",
				Message: line,
				Details: line,
			})
		}
		emit(windowsInstallEvent{
			Type:    "uninstall_finished",
			Phase:   "uninstall",
			Success: false,
			Message: "Uninstall failed.",
			Details: result.Err.Error(),
		})
		return 1
	}

	emit(windowsInstallEvent{
		Type:    "uninstall_finished",
		Phase:   "uninstall",
		Success: true,
		Message: "Uninstall finished successfully.",
	})

	return 0
}

func mapProgressEvent(e pipeline.ProgressEvent) windowsInstallEvent {
	event := windowsInstallEvent{
		Phase:  string(e.Stage),
		StepID: e.StepID,
	}
	switch e.Status {
	case pipeline.StepStatusRunning:
		event.Type = "step_started"
		event.Message = "Step started."
	case pipeline.StepStatusSucceeded:
		event.Type = "step_succeeded"
		event.Message = "Step completed."
	case pipeline.StepStatusFailed:
		event.Type = "step_failed"
		event.Message = "Step failed."
	case pipeline.StepStatusRolledBack:
		event.Type = "rollback_finished"
		event.Message = "Rollback step completed."
	case pipeline.StepStatusDegraded:
		event.Type = "step_degraded"
		event.Message = "Step completed with warnings."
	default:
		event.Type = "step_output"
		event.Message = string(e.Status)
	}
	if e.Err != nil {
		event.Details = e.Err.Error()
	}
	return event
}

func windowsPhaseMessage(stage pipeline.Stage) string {
	switch stage {
	case pipeline.StagePrepare:
		return "Preparing installation."
	case pipeline.StageApply:
		return "Applying installation steps."
	case pipeline.StageRollback:
		return "Rolling back changes."
	default:
		return "Running installation."
	}
}

func windowsUninstallPhaseMessage(stage pipeline.Stage) string {
	switch stage {
	case pipeline.StagePrepare:
		return "Preparing uninstall."
	case pipeline.StageApply:
		return "Removing installed items."
	case pipeline.StageRollback:
		return "Restoring previous state."
	default:
		return "Running uninstall."
	}
}

func configStateAgent(id string) (model.Agent, bool) {
	switch id {
	case "claude-code":
		return model.AgentClaude, true
	case "opencode":
		return model.AgentOpenCode, true
	case "gemini-cli":
		return model.AgentGemini, true
	case "cursor":
		return model.AgentCursor, true
	case "vscode-copilot":
		return model.AgentVSCode, true
	case "codex":
		return model.AgentCodex, true
	case "antigravity":
		return model.AgentAntigravity, true
	case "windsurf":
		return model.AgentWindsurf, true
	default:
		return "", false
	}
}

func windowsComponentDescription(h model.Harness) string {
	switch {
	case h.ID == install.SecurityFirstHarnessID:
		return "Basic protection for safer setup. This is always installed."
	case h.Description != "":
		return h.Description
	case h.Type == model.HarnessExternal:
		return "Downloads and configures an external tool."
	case h.Type == model.HarnessConfig:
		return "Applies recommended configuration."
	case h.Type == model.HarnessSkill:
		return "Adds guided workflow helpers."
	default:
		return fmt.Sprintf("Installs %s.", h.Name)
	}
}

func installFinishedMessage(exitCode int) string {
	if exitCode == 0 {
		return "Installation finished successfully."
	}
	return "Installation failed."
}
