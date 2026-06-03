package permissions

import (
	"encoding/json"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// securityFloorDeny is the single source of truth for the C-21 deny rules that
// apply to ALL tiers (including bypass). Claude Code evaluates deny → ask → allow,
// so deny ALWAYS wins regardless of defaultMode. This slice must NEVER be modified
// by tier composition — it is the floor that makes the security guarantee real.
//
// The five rules are the normative contract tested in permissions_test.go.
// The implementation is a superset (extra rules are allowed per design.md §D2).
var securityFloorDeny = []string{
	// ─── Normative floor (must-have, tested) ───
	"Read(.env)",
	"Read(.env.*)",
	"Bash(rm -rf /)",
	"Bash(rm -rf ~)",
	"Bash(git push --force:*)",
	// ─── Superset extras (defense in depth) ───
	"Bash(sudo rm -rf /)",
	"Bash(sudo rm -rf ~)",
	"Edit(.env)",
	"Edit(.env.*)",
}

// SecurityFloorDeny returns a copy of the security floor deny rules.
// Exported for testing; the internal slice must not be mutated.
func SecurityFloorDeny() []string {
	out := make([]string, len(securityFloorDeny))
	copy(out, securityFloorDeny)
	return out
}

// ── Claude Code overlay composition ─────────────────────────────────────────

// claudePermissionsShape is the struct that serializes to the Claude Code
// "permissions" block in settings.json.
type claudePermissionsShape struct {
	DefaultMode string   `json:"defaultMode"`
	Allow       []string `json:"allow,omitempty"`
	Deny        []string `json:"deny"`
}

// claudeSettingsShape is the top-level settings.json for Claude Code.
type claudeSettingsShape struct {
	Permissions claudePermissionsShape `json:"permissions"`
}

// claudeAllowList is the curated allow-list for tier balanceado.
// Kept as a package-level var so it can be referenced in tests if needed.
var claudeAllowList = []string{
	"Read",
	"Edit",
	"Bash(go test:*)",
	"Bash(go build:*)",
	"Bash(git status)",
	"Bash(git diff:*)",
}

// composeClaudeOverlay returns the JSON overlay bytes for Claude Code given a
// permission tier. Zero-value tier normalizes to TierBalanceado (never TierBypass).
//
// Claude semantics: deny → ask → allow; deny ALWAYS wins.
// The security floor is present in every tier.
func composeClaudeOverlay(tier model.PermissionTier) []byte {
	tier = tier.Normalize()

	var shape claudeSettingsShape
	shape.Permissions.Deny = SecurityFloorDeny()

	switch tier {
	case model.TierEstricto:
		shape.Permissions.DefaultMode = "default"
		// No allow-list for estricto — agent must ask for everything not denied.
	case model.TierBypass:
		shape.Permissions.DefaultMode = "bypassPermissions"
		// No allow-list needed — bypassPermissions auto-approves everything not denied.
	default: // TierBalanceado (and any unknown after normalize)
		shape.Permissions.DefaultMode = "default"
		shape.Permissions.Allow = claudeAllowList
	}

	b, err := json.Marshal(shape)
	if err != nil {
		// This should never happen with a static struct — panic so it surfaces in tests.
		panic("permissions: composeClaudeOverlay: marshal failed: " + err.Error())
	}
	return b
}

// ── OpenCode overlay composition (LAST-MATCH-WINS) ─────────────────────────
//
// CRITICAL: opencode evaluates pattern-objects with LAST-MATCH-WINS semantics —
// the OPPOSITE of Claude Code. If a deny rule appears BEFORE a wildcard "*", the
// wildcard overwrites it and the deny is silently ignored. The deny floor MUST be
// serialized LAST in every bash pattern-object.
//
// Because Go maps do NOT preserve insertion order, we use orderedPairs (a slice
// of key-value structs) that serializes to a JSON object in insertion order via a
// custom MarshalJSON. This is the ONLY safe way to guarantee deny-last.

// orderedPair is a single key-value entry in an ordered JSON object.
type orderedPair struct {
	Key   string
	Value string
}

// orderedBashBlock is an ordered collection of bash pattern-object entries.
// It serializes to a JSON object preserving insertion order (deny-last guarantee).
type orderedBashBlock []orderedPair

// MarshalJSON serializes the bash block as a JSON object in insertion order.
func (b orderedBashBlock) MarshalJSON() ([]byte, error) {
	buf := []byte{'{'}
	for i, p := range b {
		key, err := json.Marshal(p.Key)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(p.Value)
		if err != nil {
			return nil, err
		}
		buf = append(buf, key...)
		buf = append(buf, ':')
		buf = append(buf, val...)
		if i < len(b)-1 {
			buf = append(buf, ',')
		}
	}
	buf = append(buf, '}')
	return buf, nil
}

// opencodePermissionShape is the top-level opencode "permission" object.
// Fields use omitempty so unset keys don't pollute the output.
type opencodePermissionShape struct {
	Read              string           `json:"read,omitempty"`
	Edit              string           `json:"edit,omitempty"`
	Glob              string           `json:"glob,omitempty"`
	Grep              string           `json:"grep,omitempty"`
	List              string           `json:"list,omitempty"`
	Bash              orderedBashBlock `json:"bash"`
	Task              string           `json:"task,omitempty"`
	ExternalDirectory string           `json:"external_directory,omitempty"`
	Webfetch          string           `json:"webfetch,omitempty"`
	Websearch         string           `json:"websearch,omitempty"`
	LSP               string           `json:"lsp,omitempty"`
}

// opencodeSettingsShape is the top-level opencode.json structure.
type opencodeSettingsShape struct {
	Permission opencodePermissionShape `json:"permission"`
}

// composeOpencodeOverlay returns the JSON overlay bytes for opencode given a
// permission tier. Zero-value tier normalizes to TierBalanceado (never TierBypass).
//
// LAST-MATCH-WINS: deny rules MUST be the last entries in the bash pattern-object.
// orderedBashBlock guarantees serialization order (deny-last).
func composeOpencodeOverlay(tier model.PermissionTier) []byte {
	tier = tier.Normalize()

	var shape opencodeSettingsShape

	switch tier {
	case model.TierEstricto:
		// Everything → ask; external_directory → deny for extra hardening.
		// bash: wildcard ask first, then deny floor LAST (last-wins guarantee).
		shape.Permission = opencodePermissionShape{
			Read:              "ask",
			Edit:              "ask",
			Glob:              "ask",
			Grep:              "ask",
			List:              "ask",
			ExternalDirectory: "deny",
			Webfetch:          "ask",
			Websearch:         "ask",
			LSP:               "ask",
			Bash: orderedBashBlock{
				{Key: "*", Value: "ask"},
				// Deny floor — MUST come after the wildcard (last-wins).
				{Key: "rm -rf *", Value: "deny"},
				{Key: "rm -rf /*", Value: "deny"},
			},
		}

	case model.TierBypass:
		// Everything → allow; only the most catastrophic commands denied LAST.
		shape.Permission = opencodePermissionShape{
			Read:     "allow",
			Edit:     "allow",
			Glob:     "allow",
			Grep:     "allow",
			List:     "allow",
			Webfetch: "allow",
			Websearch: "allow",
			LSP:      "allow",
			Bash: orderedBashBlock{
				{Key: "*", Value: "allow"},
				// Catastrophic deny floor LAST (last-wins guarantee).
				{Key: "rm -rf /*", Value: "deny"},
				{Key: "rm -rf ~/", Value: "deny"},
			},
		}

	default: // TierBalanceado (and zero-value after normalize)
		// Safe operations auto-allowed; deny floor LAST in bash.
		shape.Permission = opencodePermissionShape{
			Read:      "allow",
			Edit:      "ask",
			Glob:      "allow",
			Grep:      "allow",
			List:      "allow",
			Webfetch:  "ask",
			Websearch: "ask",
			LSP:       "allow",
			Bash: orderedBashBlock{
				{Key: "*", Value: "ask"},
				// Safe operations allowed.
				{Key: "go test *", Value: "allow"},
				{Key: "go build *", Value: "allow"},
				{Key: "git status", Value: "allow"},
				{Key: "git diff *", Value: "allow"},
				// Deny floor MUST be last (last-wins guarantee).
				{Key: "rm -rf *", Value: "deny"},
				{Key: "rm -rf /*", Value: "deny"},
			},
		}
	}

	b, err := json.Marshal(shape)
	if err != nil {
		panic("permissions: composeOpencodeOverlay: marshal failed: " + err.Error())
	}
	return b
}

// ── Static overlays (gemini, vscode — tier-agnostic, TBD per design.md §D5) ──

// geminiCLIOverlayJSON sets Gemini CLI to "auto_edit" mode.
// Gemini CLI does not expose a deny-list equivalent in its schema;
// this is a known limitation. Tier mapeo: (TBD) — see design.md §D5.
var geminiCLIOverlayJSON = []byte(`{
  "general": {
    "defaultApprovalMode": "auto_edit"
  }
}
`)

// vscodeCopilotOverlayJSON enables auto-approve for VS Code Copilot chat tools.
// VS Code Copilot's settings.json does not support deny-list granularity.
// Tier mapeo: (TBD) — see design.md §D5.
var vscodeCopilotOverlayJSON = []byte(`{
  "chat.tools.autoApprove": true
}
`)

// agentOverlay returns the permission overlay bytes for the given agent and tier.
// Returns nil for agents with no injectable settings file (explicit no-op).
//
// Only claude and opencode differentiate by tier. gemini/vscode use static overlays
// regardless of tier (TBD per design.md §D5). Other agents are explicit no-ops.
//
// (TBD) Tier mapping for gemini: only exposes defaultApprovalMode (default|auto_edit|yolo)
// with no deny-list equivalent — see design.md §D5 for why this is deferred.
//
// (TBD) Tier mapping for vscode: only exposes chat.tools.autoApprove (bool, binary)
// with no deny-list granularity — see design.md §D5 for why this is deferred.
func agentOverlay(agent model.Agent, tier model.PermissionTier) []byte {
	switch agent {
	case model.AgentClaude:
		return composeClaudeOverlay(tier)
	case model.AgentOpenCode:
		return composeOpencodeOverlay(tier)
	case model.AgentGemini:
		return geminiCLIOverlayJSON
	case model.AgentVSCode:
		return vscodeCopilotOverlayJSON
	case model.AgentCursor:
		// Cursor manages permissions via cli-config.json, not settings.json.
		return nil
	case model.AgentCodex:
		// Codex has no known settings.json path for permission injection.
		return nil
	case model.AgentAntigravity:
		// Antigravity manages permissions via IDE UI. No injectable settings.json.
		return nil
	case model.AgentWindsurf:
		// Windsurf manages permissions via IDE UI. No injectable settings.json schema.
		return nil
	default:
		return nil
	}
}
