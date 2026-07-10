package model

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LocalizedText is a per-language text field (catalog description or
// long_description). Keys are ISO-ish language codes ("es", "en"); the
// catalog currently only requires "es" and "en" (D1, design.md).
type LocalizedText map[string]string

// Localized returns the value for lang when non-empty, otherwise falls back
// to "en", otherwise "es", otherwise "" (D1). It is safe on a nil or empty
// map — it never panics.
func (t LocalizedText) Localized(lang string) string {
	if v, ok := t[lang]; ok && v != "" {
		return v
	}
	if v, ok := t["en"]; ok && v != "" {
		return v
	}
	if v, ok := t["es"]; ok && v != "" {
		return v
	}
	return ""
}

// Validate returns an error unless BOTH "es" and "en" are present and
// non-empty. field and owner name the offending value in the error message,
// in the style: catalog: <owner> <field> missing language "<lang>" (D1).
func (t LocalizedText) Validate(field, owner string) error {
	for _, lang := range []string{"es", "en"} {
		if t[lang] == "" {
			return fmt.Errorf("catalog: %s %s missing language %q", owner, field, lang)
		}
	}
	return nil
}

// UnmarshalYAML accepts either a scalar node — assigned to the "es" key for
// parse-time backward compatibility during migration — or a mapping node
// decoded key-for-key into the map (D1). Accepting the scalar form does not
// bypass Validate: a scalar-sourced value populates only "es", so Validate
// still fails on the missing "en".
func (t *LocalizedText) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		*t = LocalizedText{"es": s}
		return nil
	case yaml.MappingNode:
		m := make(map[string]string, len(value.Content)/2)
		if err := value.Decode(&m); err != nil {
			return err
		}
		*t = LocalizedText(m)
		return nil
	default:
		return fmt.Errorf("localized text: unsupported YAML node kind %v", value.Kind)
	}
}
