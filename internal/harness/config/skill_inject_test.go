package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
)

func testSpec() config.SkillSpec {
	return config.SkillSpec{
		ID:          "sdd-orchestrator",
		Description: "Coordinator role",
		Body:        "BODY-WITNESS\n",
		Modules:     map[string]string{"tdd": "TDD-WITNESS\n", "governance": "GOV-WITNESS\n"},
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
		return ""
	}
	return string(b)
}

// tree returns relative path -> content for every file under root.
func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, _ := os.ReadFile(p)
		rel, _ := filepath.Rel(root, p)
		out[rel] = string(b)
		return nil
	})
	return out
}

func sameTree(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

// 5.1: writes SKILL.md (frontmatter + ownership mark) and modules/; a second
// run reports no change and leaves identical bytes.
func TestInjectManagedSkill_WritesAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	skills := filepath.Join(root, "skills")
	opts := config.SkillInjectOptions{BackupDir: filepath.Join(root, "backups")}

	res, err := config.InjectManagedSkill(skills, testSpec(), opts)
	if err != nil {
		t.Fatalf("first inject: %v", err)
	}
	if !res.Changed {
		t.Error("first inject must report Changed")
	}
	md := mustRead(t, filepath.Join(skills, "sdd-orchestrator", "SKILL.md"))
	for _, want := range []string{"name: sdd-orchestrator", "disable-model-invocation: true", config.SkillOwnerKey + ": true", "BODY-WITNESS"} {
		if !strings.Contains(md, want) {
			t.Errorf("SKILL.md missing %q:\n%s", want, md)
		}
	}
	if got := mustRead(t, filepath.Join(skills, "sdd-orchestrator", "modules", "tdd.md")); got != "TDD-WITNESS\n" {
		t.Errorf("modules/tdd.md = %q", got)
	}
	if got := mustRead(t, filepath.Join(skills, "sdd-orchestrator", "modules", "governance.md")); got != "GOV-WITNESS\n" {
		t.Errorf("modules/governance.md = %q", got)
	}
	before := tree(t, skills)

	res2, err := config.InjectManagedSkill(skills, testSpec(), opts)
	if err != nil {
		t.Fatalf("second inject: %v", err)
	}
	if res2.Changed {
		t.Error("second inject must report no change")
	}
	if !sameTree(before, tree(t, skills)) {
		t.Error("second inject changed bytes")
	}
}

// Module names are file names: a traversal attempt writes nothing.
func TestInjectManagedSkill_RejectsModuleTraversal(t *testing.T) {
	root := t.TempDir()
	skills := filepath.Join(root, "skills")
	spec := testSpec()
	spec.Modules = map[string]string{"../evil": "x"}
	if _, err := config.InjectManagedSkill(skills, spec, config.SkillInjectOptions{BackupDir: filepath.Join(root, "b")}); err == nil {
		t.Error("expected an error for a module name with path separators")
	}
	if _, err := os.Stat(filepath.Join(skills, "sdd-orchestrator")); err == nil {
		t.Error("nothing must be written when a module name is invalid")
	}
}

// 5.2 (C9): a directory with the same ID and no ownership mark is never touched,
// in Machine and Project layouts; sibling skills stay identical.
func TestInjectManagedSkill_RefusesUnownedDirectory(t *testing.T) {
	for _, target := range []string{"machine-home/.claude/skills", "project-root/.claude/skills"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			skills := filepath.Join(root, filepath.FromSlash(target))
			own := filepath.Join(skills, "sdd-orchestrator")
			other := filepath.Join(skills, "other-skill")
			for p, c := range map[string]string{
				filepath.Join(own, "SKILL.md"):   "---\nname: sdd-orchestrator\n---\nMY OWN SKILL\n",
				filepath.Join(own, "notes.txt"):  "mine",
				filepath.Join(other, "SKILL.md"): "other",
			} {
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			before := tree(t, skills)

			_, err := config.InjectManagedSkill(skills, testSpec(), config.SkillInjectOptions{BackupDir: filepath.Join(root, "b")})
			if err == nil {
				t.Fatal("expected an error for a skill without the ownership mark")
			}
			if !strings.Contains(err.Error(), own) {
				t.Errorf("error must name the path %q, got: %v", own, err)
			}
			if !sameTree(before, tree(t, skills)) {
				t.Error("files changed despite the refusal")
			}
		})
	}
}

// 5.3: installing into Project while a Machine skill with the same ID exists
// warns about the precedence and names both paths.
func TestInjectManagedSkill_WarnsWhenMachineSkillShadowsProject(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project", ".claude", "skills")
	machine := filepath.Join(root, "home", ".claude", "skills")
	if err := os.MkdirAll(filepath.Join(machine, "sdd-orchestrator"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(machine, "sdd-orchestrator", "SKILL.md"), []byte("personal"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := config.InjectManagedSkill(project, testSpec(), config.SkillInjectOptions{
		BackupDir: filepath.Join(root, "b"), ShadowSkillsDir: machine,
	})
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	pj, mc := filepath.Join(project, "sdd-orchestrator"), filepath.Join(machine, "sdd-orchestrator")
	ok := false
	for _, w := range res.Warnings {
		if strings.Contains(w, pj) && strings.Contains(w, mc) && strings.Contains(strings.ToLower(w), "precedence") {
			ok = true
		}
	}
	if !ok {
		t.Errorf("want a precedence warning naming %q and %q, got %q", pj, mc, res.Warnings)
	}
}
