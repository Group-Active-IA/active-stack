package permissions_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config/permissions"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// R-18 · riesgo de arrays y escalares en MergeJSONObjects.
//
// filemerge.MergeJSONObjects reemplaza los arrays del base por los del overlay
// y los escalares también. Sin tratamiento previo, instalar permisos pisaría los
// `allow`, `deny` y `ask` que el usuario ya tenga en ~/.claude/settings.json.
// Estas pruebas fijan el contrato: lo del usuario sobrevive; lo que el stack
// administra (piso de deny, allow del nivel, defaultMode) lo decide el nivel.

const userSettings = `{
  "model": "sonnet",
  "theme": "dark",
  "hooks": {
    "SessionStart": [
      {"hooks": [{"type": "command", "command": "echo user-hook"}]}
    ]
  },
  "permissions": {
    "defaultMode": "acceptEdits",
    "additionalDirectories": ["C:\\work\\extra"],
    "allow": ["Bash(npm test:*)", "mcp__mine__tool"],
    "deny": ["Bash(curl:*)", "Read(secrets/**)"],
    "ask": ["Bash(docker:*)"]
  }
}
`

func writeClaudeSettings(t *testing.T, home, content string) string {
	t.Helper()
	p := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func readSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("settings.json inválido: %v\n%s", err, raw)
	}
	return m
}

func strList(t *testing.T, v any) []string {
	t.Helper()
	arr, _ := v.([]any)
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		s, ok := e.(string)
		if !ok {
			t.Fatalf("entrada no string: %#v", e)
		}
		out = append(out, s)
	}
	return out
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func count(list []string, want string) int {
	n := 0
	for _, s := range list {
		if s == want {
			n++
		}
	}
	return n
}

func TestInstallClaudePreservesUserArrays(t *testing.T) {
	home := t.TempDir()
	path := writeClaudeSettings(t, home, userSettings)

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatalf("Install: %v", err)
	}

	perms := readSettings(t, path)["permissions"].(map[string]any)
	allow := strList(t, perms["allow"])
	deny := strList(t, perms["deny"])
	ask := strList(t, perms["ask"])

	for _, want := range []string{"Bash(npm test:*)", "mcp__mine__tool"} {
		if !contains(allow, want) {
			t.Errorf("allow del usuario perdido: falta %q (allow=%v)", want, allow)
		}
	}
	for _, want := range []string{"Bash(curl:*)", "Read(secrets/**)"} {
		if !contains(deny, want) {
			t.Errorf("deny del usuario perdido: falta %q (deny=%v)", want, deny)
		}
	}
	for _, want := range permissions.SecurityFloorDeny() {
		if !contains(deny, want) {
			t.Errorf("piso de deny ausente: falta %q", want)
		}
	}
	if !contains(allow, "Bash(go test:*)") {
		t.Errorf("allow del nivel balanceado ausente (allow=%v)", allow)
	}
	if !contains(ask, "Bash(docker:*)") {
		t.Errorf("ask del usuario perdido (ask=%v)", ask)
	}
}

func TestInstallClaudePreservesUserScalarsAndForeignKeys(t *testing.T) {
	home := t.TempDir()
	path := writeClaudeSettings(t, home, userSettings)

	if _, err := permissions.Install(home, []permissions.PermissionsAdapter{claudeAdapter()}, model.TierBalanceado); err != nil {
		t.Fatalf("Install: %v", err)
	}

	got := readSettings(t, path)
	if got["model"] != "sonnet" || got["theme"] != "dark" {
		t.Errorf("escalares ajenos modificados: model=%v theme=%v", got["model"], got["theme"])
	}
	if _, ok := got["hooks"].(map[string]any)["SessionStart"]; !ok {
		t.Errorf("bloque hooks ajeno perdido: %#v", got["hooks"])
	}
	perms := got["permissions"].(map[string]any)
	if dirs := strList(t, perms["additionalDirectories"]); !contains(dirs, `C:\work\extra`) {
		t.Errorf("additionalDirectories del usuario perdido: %v", dirs)
	}
	// defaultMode lo decide el nivel (decisión de diseño existente): se documenta el valor.
	if perms["defaultMode"] != "default" {
		t.Errorf("defaultMode = %v, el nivel balanceado debe fijarlo en \"default\"", perms["defaultMode"])
	}
}

func TestInstallClaudeArraysAreIdempotentWithoutDuplicates(t *testing.T) {
	home := t.TempDir()
	path := writeClaudeSettings(t, home, userSettings)
	adapters := []permissions.PermissionsAdapter{claudeAdapter()}

	if _, err := permissions.Install(home, adapters, model.TierBalanceado); err != nil {
		t.Fatalf("Install 1: %v", err)
	}
	first, _ := os.ReadFile(path)
	res, err := permissions.Install(home, adapters, model.TierBalanceado)
	if err != nil {
		t.Fatalf("Install 2: %v", err)
	}
	second, _ := os.ReadFile(path)

	if !bytes.Equal(first, second) {
		t.Errorf("la reinstalación cambió el archivo:\n--- 1\n%s\n--- 2\n%s", first, second)
	}
	if res.Changed {
		t.Errorf("la reinstalación idéntica no debe reportar Changed")
	}
	perms := readSettings(t, path)["permissions"].(map[string]any)
	for _, list := range [][]string{strList(t, perms["allow"]), strList(t, perms["deny"])} {
		for _, s := range list {
			if count(list, s) != 1 {
				t.Errorf("entrada duplicada %q en %v", s, list)
			}
		}
	}
}

func TestInstallClaudeTierSwitchDropsManagedAllowKeepsUser(t *testing.T) {
	home := t.TempDir()
	path := writeClaudeSettings(t, home, userSettings)
	adapters := []permissions.PermissionsAdapter{claudeAdapter()}

	if _, err := permissions.Install(home, adapters, model.TierBalanceado); err != nil {
		t.Fatalf("Install balanceado: %v", err)
	}
	if _, err := permissions.Install(home, adapters, model.TierEstricto); err != nil {
		t.Fatalf("Install estricto: %v", err)
	}

	allow := strList(t, readSettings(t, path)["permissions"].(map[string]any)["allow"])
	if contains(allow, "Bash(go test:*)") {
		t.Errorf("estricto no gana entradas allow: quedó la del nivel anterior (allow=%v)", allow)
	}
	if !contains(allow, "Bash(npm test:*)") {
		t.Errorf("allow del usuario perdido al cambiar de nivel (allow=%v)", allow)
	}
}
