package headless_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/harness/config"
	"github.com/Group-Active-IA/active-stack/internal/model"
	"github.com/Group-Active-IA/active-stack/internal/uninstall"
)

// layeredBoth satisfies config.AgentAdapter and uninstall.AgentAdapter.
type layeredBoth struct{ home string }

func (a layeredBoth) Agent() model.Agent { return model.AgentClaude }
func (a layeredBoth) InstructionsPath(string) string {
	return filepath.Join(a.home, ".claude", "CLAUDE.md")
}
func (a layeredBoth) SkillsDir(string) string { return filepath.Join(a.home, ".claude", "skills") }
func (a layeredBoth) SettingsPath(string) string {
	return filepath.Join(a.home, ".claude", "settings.json")
}
func (a layeredBoth) CommandsDir(string) string            { return "" }
func (a layeredBoth) VariantKey() string                   { return "claude" }
func (a layeredBoth) ConfigDelivery() model.ConfigDelivery { return model.ConfigDeliveryManagedSkill }

// 7.2 (M): --yes alone never overwrites an edited CLAUDE.md; --overwrite-edited does.
func TestHeadlessUninstall_YesDoesNotSkipTheHashGuard(t *testing.T) {
	home := t.TempDir()
	a := layeredBoth{home}
	md := a.InstructionsPath(home)
	if err := os.MkdirAll(filepath.Dir(md), 0o755); err != nil {
		t.Fatal(err)
	}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig, Toggles: []string{"tdd"}, InstallModes: []model.InstallMode{model.ModeLite, model.ModeFull}}
	composed, _ := config.Compose(h.Toggles, "claude")
	original := "# mine\n\n" + filemerge.InjectMarkdownSection("", "sdd-orchestrator", composed)
	if err := os.WriteFile(md, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InstallWith(h, []config.AgentAdapter{a}, home, config.InstallOptions{Home: home, DetectVersion: func() claude.VersionCheck {
		return claude.VersionCheck{State: claude.VersionCompatible, Version: "2.1.318"}
	}}); err != nil {
		t.Fatal(err)
	}
	rec, _, _ := config.LoadMigrationRecord(home, md)
	edited := "# edited after migration\n"
	if err := os.WriteFile(md, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	cat := &fakeUninstallCatalog{harnesses: []model.Harness{h}}
	reg := &fakeUninstallRegistry{adapters: map[model.Agent]uninstall.AgentAdapter{model.AgentClaude: a}}
	run := func(flags []string) (int, string) {
		args := append([]string{"--strategy", "restore", "--restore-manifest", filepath.Join(rec.ManifestDir, "manifest.json"),
			"--agent", "claude", "--mode", "lite", "--home", home}, flags...)
		params, err := headless.ParseUninstallFlags(args)
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		return headless.RunHeadlessUninstall(params, cat, reg, &buf), buf.String()
	}

	if code, out := run([]string{"--yes"}); code == 0 || !strings.Contains(out, "--overwrite-edited") {
		t.Errorf("--yes alone must not skip the guard: exit=%d\n%s", code, out)
	}
	if got, _ := os.ReadFile(md); string(got) != edited {
		t.Errorf("CLAUDE.md must stay as the user left it, got %q", got)
	}
	if code, out := run([]string{"--yes", "--overwrite-edited"}); code != 0 {
		t.Errorf("--overwrite-edited must restore: exit=%d\n%s", code, out)
	}
	if got, _ := os.ReadFile(md); string(got) != original {
		t.Errorf("CLAUDE.md not restored byte for byte: %q", got)
	}
}

func TestParseUninstallFlags_OverwriteEditedIsSeparateFromYes(t *testing.T) {
	p, err := headless.ParseUninstallFlags([]string{"--overwrite-edited", "--home", t.TempDir()})
	if err != nil || !p.OverwriteEdited {
		t.Fatalf("--overwrite-edited not parsed: %+v %v", p.OverwriteEdited, err)
	}
	if p, _ := headless.ParseUninstallFlags([]string{"--yes", "--home", t.TempDir()}); p.OverwriteEdited {
		t.Error("--yes must not set OverwriteEdited")
	}
}
