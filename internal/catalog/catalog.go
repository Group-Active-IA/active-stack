// Package catalog loads and validates the embedded master catalog of
// harnesses that the installer can install or configure.
package catalog

import (
	_ "embed"
	"fmt"

	"github.com/Group-Active-IA/active-stack/internal/model"
	"gopkg.in/yaml.v3"
)

//go:embed harnesses.yaml
var rawCatalog []byte

// Catalog is the parsed, validated set of available harnesses and starters.
type Catalog struct {
	Harnesses []model.Harness  `yaml:"harnesses"`
	Starters  []model.Starter  `yaml:"starters"`

	index        map[string]model.Harness
	starterIndex map[string]model.Starter
}

// Load parses the embedded catalog and validates it. It is the single entry
// point: a malformed catalog is a build/release error, so this fails loudly.
func Load() (*Catalog, error) {
	return parse(rawCatalog)
}

func parse(data []byte) (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("catalog: parse: %w", err)
	}
	c.index = make(map[string]model.Harness, len(c.Harnesses))
	// Infer Source.Method for skill harnesses that omit it.
	for i, h := range c.Harnesses {
		if h.Type == model.HarnessSkill && h.Source != nil && h.Source.Method == "" {
			// Both first-party and third-party skills clone (npx support removed).
			c.Harnesses[i].Source.Method = "clone"
		}
	}
	for _, h := range c.Harnesses {
		c.index[h.ID] = h
	}
	// Build starter index.
	c.starterIndex = make(map[string]model.Starter, len(c.Starters))
	for _, s := range c.Starters {
		c.starterIndex[s.ID] = s
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Catalog) validate() error {
	seen := make(map[string]bool, len(c.Harnesses))
	for _, h := range c.Harnesses {
		switch {
		case h.ID == "":
			return fmt.Errorf("catalog: harness with empty id")
		case seen[h.ID]:
			return fmt.Errorf("catalog: duplicate harness id %q", h.ID)
		case !h.Type.IsValid():
			return fmt.Errorf("catalog: harness %q has invalid type %q", h.ID, h.Type)
		case len(h.InstallModes) == 0:
			return fmt.Errorf("catalog: harness %q has no install_modes", h.ID)
		}
		seen[h.ID] = true

		for _, m := range h.InstallModes {
			if !m.IsValid() {
				return fmt.Errorf("catalog: harness %q has invalid mode %q", h.ID, m)
			}
		}
		// Rule 2 (C-32): a starter-only harness must not list lite in install_modes.
		// A starter-only skill cannot be part of the global Lite minimum.
		if h.IsStarterOnly() {
			for _, m := range h.InstallModes {
				if m == model.ModeLite {
					return fmt.Errorf("catalog: harness %q is starter-only but lists install_mode %q — starter-only skills cannot be part of the Lite minimum", h.ID, model.ModeLite)
				}
			}
		}
		switch h.Type {
		case model.HarnessSkill:
			if h.Source == nil || h.Source.Repo == "" {
				return fmt.Errorf("catalog: skill harness %q needs a source.repo", h.ID)
			}
			switch h.Source.Method {
			case "clone", "embed":
				// valid
			default:
				return fmt.Errorf("catalog: skill harness %q has unknown source.method %q (want clone|embed)", h.ID, h.Source.Method)
			}
		case model.HarnessExternal:
			if h.External == nil || h.External.Method == "" {
				return fmt.Errorf("catalog: external harness %q needs an external.method", h.ID)
			}
			// Validate the optional MCP stdio spec when present. A malformed
			// catalog is a build/release error — fail loudly.
			if h.External.MCP != nil {
				if err := h.External.MCP.Validate(); err != nil {
					return fmt.Errorf("catalog: external harness %q mcp: %w", h.ID, err)
				}
			}
		}
		for _, dep := range h.DependsOn {
			if _, ok := c.index[dep]; !ok {
				return fmt.Errorf("catalog: harness %q depends on unknown harness %q", h.ID, dep)
			}
		}
	}
	if err := c.validateStarters(); err != nil {
		return err
	}
	// Rule 1 (C-32): every starter-only harness must be referenced by at least
	// one starter. Run AFTER validateStarters() so the includes graph is proven
	// free of cycles and broken references (safe to call ResolveStarter).
	return c.validateStarterOnlyReferences()
}

