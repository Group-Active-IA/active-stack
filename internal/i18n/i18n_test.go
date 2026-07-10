// Package i18n_test — RED tests for the internal/i18n core (Parse, T, Tf,
// LoadPreferred). i18n-engine-locales, tasks 1.1/1.3/1.5/1.7.
package i18n_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
)

// TestDefault asserts Default is LangEN (task 1.1).
func TestDefault(t *testing.T) {
	if i18n.Default != i18n.LangEN {
		t.Errorf("Default = %q, want %q", i18n.Default, i18n.LangEN)
	}
}

// TestParse covers valid en/es (trim + lower-case) and invalid values
// (task 1.1).
func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    i18n.Lang
		wantErr bool
	}{
		{name: "lowercase en", input: "en", want: i18n.LangEN},
		{name: "lowercase es", input: "es", want: i18n.LangES},
		{name: "uppercase EN", input: "EN", want: i18n.LangEN},
		{name: "mixed case Es", input: "Es", want: i18n.LangES},
		{name: "padded with spaces", input: "  en  ", want: i18n.LangEN},
		{name: "invalid value", input: "xx", wantErr: true},
		{name: "empty value", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := i18n.Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestParse_InvalidErrorNamesValue asserts the error message names the
// invalid value (task 1.1).
func TestParse_InvalidErrorNamesValue(t *testing.T) {
	_, err := i18n.Parse("xx")
	if err == nil {
		t.Fatal("Parse(\"xx\") expected error, got nil")
	}
	if !strings.Contains(err.Error(), "xx") {
		t.Errorf("Parse error = %q, want it to name the invalid value %q", err.Error(), "xx")
	}
}

// TestT_FallbackChain covers the T fallback chain: present key in the
// requested language, missing-in-language fallback to EN, missing-everywhere
// fallback to the key itself, and zero-value Lang resolving through Default
// (task 1.3).
func TestT_FallbackChain(t *testing.T) {
	t.Run("present key in es returns es value", func(t *testing.T) {
		got := i18n.T(i18n.LangES, "mode.lite.label")
		want := "Rápido"
		if got != want {
			t.Errorf("T(es, mode.lite.label) = %q, want %q", got, want)
		}
	})

	t.Run("present key in en returns en value", func(t *testing.T) {
		got := i18n.T(i18n.LangEN, "mode.lite.label")
		want := "Quick"
		if got != want {
			t.Errorf("T(en, mode.lite.label) = %q, want %q", got, want)
		}
	})

	t.Run("missing key everywhere returns key itself", func(t *testing.T) {
		got := i18n.T(i18n.LangEN, "this.key.does.not.exist")
		want := "this.key.does.not.exist"
		if got != want {
			t.Errorf("T(en, missing) = %q, want the key itself %q", got, want)
		}
	})

	t.Run("zero-value lang resolves through Default", func(t *testing.T) {
		var zero i18n.Lang
		got := i18n.T(zero, "mode.lite.label")
		want := i18n.T(i18n.Default, "mode.lite.label")
		if got != want {
			t.Errorf("T(zero-value, mode.lite.label) = %q, want Default value %q", got, want)
		}
	})

	t.Run("unknown language falls back to Default", func(t *testing.T) {
		got := i18n.T(i18n.Lang("fr"), "mode.lite.label")
		want := i18n.T(i18n.Default, "mode.lite.label")
		if got != want {
			t.Errorf("T(fr, mode.lite.label) = %q, want Default fallback %q", got, want)
		}
	})
}

// TestTf_Formats verifies Tf formats via the stored format string, exercised
// against component.fallback.generic_fmt = "Installs %s." (task 1.5).
func TestTf_Formats(t *testing.T) {
	got := i18n.Tf(i18n.LangEN, "component.fallback.generic_fmt", "Foo")
	want := "Installs Foo."
	if got != want {
		t.Errorf("Tf(en, generic_fmt, Foo) = %q, want %q", got, want)
	}
}

func TestTf_FormatsSpanish(t *testing.T) {
	got := i18n.Tf(i18n.LangES, "component.fallback.generic_fmt", "Foo")
	want := "Instala Foo."
	if got != want {
		t.Errorf("Tf(es, generic_fmt, Foo) = %q, want %q", got, want)
	}
}

// TestLoadPreferred covers reading <home>/.active-stack/config.json
// {"language":"es"} plus the Default fallback for missing/malformed/unknown
// cases (task 1.7).
func TestLoadPreferred(t *testing.T) {
	t.Run("reads language es from config.json", func(t *testing.T) {
		home := t.TempDir()
		writeConfig(t, home, `{"language":"es"}`)
		got := i18n.LoadPreferred(home)
		if got != i18n.LangES {
			t.Errorf("LoadPreferred() = %q, want %q", got, i18n.LangES)
		}
	})

	t.Run("reads language en from config.json", func(t *testing.T) {
		home := t.TempDir()
		writeConfig(t, home, `{"language":"en"}`)
		got := i18n.LoadPreferred(home)
		if got != i18n.LangEN {
			t.Errorf("LoadPreferred() = %q, want %q", got, i18n.LangEN)
		}
	})

	t.Run("missing file falls back to Default", func(t *testing.T) {
		home := t.TempDir()
		got := i18n.LoadPreferred(home)
		if got != i18n.Default {
			t.Errorf("LoadPreferred() = %q, want Default %q", got, i18n.Default)
		}
	})

	t.Run("malformed json falls back to Default", func(t *testing.T) {
		home := t.TempDir()
		writeConfig(t, home, `{not valid json`)
		got := i18n.LoadPreferred(home)
		if got != i18n.Default {
			t.Errorf("LoadPreferred() = %q, want Default %q", got, i18n.Default)
		}
	})

	t.Run("unknown value falls back to Default", func(t *testing.T) {
		home := t.TempDir()
		writeConfig(t, home, `{"language":"xx"}`)
		got := i18n.LoadPreferred(home)
		if got != i18n.Default {
			t.Errorf("LoadPreferred() = %q, want Default %q", got, i18n.Default)
		}
	})
}

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".active-stack")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config.json: %v", err)
	}
}
