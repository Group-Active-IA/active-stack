package config

import (
	"fmt"
	"io/fs"
	"strings"
)

// CodexPackage is the content of a machine-level native OPSX delivery.
// Files are relative to the skill directory resolved by the agent adapter.
// Composition does not write or migrate a user's configuration.
type CodexPackage struct {
	Universal string
	Files     map[string]string
}

// ComposeCodexPackage composes the Codex skill and its universal rules.
// It is kept separate from legacy Compose until migration, rollback and
// uninstall can install the whole native delivery consistently.
func ComposeCodexPackage(toggles []string) (CodexPackage, error) {
	active := make(map[string]bool, len(toggles))
	for _, toggle := range toggles {
		if !knownToggles[toggle] {
			return CodexPackage{}, fmt.Errorf("codex package: unknown toggle %q", toggle)
		}
		active[toggle] = true
	}
	skill, err := fs.ReadFile(embeddedAssets, "assets/codex/native/SKILL.md")
	if err != nil {
		return CodexPackage{}, err
	}
	base, err := loadVariant("codex")
	if err != nil {
		return CodexPackage{}, err
	}
	artifacts, _, err := splitCodexAsset(base, "## Artifact Lifecycle", "## Engram Integration")
	if err != nil {
		return CodexPackage{}, err
	}
	pkg := CodexPackage{
		Universal: "OPSX is available through explicit invocation of `$sdd-orchestrator`. The OpenSpec CLI owns artifact state. Never guess artifact state or create openspec directories manually.\n",
		Files: map[string]string{
			"SKILL.md":             string(skill),
			"agents/openai.yaml":   "# managed-by: active-stack\npolicy:\n  allow_implicit_invocation: false\n",
			"modules/artifacts.md": artifacts,
		},
	}
	for _, name := range []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"} {
		content, err := fs.ReadFile(embeddedAssets, "assets/codex/native/modules/"+name+".md")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Files["modules/"+name+".md"] = string(content)
	}
	if active["governance"] {
		governance, err := loadFragment("governance.md")
		if err != nil {
			return CodexPackage{}, err
		}
		universal, coordinator, err := splitCodexAsset(governance, "## Agent Governance", "### How to Apply")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Universal += "\n" + universal
		pkg.Files["SKILL.md"] += "\n" + coordinator
	}
	if active["engram"] {
		engram, err := loadFragment("engram-protocol.md")
		if err != nil {
			return CodexPackage{}, err
		}
		header, wiring, err := splitCodexAsset(engram, "## Engram Persistent Memory", "### OPSX wiring")
		if err != nil {
			return CodexPackage{}, err
		}
		opsx, _, err := splitCodexAsset(wiring, "### OPSX wiring", "### PROACTIVE SAVE TRIGGERS")
		if err != nil {
			return CodexPackage{}, err
		}
		universal, lifecycle, err := splitCodexAsset(engram, "### PROACTIVE SAVE TRIGGERS", "### SESSION CLOSE PROTOCOL")
		if err != nil {
			return CodexPackage{}, err
		}
		coordinator, sync, err := splitCodexAsset(lifecycle, "### SESSION CLOSE PROTOCOL", "### GIT SYNC")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Universal += "\n" + header + "\n" + universal
		pkg.Files["SKILL.md"] += "\n" + coordinator
		pkg.Files["modules/engram-opsx.md"] = opsx + "\n" + sync
	}
	// opsx-selective-delegation (07): same rule and module text as Claude.
	rule, err := delegationRule(active["delegation"])
	if err != nil {
		return CodexPackage{}, err
	}
	pkg.Files["SKILL.md"] += "\n" + rule
	if active["delegation"] {
		delegation, err := delegationModule()
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Files["modules/delegation.md"] = delegation
	}
	if active["tdd"] {
		tdd, err := loadFragment("strict-tdd.md")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Files["modules/tdd.md"] = "Strict TDD Mode: enabled\n\n" + tdd
	}
	table, err := fs.ReadFile(embeddedAssets, "assets/codex/native/action-table.md")
	if err != nil {
		return CodexPackage{}, err
	}
	pkg.Files["SKILL.md"] += codexActionModules(active, string(table))
	return pkg, nil
}

// codexActionModules applies Codex-relative paths and existing toggle rules to
// the action table delivered by change 05. It does not compose Claude assets.
func codexActionModules(active map[string]bool, table string) string {
	tdd := "Strict TDD is disabled; stop apply if Strict TDD is required"
	delegation := "delegation disabled — verify stopped"
	delegate := "delegation disabled; do not delegate"
	engram := "Engram OPSX wiring disabled"
	if active["tdd"] {
		tdd = "`modules/tdd.md` when Strict TDD is on and a test runner exists"
	}
	if active["delegation"] {
		delegation = "`modules/delegation.md`"
		delegate = "`modules/delegation.md`"
	}
	if active["engram"] {
		engram = "`modules/engram-opsx.md`"
	}
	return fmt.Sprintf(table, tdd, delegation, delegate, engram)
}

// splitCodexAsset keeps the source fragments authoritative while assigning each
// section to its delivery layer. Missing boundaries fail composition rather than
// silently omitting mandatory instructions after an asset is edited.
func splitCodexAsset(content, start, boundary string) (string, string, error) {
	_, section, found := strings.Cut(content, start)
	if !found {
		return "", "", fmt.Errorf("codex package: missing asset section %q", start)
	}
	before, after, found := strings.Cut(section, boundary)
	if !found {
		return "", "", fmt.Errorf("codex package: missing asset boundary %q", boundary)
	}
	return strings.TrimSpace(start+before) + "\n", boundary + after, nil
}
