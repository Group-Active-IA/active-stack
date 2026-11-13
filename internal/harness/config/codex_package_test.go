package config

import (
	"reflect"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"testing"
)

func TestComposeCodexPackageExplicitActivation(t *testing.T) {
	pkg, err := ComposeCodexPackage(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pkg.Files["agents/openai.yaml"], "allow_implicit_invocation: false") {
		t.Fatal("package must disable implicit skill invocation")
	}
	if !strings.Contains(pkg.Files["SKILL.md"], "openspec list --json") {
		t.Fatal("coordinator must consult OpenSpec before acting")
	}
	if strings.Contains(pkg.Universal, "You are a COORDINATOR") {
		t.Fatal("universal instructions must not activate coordination")
	}
}

func TestComposeCodexPackageToggleModules(t *testing.T) {
	pkg, err := ComposeCodexPackage([]string{"governance", "engram", "delegation", "tdd", "model-routing"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"modules/artifacts.md", "modules/delegation.md", "modules/engram-opsx.md", "modules/tdd.md"} {
		if pkg.Files[path] == "" {
			t.Errorf("missing module %s", path)
		}
		if !strings.Contains(pkg.Files["SKILL.md"], path) {
			t.Errorf("missing module reference %s", path)
		}
	}
	for _, rule := range []string{"Governance Levels", "PROACTIVE SAVE TRIGGERS", "WHEN TO SEARCH MEMORY"} {
		if !strings.Contains(pkg.Universal, rule) {
			t.Errorf("missing universal rule %s", rule)
		}
	}
	if strings.Contains(pkg.Universal, "Strict TDD Mode") {
		t.Fatal("TDD must load by action")
	}
	if strings.Contains(pkg.Universal, "SESSION CLOSE PROTOCOL") {
		t.Fatal("coordinator session lifecycle must not reach specialists")
	}
	for _, alias := range []string{"opus", "sonnet", "haiku", "/opsx:"} {
		for path, content := range pkg.Files {
			if strings.Contains(content, alias) {
				t.Errorf("legacy alias %q in %s", alias, path)
			}
		}
	}
}

func TestComposeCodexPackageRejectsUnknownToggle(t *testing.T) {
	_, err := ComposeCodexPackage([]string{"engram", "typo"})
	if err == nil {
		t.Fatal("unknown toggle must fail before producing content")
	}
}

func TestComposeCodexPackageAllToggleCombinations(t *testing.T) {
	names := []string{"delegation", "engram", "governance", "model-routing", "tdd"}
	references := regexp.MustCompile("modules/[a-z-]+\\.md")
	for mask := 0; mask < 1<<len(names); mask++ {
		var toggles []string
		for i, name := range names {
			if mask&(1<<i) != 0 {
				toggles = append(toggles, name)
			}
		}
		t.Run(strings.Join(toggles, "+"), func(t *testing.T) {
			pkg, err := ComposeCodexPackage(toggles)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range references.FindAllString(pkg.Files["SKILL.md"], -1) {
				if pkg.Files[file] == "" {
					t.Errorf("dangling action reference %s", file)
				}
			}
			for name, file := range map[string]string{"delegation": "modules/delegation.md", "engram": "modules/engram-opsx.md", "tdd": "modules/tdd.md"} {
				enabled := false
				for _, toggle := range toggles {
					if toggle == name {
						enabled = true
					}
				}
				if (pkg.Files[file] != "") != enabled {
					t.Errorf("toggle %s does not control module %s", name, file)
				}
			}
			sort.Sort(sort.Reverse(sort.StringSlice(toggles)))
			if len(toggles) > 0 {
				toggles = append(toggles, toggles[0])
			}
			repeated, err := ComposeCodexPackage(toggles)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(pkg, repeated) {
				t.Fatal("composition depends on toggle order or duplicates")
			}
		})
	}
}

func TestComposeCodexPackageMetadata(t *testing.T) {
	pkg, err := ComposeCodexPackage(nil)
	if err != nil {
		t.Fatal(err)
	}
	body := pkg.Files["SKILL.md"]
	if !strings.HasPrefix(body, "---\n") {
		t.Fatal("skill requires real YAML frontmatter")
	}
	frontmatter, _, ok := strings.Cut(strings.TrimPrefix(body, "---\n"), "\n---\n")
	if !ok {
		t.Fatal("unterminated skill frontmatter")
	}
	var skill struct {
		Name        string            `yaml:"name"`
		Description string            `yaml:"description"`
		Metadata    map[string]string `yaml:"metadata"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		t.Fatal(err)
	}
	if skill.Name != "sdd-orchestrator" || skill.Description == "" || skill.Metadata["managed-by"] != "active-stack" {
		t.Fatalf("invalid skill metadata %+v", skill)
	}
	var policy struct {
		Policy struct {
			AllowImplicit *bool `yaml:"allow_implicit_invocation"`
		} `yaml:"policy"`
	}
	if err := yaml.Unmarshal([]byte(pkg.Files["agents/openai.yaml"]), &policy); err != nil {
		t.Fatal(err)
	}
	if policy.Policy.AllowImplicit == nil || *policy.Policy.AllowImplicit {
		t.Fatal("policy must explicitly disable implicit invocation")
	}
}

func TestSplitCodexAssetMissingBoundaries(t *testing.T) {
	for _, source := range []string{"unrelated content", "## start\nbody"} {
		_, _, err := splitCodexAsset(source, "## start", "## boundary")
		if err == nil {
			t.Fatalf("missing boundary accepted in %q", source)
		}
	}
}
