package config

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed assets
var embeddedAssets embed.FS

// knownVariants is the set of agent variants that have their own asset directory.
// Any variant NOT in this set falls back to "generic".
var knownVariants = map[string]bool{
	"claude":      true,
	"opencode":    true,
	"codex":       true,
	"gemini":      true,
	"cursor":      true,
	"windsurf":    true,
	"antigravity": true,
	"generic":     true,
}

// loadVariant reads the sdd-orchestrator.md asset for the given variant key.
// If the variant is not known, it falls back to "generic".
func loadVariant(variantKey string) (string, error) {
	key := variantKey
	if !knownVariants[key] {
		key = "generic"
	}
	data, err := fs.ReadFile(embeddedAssets, "assets/"+key+"/sdd-orchestrator.md")
	if err != nil {
		return "", fmt.Errorf("load variant %q (resolved %q): %w", variantKey, key, err)
	}
	return string(data), nil
}

// loadFragment reads a top-level asset fragment file by name (without the assets/ prefix).
func loadFragment(name string) (string, error) {
	data, err := fs.ReadFile(embeddedAssets, "assets/"+name)
	if err != nil {
		return "", fmt.Errorf("load fragment %q: %w", name, err)
	}
	return string(data), nil
}
