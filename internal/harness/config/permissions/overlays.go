package permissions

import "github.com/JuanCruzRobledo/jr-stack/internal/model"

// claudeCodeOverlayJSON sets Claude Code to acceptEdits mode.
//
// Orchestrator decision (security-first but usable):
//   - defaultMode: "acceptEdits" — auto-approve file edits, but ASK for Bash/commands.
//   - defaultMode: "bypassPermissions" is explicitly FORBIDDEN per orchestrator override.
//
// Valid Claude Code modes: "acceptEdits", "bypassPermissions", "default", "dontAsk", "plan".
// The deny-list blocks the most destructive operations unconditionally.
var claudeCodeOverlayJSON = []byte(`{
  "permissions": {
    "defaultMode": "acceptEdits",
    "deny": [
      "Bash(rm -rf /)",
      "Bash(sudo rm -rf /)",
      "Bash(rm -rf ~)",
      "Bash(sudo rm -rf ~)",
      "Read(.env)",
      "Read(.env.*)",
      "Edit(.env)",
      "Edit(.env.*)"
    ]
  }
}
`)

// openCodeOverlayJSON uses the OpenCode "permission" key (singular) with
// bash/read granularity. Git destructive commands require explicit confirmation.
var openCodeOverlayJSON = []byte(`{
  "permission": {
    "bash": {
      "*": "allow",
      "git commit *": "ask",
      "git push *": "ask",
      "git push": "ask",
      "git push --force *": "ask",
      "git rebase *": "ask",
      "git reset --hard *": "ask"
    },
    "read": {
      "*": "allow",
      "*.env": "deny",
      "*.env.*": "deny",
      "**/.env": "deny",
      "**/.env.*": "deny",
      "**/secrets/**": "deny",
      "**/credentials.json": "deny"
    }
  }
}
`)

// geminiCLIOverlayJSON sets Gemini CLI to "auto_edit" mode (auto-approve edit
// tools). Gemini CLI does not expose a deny-list equivalent in its schema;
// this is a known limitation documented in design.md §Risks.
var geminiCLIOverlayJSON = []byte(`{
  "general": {
    "defaultApprovalMode": "auto_edit"
  }
}
`)

// vscodeCopilotOverlayJSON enables auto-approve for VS Code Copilot chat tools.
// Note: VS Code Copilot's settings.json does not support a deny-list granularity
// equivalent to Claude's. The auto_approve is binary (on/off). This is a known
// limitation documented in design.md §Risks. The field uses dot-notation as a
// flat key (VS Code convention).
var vscodeCopilotOverlayJSON = []byte(`{
  "chat.tools.autoApprove": true
}
`)

// agentOverlay returns the permission overlay bytes for the given agent, or nil
// if the agent does not support permission injection via a settings.json file.
// A nil return is an EXPLICIT no-op — not a missing case. Agents without
// settings-file injection have their permissions managed via other mechanisms
// (IDE UI, cli-config.json, Artifact Review Policy) documented in the comments.
func agentOverlay(agent model.Agent) []byte {
	switch agent {
	case model.AgentClaude:
		return claudeCodeOverlayJSON
	case model.AgentOpenCode:
		return openCodeOverlayJSON
	case model.AgentGemini:
		return geminiCLIOverlayJSON
	case model.AgentVSCode:
		return vscodeCopilotOverlayJSON
	case model.AgentCursor:
		// Cursor manages permissions via cli-config.json, not settings.json.
		// No injectable overlay — explicit no-op.
		return nil
	case model.AgentCodex:
		// Codex has no known settings.json path for permission injection.
		// Explicit no-op.
		return nil
	case model.AgentAntigravity:
		// Antigravity manages permissions via IDE UI (Artifact Review Policy /
		// Terminal Command Auto Execution). No injectable settings.json schema.
		// Explicit no-op.
		return nil
	default:
		// Unknown agent (e.g. Windsurf — marked TBD in design.md). Explicit no-op.
		return nil
	}
}
