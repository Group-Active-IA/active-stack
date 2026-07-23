package headless

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
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
	Modes             []windowsModeOption      `json:"modes"`
	ForcedComponents  []windowsComponentOption `json:"forced_components"`
	CustomComponents  []windowsComponentOption `json:"custom_components"`
	TierCapable       bool                     `json:"tier_capable"`
	TierCapableAgents []string                 `json:"tier_capable_agents"`
	PermissionTiers   []windowsPermissionTier  `json:"permission_tiers"`
}

type windowsModeOption struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	LongDescription string `json:"long_description,omitempty"`
}

// windowsPermissionTier describes one permission tier for the GUI's tier
// selection screen (D3, design.md). PermissionTiers is always the same three
// entries in display order; only TierCapable / TierCapableAgents vary with
// the requested agent set.
type windowsPermissionTier struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	Default         bool   `json:"default"`
	Warning         string `json:"warning,omitempty"`
	LongDescription string `json:"long_description,omitempty"`
}

type windowsComponentOption struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	Recommended     bool   `json:"recommended,omitempty"`
	LongDescription string `json:"long_description,omitempty"`
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

func RunWindowsDetect(homeDir string, reg install.Registry, w io.Writer) error {
	all := detectAgents(homeDir)
	if reg == nil {
		return json.NewEncoder(w).Encode(windowsDetectResponse{DetectedAgents: all})
	}
	// Only surface agents that have a registered adapter. This mirrors
	// tui.AvailableAgentsList and prevents "no adapter registered for agent X"
	// failures in BuildPlan when the user selects an agent detected on the
	// machine but not yet implemented by the installer.
	supported := make([]string, 0, len(all))
	for _, id := range all {
		if _, ok := reg.Get(model.Agent(id)); ok {
			supported = append(supported, id)
		}
	}
	return json.NewEncoder(w).Encode(windowsDetectResponse{DetectedAgents: supported})
}

// detectAgents scans homeDir for agent configs and returns the detected agent
// ids as strings. Extracted from RunWindowsDetect (design D6,
// windows-contract-hub-operations) so RunWindowsUninstallOptions can report
// the same detected_agents without duplicating the ScanConfigs +
// configStateAgent loop.
func detectAgents(homeDir string) []string {
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
	return detected
}

