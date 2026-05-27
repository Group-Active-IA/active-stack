package config

import (
	"fmt"
	"strings"

	"github.com/JuanCruzRobledo/jr-stack/internal/filemerge"
)

// knownToggles is the catalogue of recognised toggle names.
// Any toggle not in this set causes Compose to return an error.
var knownToggles = map[string]bool{
	"delegation":    true,
	"model-routing": true,
	"engram":        true,
	"tdd":           true,
	"governance":    true,
}

// Compose assembles the sdd-orchestrator block for the given variant and
// active toggles. It is a PURE function: same inputs always produce the same
// output. No I/O is performed; all assets are read from the embedded FS.
//
// Assembly order (deterministic, independent of toggle slice order):
//
//  1. base (always): assets/<variant>/sdd-orchestrator.md
//  2. Subtractive: if "delegation" NOT active → remove <!-- jr-stack:sdd-delegation --> section
//  3. Subtractive: if "model-routing" NOT active → remove <!-- jr-stack:sdd-model-assignments --> section
//  4. Additive: if "governance" active → append assets/governance.md
//  5. Additive: if "engram" active → append assets/engram-protocol.md
//  6. Additive: if "tdd" active → append inline TDD flag + assets/strict-tdd.md
//
// Returns an error if any toggle is not in the known catalogue.
func Compose(toggles []string, variantKey string) (string, error) {
	// Validate all toggles first.
	for _, t := range toggles {
		if !knownToggles[t] {
			return "", fmt.Errorf("config compose: unknown toggle %q (known: delegation, model-routing, engram, tdd, governance)", t)
		}
	}

	// Build an active toggle set for O(1) lookup.
	active := make(map[string]bool, len(toggles))
	for _, t := range toggles {
		active[t] = true
	}

	// Step 1: load base variant.
	base, err := loadVariant(variantKey)
	if err != nil {
		return "", fmt.Errorf("config compose: %w", err)
	}

	result := base

	// Step 2: subtractive — delegation.
	if !active["delegation"] {
		result = filemerge.InjectMarkdownSection(result, "sdd-delegation", "")
	}

	// Step 3: subtractive — model-routing.
	if !active["model-routing"] {
		result = filemerge.InjectMarkdownSection(result, "sdd-model-assignments", "")
	}

	// Steps 4-6: additive fragments (always appended in fixed order).
	var additives []string

	// Step 4: governance.
	if active["governance"] {
		gov, err := loadFragment("governance.md")
		if err != nil {
			return "", fmt.Errorf("config compose: %w", err)
		}
		additives = append(additives, gov)
	}

	// Step 5: engram.
	if active["engram"] {
		eng, err := loadFragment("engram-protocol.md")
		if err != nil {
			return "", fmt.Errorf("config compose: %w", err)
		}
		additives = append(additives, eng)
	}

	// Step 6: tdd flag + module.
	if active["tdd"] {
		tddModule, err := loadFragment("strict-tdd.md")
		if err != nil {
			return "", fmt.Errorf("config compose: %w", err)
		}
		tddBlock := "Strict TDD Mode: enabled\n\n" + tddModule
		additives = append(additives, tddBlock)
	}

	if len(additives) > 0 {
		combined := strings.Join(additives, "\n\n")
		if !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		result += "\n" + combined
	}

	return result, nil
}
