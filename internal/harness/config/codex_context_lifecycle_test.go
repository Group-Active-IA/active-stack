package config

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestCodexContextModulesInstallUpdateReinstallAndRemove(t *testing.T) {
	home := t.TempDir()
	a := managedTestAdapter{}
	h := model.Harness{ID: "sdd-orchestrator", Type: model.HarnessConfig}
	if err := os.MkdirAll(filepath.Dir(a.InstructionsPath(home)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(a.InstructionsPath(home), []byte("my user rules\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallManagedSkill(h, a, home); err != nil {
		t.Fatal(err)
	}
	point := migrationRoot(a, home, h.ID)
	before, err := os.ReadFile(filepath.Join(point, backup.ArchiveFilename))
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(a.SkillsDir(home), h.ID)
	note := filepath.Join(root, "my-note.md")
	if err = os.WriteFile(note, []byte("my note"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, toggles := range [][]string{{"tdd"}, {"tdd", "engram"}} {
		h.Toggles = toggles
		if _, err = InstallManagedSkill(h, a, home); err != nil {
			t.Fatal(err)
		}
		pkg, err := ComposeCodexPackage(toggles)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"} {
			path := "modules/" + id + ".md"
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
			if err != nil || string(raw) != pkg.Files[path] {
				t.Fatalf("updated module %s: %v", id, err)
			}
		}
		if err = VerifyManagedSkill(h, a, home); err != nil {
			t.Fatal(err)
		}
		res, err := InstallManagedSkill(h, a, home)
		if err != nil || !res.AllAlready || len(res.ManagedWrites) != 0 {
			t.Fatalf("reinstall rewrote managed files: %+v, %v", res, err)
		}
	}
	after, err := os.ReadFile(filepath.Join(point, backup.ArchiveFilename))
	if err != nil || sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatalf("initial migration archive changed: %v", err)
	}
	if _, err = RemoveManagedSkill(a, home, h.ID); err != nil {
		t.Fatal(err)
	}
	if raw, err := os.ReadFile(note); err != nil || string(raw) != "my note" {
		t.Fatalf("user note changed: %s, %v", raw, err)
	}
	for _, id := range []string{"directed-code-reading", "bounded-command-output", "task-tool-selection", "concise-agent-reporting"} {
		if _, err := os.Stat(filepath.Join(root, "modules", id+".md")); !os.IsNotExist(err) {
			t.Fatalf("module %s remained: %v", id, err)
		}
	}
	raw, err := os.ReadFile(a.InstructionsPath(home))
	if err != nil || !strings.Contains(string(raw), "my user rules\n") || strings.Contains(string(raw), "<!-- active-stack:sdd-orchestrator -->") {
		t.Fatalf("instructions not preserved: %s, %v", raw, err)
	}
}
