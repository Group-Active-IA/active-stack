package permissions_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

type codexPermissionsAdapter struct{ path string }

func (a codexPermissionsAdapter) Agent() model.Agent         { return model.AgentCodex }
func (a codexPermissionsAdapter) SettingsPath(string) string { return a.path }

func TestInstallCodexPermissionTiers(t *testing.T) {
	tests := []struct {
		tier model.PermissionTier
		want []string
	}{
		{model.TierEstricto, []string{`approval_policy = "untrusted"`, `sandbox_mode = "read-only"`}},
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
