// Package assets holds files that are embedded in the JR Stack binary.
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
