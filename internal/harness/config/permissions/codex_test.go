package permissions_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestCodexStrictPreservesConfigOnFailure(t *testing.T) {
	for _, test := range []struct {
		name        string
		initial     string
		backupFails bool
	}{
		{"invalid TOML", "approval_policy = [\n", false},
		{"backup failure", "approval_policy = \"untrusted\"\n", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, "config.toml")
			if err := os.WriteFile(path, []byte(test.initial), 0o600); err != nil {
				t.Fatal(err)
			}
			if test.backupFails {
				t.Cleanup(permissions.SetSnapshotterCreate(func(string, []string) error {
					return errors.New("backup failed")
				}))
			}
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{codexPermissionsAdapter{path}}, model.TierEstricto); err == nil {
				t.Fatal("expected installation to fail")
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(raw) != test.initial {
				t.Fatalf("configuration changed: %q", raw)
			}
		})
	}
}

type codexPermissionsAdapter struct{ path string }

func (a codexPermissionsAdapter) Agent() model.Agent         { return model.AgentCodex }
func (a codexPermissionsAdapter) SettingsPath(string) string { return a.path }

func TestInstallCodexPermissionTiers(t *testing.T) {
	tests := []struct {
		tier model.PermissionTier
		want []string
	}{
		{model.TierEstricto, []string{`approval_policy = "on-request"`, `sandbox_mode = "read-only"`}},
		{model.TierBalanceado, []string{`approval_policy = "on-request"`, `sandbox_mode = "workspace-write"`, `network_access = false`}},
		{model.TierBypass, []string{`approval_policy = "never"`, `sandbox_mode = "danger-full-access"`}},
	}
	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".codex", "config.toml")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("model = \"keep-me\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := permissions.Install(t.TempDir(), []permissions.PermissionsAdapter{codexPermissionsAdapter{path}}, tt.tier); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			got := string(raw)
			if !strings.Contains(got, `model = "keep-me"`) {
				t.Fatalf("unrelated config lost:\n%s", got)
			}
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q:\n%s", want, got)
				}
			}
		})
	}
}

func TestReinstallCodexStrictMigratesLegacyPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	initial := "approval_policy = \"untrusted\"\n[profiles.personal]\nmodel = \"keep-me\"\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	adapters := []permissions.PermissionsAdapter{codexPermissionsAdapter{path}}
	var previous string
	for range 2 {
		if _, err := permissions.Install(t.TempDir(), adapters, model.TierEstricto); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got := string(raw)
		if strings.Contains(got, "untrusted") ||
			!strings.Contains(got, `approval_policy = "on-request"`) ||
			!strings.Contains(got, "[profiles.personal]\nmodel = \"keep-me\"") {
			t.Fatalf("unexpected config after migration:\n%s", got)
		}
		if previous != "" && got != previous {
			t.Fatalf("reinstall changed config:\n%s", got)
		}
		previous = got
	}
}
