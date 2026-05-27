// Package catalog loads and validates the embedded master catalog of
// harnesses that the installer can install or configure.
package catalog

import (
	_ "embed"
	"fmt"

	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"gopkg.in/yaml.v3"
)

//go:embed harnesses.yaml
var rawCatalog []byte

// Catalog is the parsed, validated set of available harnesses.
type Catalog struct {
	Harnesses []model.Harness `yaml:"harnesses"`

	index map[string]model.Harness
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
			if h.ThirdParty {
				c.Harnesses[i].Source.Method = "npx"
			} else {
				c.Harnesses[i].Source.Method = "clone"
			}
		}
	}
	for _, h := range c.Harnesses {
		c.index[h.ID] = h
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
		switch h.Type {
		case model.HarnessSkill:
			if h.Source == nil || h.Source.Repo == "" {
				return fmt.Errorf("catalog: skill harness %q needs a source.repo", h.ID)
			}
			switch h.Source.Method {
			case "clone", "npx", "embed":
				// valid
			default:
				return fmt.Errorf("catalog: skill harness %q has unknown source.method %q (want clone|npx|embed)", h.ID, h.Source.Method)
			}
		case model.HarnessExternal:
			if h.External == nil || h.External.Method == "" {
				return fmt.Errorf("catalog: external harness %q needs an external.method", h.ID)
			}
		}
		for _, dep := range h.DependsOn {
			if _, ok := c.index[dep]; !ok {
				return fmt.Errorf("catalog: harness %q depends on unknown harness %q", h.ID, dep)
			}
		}
	}
	return nil
}

// ByID returns the harness with the given id.
func (c *Catalog) ByID(id string) (model.Harness, bool) {
	h, ok := c.index[id]
	return h, ok
}

// ForMode returns the harnesses that belong to the given install mode, in
// catalog order.
func (c *Catalog) ForMode(m model.InstallMode) []model.Harness {
	var out []model.Harness
	for _, h := range c.Harnesses {
		if h.InMode(m) {
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
