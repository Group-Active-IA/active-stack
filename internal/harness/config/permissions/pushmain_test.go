package permissions_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// R-18 · [EQ] Fase 6: deny de push a main y lecturas de gh.
// Reglas medidas en VE-5 (docs/equipo/fichas/R-18.md): las formas con `:*` cubren
// literales; para `git -C` y `git -c` solo sirve el comodín con espacio, y las
// dos reglas exactas evitan bloquear ramas como main-fix o maintenance.

var allTiers = []model.PermissionTier{model.TierEstricto, model.TierBalanceado, model.TierBypass}

func claudePerms(t *testing.T, home string) map[string]any {
	t.Helper()
	return readSettings(t, filepath.Join(home, ".claude", "settings.json"))["permissions"].(map[string]any)
}

// T-F6-1 · el deny aparece en todos los niveles de Claude.
func TestInstallClaudePushMainDenyInAllTiers(t *testing.T) {
	for _, tier := range allTiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, tier); err != nil {
				t.Fatalf("Install: %v", err)
			}
			deny := strList(t, claudePerms(t, home)["deny"])
			for _, want := range permissions.PushMainDeny() {
				if !contains(deny, want) {
					t.Errorf("nivel %s: falta el deny %q", tier, want)
				}
			}
		})
	}
}

// T-F6-2 · el piso de denegación existente queda intacto en todos los niveles.
func TestInstallClaudeKeepsExistingFloorWithPushDeny(t *testing.T) {
	for _, tier := range allTiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, tier); err != nil {
				t.Fatalf("Install: %v", err)
			}
			deny := strList(t, claudePerms(t, home)["deny"])
			for _, want := range permissions.SecurityFloorDeny() {
				if !contains(deny, want) {
					t.Errorf("nivel %s: piso perdido %q", tier, want)
				}
			}
		})
	}
	// SecurityFloorDeny no debe absorber las reglas nuevas (fuente única del piso).
	for _, r := range permissions.SecurityFloorDeny() {
		if strings.Contains(r, "push origin main") || strings.HasPrefix(r, "PowerShell(") {
			t.Errorf("el piso no debe contener reglas de R-18: %q", r)
		}
	}
}

// Contenido exacto medido en VE-5; sin las reglas con `main*` que bloquean main-fix.
func TestPushMainDenyRuleSet(t *testing.T) {
	rules := permissions.PushMainDeny()
	must := []string{
		"Bash(git push origin main:*)",
		"Bash(git push origin HEAD:main:*)",
		"Bash(git push -u origin main:*)",
		"Bash(git push --set-upstream origin main:*)",
		"Bash(git push origin refs/heads/main:*)",
		"Bash(git push origin HEAD:refs/heads/main:*)",
		"Bash(git push origin +main:*)",
		"Bash(git push origin :main:*)",
		"Bash(git push --delete origin main:*)",
		"Bash(git push)",
		"Bash(git * push origin main)",
		"Bash(git * push origin main *)",
		"PowerShell(git push origin main)",
		"PowerShell(git push origin main *)",
		"PowerShell(git push)",
		"PowerShell(git * push origin main)",
		"PowerShell(git * push origin main *)",
	}
	for _, m := range must {
		if !contains(rules, m) {
			t.Errorf("falta la regla %q", m)
		}
	}
	for _, r := range rules {
		if strings.HasSuffix(r, "main*)") || strings.HasSuffix(r, "main*") {
			t.Errorf("regla demasiado amplia, bloquea main-fix/maintenance: %q", r)
		}
		if r == "Bash(git push:*)" || r == "PowerShell(git push *)" {
			t.Errorf("regla demasiado amplia, bloquea el push de ramas de ficha: %q", r)
		}
	}
	seen := map[string]bool{}
	for _, r := range rules {
		if seen[r] {
			t.Errorf("regla duplicada: %q", r)
		}
		seen[r] = true
	}
}

// PushMainDeny devuelve una copia: mutarla no altera la fuente.
func TestPushMainDenyReturnsCopy(t *testing.T) {
	a := permissions.PushMainDeny()
	a[0] = "mutado"
	if permissions.PushMainDeny()[0] == "mutado" {
		t.Fatal("PushMainDeny expone el slice interno")
	}
}

// T-F6-3 · lecturas de gh solo en balanceado y sin argumentos libres de escritura.
func TestGhReadAllowOnlyInBalanceado(t *testing.T) {
	reads := []string{"Bash(gh pr list:*)", "Bash(gh pr diff:*)", "Bash(gh api user)"}
	for _, tier := range allTiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, tier); err != nil {
				t.Fatalf("Install: %v", err)
			}
			allow := strList(t, claudePerms(t, home)["allow"])
			for _, r := range reads {
				if got := contains(allow, r); got != (tier == model.TierBalanceado) {
					t.Errorf("nivel %s: %q presente=%v", tier, r, got)
				}
			}
			for _, a := range allow {
				if strings.HasPrefix(a, "Bash(gh ") && (strings.Contains(a, "api:*") || strings.Contains(a, " --method") || strings.Contains(a, " -X")) {
					t.Errorf("allow de gh con argumentos libres de escritura: %q", a)
				}
			}
		})
	}
}