// validateStarterOnlyReferences checks that every harness with Scope==ScopeStarterOnly
// is reachable from at least one starter (directly or via includes). An orphaned
// starter-only harness is a catalog error: it can never be installed.
func (c *Catalog) validateStarterOnlyReferences() error {
	// Build the union set of all harness ids referenced across all starters.
	referenced := make(map[string]bool)
	for _, s := range c.Starters {
		harnesses, err := c.ResolveStarter(s.ID)
		if err != nil {
			// validateStarters() already checked this; should not happen.
			return fmt.Errorf("catalog: unexpected error resolving starter %q: %w", s.ID, err)
		}
		for _, h := range harnesses {
			referenced[h.ID] = true
		}
	}
	for _, h := range c.Harnesses {
		if h.IsStarterOnly() && !referenced[h.ID] {
			return fmt.Errorf("catalog: harness %q has scope starter-only but is not referenced by any starter", h.ID)
		}
	}
	return nil
}

// validateStarters checks all starters in the catalog:
//   - each starter's own fields are valid (via Starter.Validate())
//   - no empty ids
//   - no duplicate ids
//   - every Harnesses[i] references an existing harness id
//   - every Includes[i] references an existing starter id
//   - no cycles in the includes graph (DFS tri-state: unvisited / in-stack / done)
//
// A malformed starters section is a build/release error, same as harnesses.
func (c *Catalog) validateStarters() error {
	seen := make(map[string]bool, len(c.Starters))
	for _, s := range c.Starters {
		// Field-level validation.
		if err := s.Validate(); err != nil {
			return fmt.Errorf("catalog: starter validation: %w", err)
		}
		// Duplicate id check.
		if seen[s.ID] {
			return fmt.Errorf("catalog: duplicate starter id %q", s.ID)
		}
		seen[s.ID] = true

		// Harness references must exist.
		for _, hid := range s.Harnesses {
			if _, ok := c.index[hid]; !ok {
				return fmt.Errorf("catalog: starter %q references unknown harness %q", s.ID, hid)
			}
		}
		// Include references must exist.
		for _, inc := range s.Includes {
			if _, ok := c.starterIndex[inc]; !ok {
				return fmt.Errorf("catalog: starter %q includes unknown starter %q", s.ID, inc)
			}
		}
	}
	// Cycle detection — DFS with tri-state marking over the includes graph.
	// States: 0 = unvisited, 1 = in-stack (being explored), 2 = done.
	state := make(map[string]int, len(c.Starters))
	var dfs func(id string) error
	dfs = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("catalog: cycle detected in starter includes involving %q", id)
		case 2:
			return nil
		}
		state[id] = 1 // mark as in-stack
		s := c.starterIndex[id]
		for _, inc := range s.Includes {
			if err := dfs(inc); err != nil {
				return err
			}
		}
		state[id] = 2 // mark as done
		return nil
	}
	for _, s := range c.Starters {
		if err := dfs(s.ID); err != nil {
			return err
		}
	}
	return nil
}

// ByID returns the harness with the given id.
func (c *Catalog) ByID(id string) (model.Harness, bool) {
	h, ok := c.index[id]
	return h, ok
}

// AllHarnesses returns every harness in the catalog in catalog order, applying
// no scope filter and no install-mode filter. It is the resolution universe for
// the planner's dependency-resolution index: the planner must be able to look up
// any harness that SelectHarnesses may have placed into the selected set, including
// starter-only harnesses that ForMode purposely excludes.
//
// This is intentionally distinct from ForMode: ForMode is a selection accessor
// (which harnesses are eligible for a given mode?); AllHarnesses is a resolution
// accessor (what is the complete dependency graph?).
func (c *Catalog) AllHarnesses() []model.Harness {
	return c.Harnesses
}