func RunWindowsOptions(cat install.Catalog, agents []model.Agent, lang i18n.Lang, w io.Writer) error {
	// D5 (windows-contract-tier-multiagent): the picker universe is "what can
	// be picked", not "what will be installed" — install.SelectHarnesses with
	// an empty Custom intent only ever force-adds the permissions harness, so
	// custom_components was structurally always empty. CustomPickerHarnesses
	// answers the right question and is the single source of truth shared
	// with the TUI Custom picker (internal/tui/model.go).
	customHarnesses := install.CustomPickerHarnesses(cat, agents)

	resp := windowsOptionsResponse{
		Modes:             windowsModeOptions(lang),
		ForcedComponents:  make([]windowsComponentOption, 0, 1),
		CustomComponents:  make([]windowsComponentOption, 0, len(customHarnesses)),
		TierCapable:       model.TierCapable(agents),
		TierCapableAgents: tierCapableAgentIDs(agents),
		PermissionTiers:   windowsPermissionTiers(lang),
	}

	seen := make(map[string]bool, len(customHarnesses))
	for _, h := range customHarnesses {
		if seen[h.ID] {
			continue
		}
		seen[h.ID] = true

		item := windowsComponentOption{
			ID:              h.ID,
			Label:           h.Name,
			Description:     windowsComponentDescription(lang, h),
			Recommended:     h.InMode(model.ModeLite) || h.InMode(model.ModeFull),
			LongDescription: h.LongDescription.Localized(string(lang)),
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
	return runWindowsPipeline(params, cat, reg, w, "install_finished")
}

// runWindowsPipeline owns the JSON event-stream plumbing shared by
// "windows install" and "windows starters install" (design D3,
// windows-contract-hub-operations): the event-file tailer, `emit`,
// progress/download wiring, and the plain-output replay. It is parameterized
// by the terminal event's `type` so RunWindowsInstall passes
// "install_finished" while RunWindowsStartersInstall passes
// "starter_finished" — every other event in the stream is identical.
func runWindowsPipeline(params ParsedFlags, cat install.Catalog, reg install.Registry, w io.Writer, finishedEventType string) int {
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

	lang := params.Lang

	emit(windowsInstallEvent{
		Type:    "phase_started",
		Phase:   "install",
		Message: i18n.T(lang, "windows.install.starting"),
	})

	params.ProgressEventFn = func(e pipeline.ProgressEvent) {
		if e.Stage != lastStage {
			lastStage = e.Stage
			emit(windowsInstallEvent{
				Type:    "phase_started",
				Phase:   string(e.Stage),
				Message: windowsPhaseMessage(lang, e.Stage),
			})
		}
		emit(mapProgressEvent(lang, e))
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
		Type:    finishedEventType,
		Phase:   "install",
		Success: exitCode == 0,
		Message: installFinishedMessage(lang, exitCode),
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
	lang := params.Lang

	emit(windowsInstallEvent{
		Type:    "phase_started",
		Phase:   "uninstall",
		Message: i18n.T(lang, "windows.uninstall.starting"),
	})

	progressFn := func(e pipeline.ProgressEvent) {
		if e.Stage != lastStage {
			lastStage = e.Stage
			emit(windowsInstallEvent{
				Type:    "phase_started",
				Phase:   string(e.Stage),
				Message: windowsUninstallPhaseMessage(lang, e.Stage),
			})
		}
		emit(mapProgressEvent(lang, e))
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
			Message: i18n.T(lang, "windows.uninstall.plan_failed"),
			Details: err.Error(),
		})
		emit(windowsInstallEvent{
			Type:    "uninstall_finished",
			Phase:   "uninstall",
			Success: false,
			Message: i18n.T(lang, "windows.uninstall.finished_fail"),
		})
		return 1
	}

	if params2.DryRun {
		for _, s := range plan.Prepare {
			emit(windowsInstallEvent{
				Type:    "step_started",
				Phase:   string(pipeline.StagePrepare),
				StepID:  s.ID(),
				Message: i18n.T(lang, "windows.dryrun.planned"),
			})
			emit(windowsInstallEvent{
				Type:    "step_succeeded",
				Phase:   string(pipeline.StagePrepare),
				StepID:  s.ID(),
				Message: i18n.T(lang, "windows.dryrun.listed"),
			})
		}
		for _, s := range plan.Apply {
			emit(windowsInstallEvent{
				Type:    "step_started",
				Phase:   string(pipeline.StageApply),
				StepID:  s.ID(),
				Message: i18n.T(lang, "windows.dryrun.planned"),
			})
			emit(windowsInstallEvent{
				Type:    "step_succeeded",
				Phase:   string(pipeline.StageApply),
				StepID:  s.ID(),
				Message: i18n.T(lang, "windows.dryrun.listed"),
			})
		}
		emit(windowsInstallEvent{
			Type:    "uninstall_finished",
			Phase:   "uninstall",
			Success: true,
			Message: i18n.T(lang, "windows.uninstall.finished_ok"),
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
			Message: i18n.T(lang, "windows.uninstall.finished_fail"),
			Details: result.Err.Error(),
		})
		return 1
	}

	emit(windowsInstallEvent{
		Type:    "uninstall_finished",
		Phase:   "uninstall",
		Success: true,
		Message: i18n.T(lang, "windows.uninstall.finished_ok"),
	})

	return 0
}

func mapProgressEvent(lang i18n.Lang, e pipeline.ProgressEvent) windowsInstallEvent {
	event := windowsInstallEvent{
		Phase:  string(e.Stage),
		StepID: e.StepID,
	}
	switch e.Status {
	case pipeline.StepStatusRunning:
		event.Type = "step_started"
		event.Message = i18n.T(lang, "windows.step.started")
	case pipeline.StepStatusSucceeded:
		event.Type = "step_succeeded"
		event.Message = i18n.T(lang, "windows.step.succeeded")
	case pipeline.StepStatusFailed:
		event.Type = "step_failed"
		event.Message = i18n.T(lang, "windows.step.failed")
	case pipeline.StepStatusRolledBack:
		event.Type = "rollback_finished"
		event.Message = i18n.T(lang, "windows.step.rolled_back")
	case pipeline.StepStatusDegraded:
		event.Type = "step_degraded"
		event.Message = i18n.T(lang, "windows.step.degraded")
	default:
		event.Type = "step_output"
		event.Message = string(e.Status)
	}
	if e.Err != nil {
		event.Details = e.Err.Error()
	}
	return event
}

func windowsPhaseMessage(lang i18n.Lang, stage pipeline.Stage) string {
	switch stage {
	case pipeline.StagePrepare:
		return i18n.T(lang, "windows.phase.prepare")
	case pipeline.StageApply:
		return i18n.T(lang, "windows.phase.apply")
	case pipeline.StageRollback:
		return i18n.T(lang, "windows.phase.rollback")
	default:
		return i18n.T(lang, "windows.phase.default")
	}
}

func windowsUninstallPhaseMessage(lang i18n.Lang, stage pipeline.Stage) string {
	switch stage {
	case pipeline.StagePrepare:
		return i18n.T(lang, "windows.uninstall.phase.prepare")
	case pipeline.StageApply:
		return i18n.T(lang, "windows.uninstall.phase.apply")
	case pipeline.StageRollback:
		return i18n.T(lang, "windows.uninstall.phase.rollback")
	default:
		return i18n.T(lang, "windows.uninstall.phase.default")
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

func windowsComponentDescription(lang i18n.Lang, h model.Harness) string {
	switch {
	case h.ID == install.SecurityFirstHarnessID:
		return i18n.T(lang, "component.fallback.security")
	case h.Description.Localized(string(lang)) != "":
		return h.Description.Localized(string(lang))
	case h.Type == model.HarnessExternal:
		return i18n.T(lang, "component.fallback.external")
	case h.Type == model.HarnessConfig:
		return i18n.T(lang, "component.fallback.config")
	case h.Type == model.HarnessSkill:
		return i18n.T(lang, "component.fallback.skill")
	default:
		return i18n.Tf(lang, "component.fallback.generic_fmt", h.Name)
	}
}

// tierCapableAgentIDs returns the subset of the requested agents that are
// tier-capable, as agent id strings, in the order they were requested. This
// lets the GUI recompute capability locally when the user toggles agents,
// avoiding a re-query (D3, design.md).
func tierCapableAgentIDs(agents []model.Agent) []string {
	out := make([]string, 0, len(agents))
	for _, a := range agents {
		if model.TierCapable([]model.Agent{a}) {
			out = append(out, string(a))
		}
	}
	return out
}

// windowsPermissionTiers lists the three permission tiers in display order
// (estricto, balanceado, bypass), with balanceado marked as the default and
// bypass carrying its autonomy warning (D3, design.md; mirrors the TUI's
// tierOrder / bypassWarning in internal/tui/permissions_screen.go). Labels,
// descriptions, long descriptions, and the bypass warning are sourced from
// the i18n tables so they stay language-consistent (D5, i18n-engine-locales):
// under en the labels are Strict/Balanced/Bypass; under es they are
// Estricto/Balanceado/Bypass. Tier ids never change.
func windowsPermissionTiers(lang i18n.Lang) []windowsPermissionTier {
	return []windowsPermissionTier{
		{
			ID:              string(model.TierEstricto),
			Label:           i18n.T(lang, "tier.estricto.label"),
			Description:     i18n.T(lang, "tier.estricto.desc"),
			LongDescription: i18n.T(lang, "tier.estricto.long"),
		},
		{
			ID:              string(model.TierBalanceado),
			Label:           i18n.T(lang, "tier.balanceado.label"),
			Description:     i18n.T(lang, "tier.balanceado.desc"),
			LongDescription: i18n.T(lang, "tier.balanceado.long"),
			Default:         true,
		},
		{
			ID:              string(model.TierBypass),
			Label:           i18n.T(lang, "tier.bypass.label"),
			Description:     i18n.T(lang, "tier.bypass.desc"),
			LongDescription: i18n.T(lang, "tier.bypass.long"),
			Warning:         i18n.T(lang, "tier.bypass.warning"),
		},
	}
}

func installFinishedMessage(lang i18n.Lang, exitCode int) string {
	if exitCode == 0 {
		return i18n.T(lang, "windows.install.finished_ok")
	}
	return i18n.T(lang, "windows.install.finished_fail")
}
