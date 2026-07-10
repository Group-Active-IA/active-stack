// Package i18n is the Go-side localization foundation for the Windows
// installer engine (change i18n-engine-locales, design D1/D2). It exposes a
// small Lang type, a Parse/T/Tf lookup surface backed by two flat string
// tables (table_en.go / table_es.go), and LoadPreferred for the future TUI.
//
// This package is deliberately dependency-free (no globals besides the
// read-only tables map) so callers thread Lang explicitly through the call
// graph rather than relying on ambient state.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Lang identifies a supported UI language.
type Lang string

const (
	// LangEN is English.
	LangEN Lang = "en"
	// LangES is Spanish.
	LangES Lang = "es"
)

// Default is the language used when no explicit choice is available (zero
// value Lang, unparseable LoadPreferred input, etc).
const Default = LangEN

// tables maps each supported language to its flat, dot-namespaced string
// table (design D1). Populated by table_en.go / table_es.go via init().
var tables = map[Lang]map[string]string{}

// registerTable is called by table_en.go / table_es.go to register their
// table under a language key.
func registerTable(lang Lang, table map[string]string) {
	tables[lang] = table
}

// Parse accepts "en"/"es" (trimmed, case-insensitive). Any other value
// (including empty string) returns a non-nil error naming the value.
func Parse(s string) (Lang, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	switch Lang(normalized) {
	case LangEN:
		return LangEN, nil
	case LangES:
		return LangES, nil
	default:
		return "", fmt.Errorf("invalid language %q: must be %q or %q", s, LangEN, LangES)
	}
}

// T looks up key in the table for lang. A zero-value or unknown lang
// resolves through Default. If the key is missing in the resolved language,
// it falls back to Default's table. If still missing, T returns the key
// itself so a stray key is visible, never a blank string.
func T(lang Lang, key string) string {
	if table, ok := tables[lang]; ok {
		if v, ok := table[key]; ok {
			return v
		}
	}
	if table, ok := tables[Default]; ok {
		if v, ok := table[key]; ok {
			return v
		}
	}
	return key
}

// Tf formats the value looked up via T as a fmt.Sprintf format string.
func Tf(lang Lang, key string, args ...any) string {
	return fmt.Sprintf(T(lang, key), args...)
}

// preferredConfig is the shape of <homeDir>/.active-stack/config.json read
// by LoadPreferred. Only the "language" key is relevant here.
type preferredConfig struct {
	Language string `json:"language"`
}

// LoadPreferred reads <homeDir>/.active-stack/config.json and returns the
// parsed "language" value, or Default on any error (missing file, malformed
// JSON, unknown value). It is provided for the future TUI; the engine itself
// does not call it — callers pass --lang explicitly.
func LoadPreferred(homeDir string) Lang {
	path := filepath.Join(homeDir, ".active-stack", "config.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Default
	}
	var cfg preferredConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Default
	}
	lang, err := Parse(cfg.Language)
	if err != nil {
		return Default
	}
	return lang
}
