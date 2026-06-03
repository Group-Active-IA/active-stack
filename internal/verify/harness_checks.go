package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// sddOrchestratorSectionID is the marker section written by the config harness
// installer (internal/harness/config). It is the string used by
// filemerge.InjectMarkdownSection when injecting the orchestrator block.
const sddOrchestratorSectionID = "sdd-orchestrator"

// ChecksForHarness derives the verify.Check slice for a single harness,
// one check per (harness × adapter). It dispatches by HarnessType:
//
//   - skill    → SKILL.md exists and is non-empty in adapter.SkillsDir
//   - config   → idempotent marker present exactly once in InstructionsPath
//                (special case "permissions" → checks SettingsPath for key)
//   - external → MCP config parseable (hard) + binary in PATH (Soft)
//
// Paths are ALWAYS resolved via adapters — never hardcoded.
func ChecksForHarness(h model.Harness, adapters []Adapter, homeDir string) []Check {
	var checks []Check
	for _, adapter := range adapters {
		checks = append(checks, checksForHarnessAdapter(h, adapter, homeDir)...)
	}
	return checks
}

// checksForHarnessAdapter derives checks for one (harness, adapter) pair.
func checksForHarnessAdapter(h model.Harness, adapter Adapter, homeDir string) []Check {
	switch h.Type {
	case model.HarnessSkill:
		return checkSkill(h, adapter, homeDir)
	case model.HarnessConfig:
		if h.ID == "permissions" {
			return checkPermissions(h, adapter, homeDir)
		}
		return checkConfig(h, adapter, homeDir)
	case model.HarnessExternal:
		return checkExternal(h, adapter, homeDir)
	default:
		return nil
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// skill checks
// ─────────────────────────────────────────────────────────────────────────────

func checkSkill(h model.Harness, adapter Adapter, homeDir string) []Check {
	skillsDir := adapter.SkillsDir(homeDir)
	agentID := string(adapter.Agent())

	return []Check{
		{
			ID:          fmt.Sprintf("skill:%s:%s", h.ID, agentID),
			Description: fmt.Sprintf("SKILL.md present and non-empty for %s/%s", agentID, h.ID),
			// A best-effort harness that failed to install is a warning, not a
			// hard failure: the operator was already warned at install time, and
			// the verify hook must not roll back a successful (soft) install.
			Soft: h.BestEffort,
			Run: func(_ context.Context) error {
				skillMD := filepath.Join(skillsDir, h.ID, "SKILL.md")
				info, err := os.Stat(skillMD)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("SKILL.md not found at %q", skillMD)
					}
					return fmt.Errorf("stat SKILL.md: %w", err)
				}
				if info.Size() == 0 {
					return fmt.Errorf("SKILL.md is empty at %q", skillMD)
				}
				return nil
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// config checks
// ─────────────────────────────────────────────────────────────────────────────

// checkConfig verifies that the injected block's marker appears exactly once
// in the agent's instructions file. Uses the same sectionID the config installer
// writes (internal/harness/config/inject.go: "sdd-orchestrator").
func checkConfig(h model.Harness, adapter Adapter, homeDir string) []Check {
	// Primary-agent delivery (e.g. OpenCode) writes to the settings JSON, not
	// the instructions file — verify it where it actually lives.
	if adapter.ConfigDelivery() == model.ConfigDeliveryPrimaryAgent {
		return checkPrimaryAgentConfig(h, adapter, homeDir)
	}

	instrPath := adapter.InstructionsPath(homeDir)
	agentID := string(adapter.Agent())
	openMarker := "<!-- jr-stack:" + sddOrchestratorSectionID + " -->"

	return []Check{
		{
			ID:          fmt.Sprintf("config:%s:%s", h.ID, agentID),
			Description: fmt.Sprintf("idempotent marker present exactly once in %s instructions (%s)", agentID, h.ID),
			Run: func(_ context.Context) error {
				data, err := os.ReadFile(instrPath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("instructions file not found at %q", instrPath)
					}
					return fmt.Errorf("read instructions file: %w", err)
				}
				count := strings.Count(string(data), openMarker)
				switch {
				case count == 0:
					return fmt.Errorf("marker %q not found in %q (harness not installed?)", openMarker, instrPath)
				case count > 1:
					return fmt.Errorf("marker %q appears %d times in %q (idempotency violation)", openMarker, count, instrPath)
				}
				return nil
			},
		},
	}
}

// permissionsKeyFor returns the top-level JSON key that the permissions
// installer writes for the given agent.
//
// Schema source: internal/harness/config/permissions/overlays.go
//
//   - Claude  → "permissions" (plural)  — claudeCodeOverlayJSON L13
//   - OpenCode → "permission" (singular) — openCodeOverlayJSON L32
//   - default  → "permissions" (plural)  — matches every other overlay that
//     either uses the same schema (Gemini uses "general", VS Code uses
//     "chat.tools.autoApprove") or injects no overlay at all. The default
//     covers future agents until their overlay lands — a failing verify check
//     will catch any mismatch.
func permissionsKeyFor(agent model.Agent) string {
	switch agent {
	case model.AgentOpenCode:
		return "permission"
	default:
		return "permissions"
	}
}

// checkPermissions verifies that the permissions key exists in the agent's
// settings file (written by internal/harness/config/permissions).
// The key name is resolved per-agent via permissionsKeyFor to match the exact
// JSON key each agent's overlay writes (e.g. "permission" for OpenCode,
// "permissions" for Claude). The error message names the searched key so that
// future schema drift is self-diagnosing (D4 in design.md).
func checkPermissions(h model.Harness, adapter Adapter, homeDir string) []Check {
	settingsPath := adapter.SettingsPath(homeDir)
	agentID := string(adapter.Agent())
	key := permissionsKeyFor(adapter.Agent())

	return []Check{
		{
			ID:          fmt.Sprintf("permissions:%s:%s", h.ID, agentID),
			Description: fmt.Sprintf("permissions key present in %s settings", agentID),
			Run: func(_ context.Context) error {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("settings file not found at %q", settingsPath)
					}
					return fmt.Errorf("read settings file: %w", err)
				}
				var raw map[string]json.RawMessage
				if err := json.Unmarshal(data, &raw); err != nil {
					return fmt.Errorf("parse settings file %q: %w", settingsPath, err)
				}
				if _, ok := raw[key]; !ok {
					return fmt.Errorf("expected key %q not found in %q", key, settingsPath)
				}
				return nil
			},
		},
	}
}

// checkPrimaryAgentConfig verifies that a config harness delivered as a primary
// agent is registered under agent.<id> with mode:primary in the settings JSON
// (e.g. opencode.json) — the entry that makes it a tab-able agent. Mirrors the
// install-time registration in internal/harness/config/primary_agent.go.
func checkPrimaryAgentConfig(h model.Harness, adapter Adapter, homeDir string) []Check {
	settingsPath := adapter.SettingsPath(homeDir)
	agentID := string(adapter.Agent())

	return []Check{
		{
			ID:          fmt.Sprintf("config:%s:%s", h.ID, agentID),
			Description: fmt.Sprintf("primary agent %q registered (mode:primary) in %s settings", h.ID, agentID),
			Run: func(_ context.Context) error {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("settings file not found at %q", settingsPath)
					}
					return fmt.Errorf("read settings file: %w", err)
				}
				var root struct {
					Agent map[string]struct {
						Mode string `json:"mode"`
					} `json:"agent"`
				}
				if err := json.Unmarshal(data, &root); err != nil {
					return fmt.Errorf("parse settings file %q: %w", settingsPath, err)
				}
				entry, ok := root.Agent[h.ID]
				if !ok {
					return fmt.Errorf("agent %q not found in %q (harness not installed?)", h.ID, settingsPath)
				}
				if entry.Mode != "primary" {
					return fmt.Errorf("agent %q has mode %q, want \"primary\" (not tab-able) in %q", h.ID, entry.Mode, settingsPath)
				}
				return nil
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// external checks
// ─────────────────────────────────────────────────────────────────────────────

func checkExternal(h model.Harness, adapter Adapter, homeDir string) []Check {
	if h.External == nil {
		return nil
	}

	var checks []Check

	switch h.External.Method {
	case "mcp":
		// Hard check: MCP config file is present and parseable JSON.
		checks = append(checks, checkMCPConfig(h, adapter, homeDir))
		// Soft check: remote endpoint is reachable (warning only).
		if h.External.URL != "" {
			checks = append(checks, checkMCPEndpoint(h))
		}

	case "homebrew", "npm", "go-install", "download":
		// Soft check: binary in PATH after install.
		// Network/exec failures → warning, not hard failure.
		checks = append(checks, checkBinaryInPATH(h))
	}

	return checks
}

func checkMCPConfig(h model.Harness, adapter Adapter, homeDir string) Check {
	mcpPath := adapter.MCPConfigPath(homeDir, h.ID)
	agentID := string(adapter.Agent())

	return Check{
		ID:          fmt.Sprintf("external:mcp-config:%s:%s", h.ID, agentID),
		Description: fmt.Sprintf("MCP config exists and is valid JSON at %s for %s", mcpPath, agentID),
		Soft:        false, // hard check: if we wrote it, it must be there
		Run: func(_ context.Context) error {
			data, err := os.ReadFile(mcpPath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("MCP config not found at %q", mcpPath)
				}
				return fmt.Errorf("read MCP config: %w", err)
			}
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("MCP config at %q is not valid JSON: %w", mcpPath, err)
			}
			return nil
		},
	}
}

func checkMCPEndpoint(h model.Harness) Check {
	return Check{
		ID:          fmt.Sprintf("external:mcp-endpoint:%s", h.ID),
		Description: fmt.Sprintf("MCP endpoint %s is reachable", h.External.URL),
		Soft:        true, // warning only — network may be down
		Run:         nil,  // skipped: network checks are out of scope for unit verify
	}
}

func checkBinaryInPATH(h model.Harness) Check {
	binaryName := binaryNameFromHarness(h)

	return Check{
		ID:          fmt.Sprintf("external:binary:%s", h.ID),
		Description: fmt.Sprintf("binary %q in PATH after install", binaryName),
		Soft:        true, // warning only — PATH may differ across shells
		Run: func(_ context.Context) error {
			_, err := exec.LookPath(binaryName)
			if err != nil {
				return fmt.Errorf("binary %q not found in PATH: %w", binaryName, err)
			}
			return nil
		},
	}
}

// binaryNameFromHarness derives the expected binary name from the harness.
func binaryNameFromHarness(h model.Harness) string {
	if h.External == nil || h.External.Pkg == "" {
		return h.ID
	}
	// Strip npm scope prefix: "@scope/name" → "name".
	pkg := h.External.Pkg
	if strings.HasPrefix(pkg, "@") {
		parts := strings.SplitN(pkg, "/", 2)
		if len(parts) == 2 {
			return filepath.Base(parts[1])
		}
	}
	return filepath.Base(pkg)
}
