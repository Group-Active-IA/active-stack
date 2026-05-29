package system

import (
	"context"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
)

// RequiredDependencies returns the deduplicated set of system dependencies
// required by the given harnesses. Each dependency's metadata (MinVersion,
// InstallHint) is sourced from the canonical defineDependencies list so there
// is a single source of truth.
//
// Mapping rules (from design D1):
//   - external method=="npm"   → node, npm
//   - skill   method=="npx"   → node, npm, npx
//   - skill   method=="clone" → git
//   - config / skill embed / external homebrew|download|mcp|go-install → (none)
func RequiredDependencies(harnesses []model.Harness, profile PlatformProfile) []Dependency {
	// Build a lookup map of canonical deps so we use their MinVersion/InstallHint.
	canonical := defineDependencies(profile)
	byName := make(map[string]Dependency, len(canonical))
	for _, d := range canonical {
		byName[d.Name] = d
	}

	// Collect the required names in order of first appearance.
	seen := make(map[string]bool)
	var ordered []string

	add := func(names ...string) {
		for _, name := range names {
			if !seen[name] {
				seen[name] = true
				ordered = append(ordered, name)
			}
		}
	}

	for _, h := range harnesses {
		switch h.Type {
		case model.HarnessExternal:
			if h.External != nil && h.External.Method == "npm" {
				add("node", "npm")
			}
			// homebrew | download | mcp | go-install → no per-harness runtime deps

		case model.HarnessSkill:
			if h.Source != nil {
				switch h.Source.Method {
				case "npx":
					add("node", "npm", "npx")
				case "clone":
					add("git")
				// embed → no runtime deps
				}
			}

		case model.HarnessConfig:
			// config harnesses never require external runtimes
		}
	}

	// Resolve against canonical metadata.
	result := make([]Dependency, 0, len(ordered))
	for _, name := range ordered {
		if dep, ok := byName[name]; ok {
			result = append(result, dep)
		}
	}
	return result
}

// DetectDepsFor is a thin exported wrapper over detectDeps that detects the
// given dependency slice. It is the seam used by the pre-flight gate so the
// gate can detect an explicit derived dep set rather than the full global list.
func DetectDepsFor(ctx context.Context, deps []Dependency) DependencyReport {
	return detectDeps(ctx, deps)
}