// El deny de push no queda por error en el allow, ni al revés.
func TestPushMainDenyNeverInAllow(t *testing.T) {
	home := t.TempDir()
	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatal(err)
	}
	allow := strList(t, claudePerms(t, home)["allow"])
	for _, r := range permissions.PushMainDeny() {
		if contains(allow, r) {
			t.Errorf("regla de deny presente en allow: %q", r)
		}
	}
}

// T-F6-4 · Codex: no se escribe deny (queda (TBD)) y no vuelve "untrusted".
func TestCodexInstallNoPushDenyAndNoUntrusted(t *testing.T) {
	for _, tier := range allTiers {
		t.Run(string(tier), func(t *testing.T) {
			home := t.TempDir()
			cfg := filepath.Join(home, ".codex", "config.toml")
			if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(cfg, []byte("model = \"x\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			a := stubAdapter{agent: model.AgentCodex, settingsDir: []string{".codex", "config.toml"}}
			if _, err := permissions.Install(home, []permissions.PermissionsAdapter{a}, tier); err != nil {
				t.Fatalf("Install: %v", err)
			}
			raw, _ := os.ReadFile(cfg)
			if strings.Contains(string(raw), "untrusted") {
				t.Errorf("config.toml reintrodujo untrusted:\n%s", raw)
			}
			if strings.Contains(string(raw), "git push") {
				t.Errorf("no se implementa deny de Codex (TBD):\n%s", raw)
			}
			if _, err := os.Stat(filepath.Join(home, ".codex", "rules")); err == nil {
				t.Errorf("no debe crearse ~/.codex/rules (Codex TBD)")
			}
		})
	}
}

// DA-08 · OpenCode no recibe el deny de push a main.
func TestOpenCodeDoesNotGetPushMainDeny(t *testing.T) {
	home := t.TempDir()
	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{openCodeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
	if strings.Contains(string(raw), "push origin main") {
		t.Errorf("DA-08: OpenCode queda fuera del deny de push a main:\n%s", raw)
	}
}

// Coexistencia con los hooks de R-16 en el mismo settings.json, en ambos órdenes.
func TestCoexistenceWithR16Hooks(t *testing.T) {
	restoreSnap := hooks.SetSnapshotterCreate(func(string, []string) error { return nil })
	defer restoreSnap()
	restorePath := hooks.SetLookPath(func(string) (string, error) { return "active-stack", nil })
	defer restorePath()

	pa := []permissions.PermissionsAdapter{claudeAdapter()}
	ha := []hooks.Adapter{claudeAdapter().(hooks.Adapter)}

	check := func(t *testing.T, path string) {
		t.Helper()
		got := readSettings(t, path)
		perms := got["permissions"].(map[string]any)
		deny := strList(t, perms["deny"])
		allow := strList(t, perms["allow"])
		for _, w := range permissions.PushMainDeny() {
			if !contains(deny, w) {
				t.Errorf("deny perdido tras coexistir con hooks: %q", w)
			}
		}
		if !contains(deny, "Bash(curl:*)") || !contains(allow, "Bash(npm test:*)") {
			t.Errorf("reglas del usuario perdidas: allow=%v deny=%v", allow, deny)
		}
		h, _ := got["hooks"].(map[string]any)
		if _, ok := h["SessionStart"]; !ok {
			t.Fatalf("SessionStart perdido: %#v", got["hooks"])
		}
		raw, _ := json.Marshal(h)
		if !strings.Contains(string(raw), "echo user-hook") {
			t.Errorf("hook del usuario perdido: %s", raw)
		}
		st, err := hooks.Status(filepath.Dir(filepath.Dir(path)), ha)
		if err != nil || len(st) == 0 || !st[0].Installed {
			t.Errorf("hooks.Status no ve los hooks instalados: %+v err=%v", st, err)
		}
	}

	t.Run("permisos y luego hooks", func(t *testing.T) {
		home := t.TempDir()
		path := writeClaudeSettings(t, home, userSettings)
		if _, err := permissions.Install(home, pa, model.TierBalanceado); err != nil {
			t.Fatal(err)
		}
		if _, err := hooks.Install(home, ha); err != nil {
			t.Fatal(err)
		}
		check(t, path)
	})
	t.Run("hooks y luego permisos", func(t *testing.T) {
		home := t.TempDir()
		path := writeClaudeSettings(t, home, userSettings)
		if _, err := hooks.Install(home, ha); err != nil {
			t.Fatal(err)
		}
		if _, err := permissions.Install(home, pa, model.TierBalanceado); err != nil {
			t.Fatal(err)
		}
		check(t, path)
	})
	t.Run("quitar los hooks deja los permisos", func(t *testing.T) {
		home := t.TempDir()
		path := writeClaudeSettings(t, home, userSettings)
		if _, err := permissions.Install(home, pa, model.TierBalanceado); err != nil {
			t.Fatal(err)
		}
		before := readSettings(t, path)
		if _, err := hooks.Install(home, ha); err != nil {
			t.Fatal(err)
		}
		if _, err := hooks.Remove(home, ha); err != nil {
			t.Fatal(err)
		}
		if after := readSettings(t, path); !reflect.DeepEqual(before, after) {
			t.Errorf("hooks.Remove no devolvió la configuración de permisos:\nantes:  %#v\ndespués: %#v", before, after)
		}
	})
}