// ForMode returns the harnesses that belong to the given install mode, in
// catalog order. Starter-only harnesses (Scope==ScopeStarterOnly) are excluded
// from every mode: their only installation path is `active-stack starter add`.
// InMode semantics are unchanged — the scope filter is applied here, not in InMode.
func (c *Catalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range c.Harnesses {
		if !h.IsStarterOnly() && h.InMode(m) {
			out = append(out, h)
		}
	}
	return out
}

// ForAgent returns the harnesses that apply to the given agent, in catalog
// order.
func (c *Catalog) ForAgent(a model.Agent) []model.Harness {
	var out []model.Harness
	for _, h := range c.Harnesses {
		if h.SupportsAgent(a) {
			out = append(out, h)
		}
	}
	return out
}

// ResolveStarterMCPs returns the TOTAL set of project-scope MCP entries for the
// starter with the given id, aggregating across all transitively included
// starters. The result is deduplicated by MCP.Name; on a collision the root
// starter's entry takes precedence (explicit override). Order is stable:
// included starters are visited depth-first (pre-order) so includes come first,
// then the current starter's own MCPs override any that share a name.
//
// This mirrors the traversal order of ResolveStarter (which aggregates harnesses)
// so that the two methods can be understood and maintained together.
//
// Returns an error if the id is unknown. Because Load() already validated the
// includes graph (no cycles, no broken references), traversal always terminates.
func (c *Catalog) ResolveStarterMCPs(id string) ([]model.MCP, error) {
	if _, ok := c.starterIndex[id]; !ok {
		return nil, fmt.Errorf("catalog: starter %q not found", id)
	}

	// Collect all MCPs in depth-first pre-order: includes first, then the
	// current starter's own MCPs. This means the root starter's MCPs appear
	// last in the list, which lets the dedup step below make them win.
	var ordered []model.MCP

	var collect func(sid string)
	collect = func(sid string) {
		s := c.starterIndex[sid]
		for _, inc := range s.Includes {
			collect(inc)
		}
		ordered = append(ordered, s.MCPs...)
	}
	collect(id)

	// Deduplicate by name: scan in reverse (root's MCPs are at the end, so
	// they are encountered first in a reverse scan) and keep the first
	// occurrence of each name encountered. Then reverse the kept slice to
	// restore stable forward order.
	seenName := make(map[string]bool, len(ordered))
	kept := make([]model.MCP, 0, len(ordered))
	for i := len(ordered) - 1; i >= 0; i-- {
		if !seenName[ordered[i].Name] {
			seenName[ordered[i].Name] = true
			kept = append(kept, ordered[i])
		}
	}
	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}
	return kept, nil
}

// AllStarters returns all starters in the catalog in catalog order.
// Used by the "starter add" handler to build the available-starters error message.
func (c *Catalog) AllStarters() []model.Starter {
	return c.Starters
}

// StarterByID returns the starter with the given id and a found flag.
func (c *Catalog) StarterByID(id string) (model.Starter, bool) {
	s, ok := c.starterIndex[id]
	return s, ok
}

// ResolveStarter expands the starter with the given id into its TOTAL set of
// harnesses by resolving includes recursively. The result is deduplicated
// (each harness appears at most once) and preserves first-appearance order for
// deterministic output. Returns an error if the id is unknown.
//
// Because validateStarters() already rejected cycles and broken references
// during Load(), this traversal is guaranteed to terminate and to find every
// referenced harness in the index.
func (c *Catalog) ResolveStarter(id string) ([]model.Harness, error) {
	if _, ok := c.starterIndex[id]; !ok {
		return nil, fmt.Errorf("catalog: starter %q not found", id)
	}
	var out []model.Harness
	seen := make(map[string]bool)

	var resolve func(sid string)
	resolve = func(sid string) {
		s := c.starterIndex[sid]
		// First recurse into included starters so that their harnesses appear
		// before the current starter's own harnesses (depth-first, pre-order).
		// This preserves a stable, deterministic ordering.
		for _, inc := range s.Includes {
			resolve(inc)
		}
		for _, hid := range s.Harnesses {
			if !seen[hid] {
				seen[hid] = true
				if h, ok := c.index[hid]; ok {
					out = append(out, h)
				}
			}
		}
	}

	resolve(id)
	return out, nil
}
