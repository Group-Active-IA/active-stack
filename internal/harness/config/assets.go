package config

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

// layerMarkerRe matches the layer annotations (<!-- as:<layer>/<name> -->) that
// partition an asset per claude-native-opsx. A marker alone on its line takes
// the newline with it, so removing every marker restores the original bytes.
var layerMarkerRe = regexp.MustCompile(`<!-- as:([a-z0-9/-]+) -->\n?`)

// stripLayerMarkers removes the layer annotations. Legacy (single block)
// composition goes through it, so its output is byte for byte the unannotated text.
func stripLayerMarkers(s string) string { return layerMarkerRe.ReplaceAllString(s, "") }

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

// sectionFragments maps a config harness ID to the asset fragment injected as
// its own idempotent section (active-stack:<id>) in each supported agent's
// instructions file. A harness absent from this map is treated as the
// sdd-orchestrator (composed from toggles per agent variant).
var sectionFragments = map[string]string{
	"design-mockups": "design-mockups.md",
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
	return stripLayerMarkers(string(data)), nil
}

// loadFragment reads a top-level asset fragment file by name (without the assets/ prefix).
func loadFragment(name string) (string, error) {
	data, err := fs.ReadFile(embeddedAssets, "assets/"+name)
	if err != nil {
		return "", fmt.Errorf("load fragment %q: %w", name, err)
	}
	return stripLayerMarkers(string(data)), nil
}

// layerFragment is one annotated piece of an asset.
type layerFragment struct {
	layer, name, text string
}

// loadLayerFragments reads an embedded asset (path relative to assets/) and
// splits it at its layer markers. Text before the first marker is rejected: every
// byte of an annotated asset belongs to a fragment.
func loadLayerFragments(path string) ([]layerFragment, error) {
	data, err := fs.ReadFile(embeddedAssets, "assets/"+path)
	if err != nil {
		return nil, fmt.Errorf("load layered asset %q: %w", path, err)
	}
	raw := string(data)
	locs := layerMarkerRe.FindAllStringSubmatchIndex(raw, -1)
	if len(locs) == 0 || locs[0][0] != 0 {
		return nil, fmt.Errorf("layered asset %q: text before the first layer marker", path)
	}
	var out []layerFragment
	for i, loc := range locs {
		end := len(raw)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		id := raw[loc[2]:loc[3]]
		layer, name, _ := strings.Cut(id, "/")
		out = append(out, layerFragment{layer: layer, name: name, text: raw[loc[1]:end]})
	}
	return out, nil
}
