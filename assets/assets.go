// Package assets holds files that are embedded in the Active Stack binary.
// Skills in the assets/skills/ directory are installed via the "embed" method.
package assets

import "embed"

// SkillsFS holds the embedded skill SKILL.md files for all core skills that
// ship bundled with the installer (install method: embed).
//
// Structure: skills/<skillID>/SKILL.md
//
// Current embedded skills (openspec-core):
//   - openspec-init, openspec-explore, openspec-propose, openspec-spec
//   - openspec-design, openspec-tasks, openspec-apply, openspec-verify
//   - openspec-archive, openspec-onboard, judgment-day
//
// To add a new embedded skill: create assets/skills/<id>/SKILL.md and add
// the skill harness entry to internal/catalog/harnesses.yaml with method: embed.
//
//go:embed all:skills
var SkillsFS embed.FS

// CommandsFS holds the embedded slash-command .md files for each focused agent.
// Added in C-31 (TBD-4) — extends assets.go consistent with SkillsFS.
//
// Structure: commands/<agentVariantKey>/<path>.md
//
//   Claude  : commands/claude/active/starter-add.md
//               → invoked as /active:starter-add inside Claude Code
//               → full frontmatter (name/description/category/tags)
//   OpenCode: commands/opencode/active-starter-add.md
//               → invoked as /active-starter-add inside OpenCode
//               → flat frontmatter (description only)
//
// The command body is a thin wrapper that runs `active-stack starter add $ARGUMENTS`
// via the agent's bash execution. It does not reimplement any starter logic.
//
// To add a new command variant: create the .md file under
// commands/<variantKey>/... and extend the command installer accordingly.
//
//go:embed all:commands
var CommandsFS embed.FS
