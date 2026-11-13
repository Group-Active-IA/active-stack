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
	actions := "\n## Action modules\n\nRead `modules/artifacts.md` for propose, apply and archive.\n"
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
		actions += "Read `modules/engram-opsx.md` for OPSX memory wiring and project-memory Git sync.\n"
	}
	if active["delegation"] {
		delegation, err := fs.ReadFile(embeddedAssets, "assets/codex/native/delegation.md")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Files["modules/delegation.md"] = string(delegation)
		actions += "Read `modules/delegation.md` before delegating and for verify that requires an independent reviewer.\n"
	} else {
		actions += "Delegation is disabled. Stop any verification that requires an independent reviewer unless equivalent independent review is available.\n"
	}
	if active["tdd"] {
		tdd, err := loadFragment("strict-tdd.md")
		if err != nil {
			return CodexPackage{}, err
		}
		pkg.Files["modules/tdd.md"] = "Strict TDD Mode: enabled\n\n" + tdd
		actions += "Read `modules/tdd.md` for apply and include it in implementation assignments.\n"
	} else {
		actions += "TDD is disabled. Stop an apply action if its mandatory TDD requirement cannot be satisfied.\n"
	}
	pkg.Files["SKILL.md"] += actions
	return pkg, nil
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
