package config_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
)

var allToggles = []string{"delegation", "model-routing", "engram", "tdd", "governance"}

func mustLayers(t *testing.T, toggles []string) config.Layers {
	t.Helper()
	l, err := config.ComposeLayers(toggles)
	if err != nil {
		t.Fatalf("ComposeLayers: %v", err)
	}
	return l
}

func moduleNames(l config.Layers) []string {
	var n []string
	for k := range l.Modules {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

// 4.2 (M): the legacy composition stays byte for byte equal to the 35a89f8 golden.
func TestComposeLegacy_ByteIdenticalToGolden(t *testing.T) {
	got, err := config.Compose(allToggles, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if want := loadGolden(t, "compose_all_toggles_claude"); got != want {
		t.Errorf("legacy composition differs from the 35a89f8 golden: got len=%d want len=%d", len(got), len(want))
	}
	if strings.Contains(got, "<!-- as:") {
		t.Error("layer annotations leaked into the legacy composition")
	}
}

// 4.2: the universal composition has no coordinator headings and no modules.
func TestComposeLayers_UniversalHasNoCoordinatorNorModules(t *testing.T) {
	u := mustLayers(t, allToggles).Universal
	for _, want := range []string{"/sdd-orchestrator", "NEVER guess artifact state", "NEVER create `openspec/` structure manually", "### Governance Levels", "PROACTIVE SAVE TRIGGERS", "Engram persists context"} {
		if !strings.Contains(u, want) {
			t.Errorf("universal missing %q", want)
		}
	}
	for _, bad := range []string{"## Role", "You are a COORDINATOR", "## Delegation Rules", "## Model Assignments", "## Artifact Lifecycle", "Strict TDD Mode", "### OPSX wiring", "### GIT SYNC", "### How to Apply", "SESSION CLOSE PROTOCOL", "NEVER block on phase gates", "Bind this to the dedicated", "<!-- active-stack:sdd-"} {
		if strings.Contains(u, bad) {
			t.Errorf("universal must not contain %q", bad)
		}
	}
}

// 4.2 (M): SKILL.md carries role, rules, table and the frontmatter flag, and none
// of the module text.
func TestComposeLayers_SkillMDHasRoleTableAndNoModuleText(t *testing.T) {
	l := mustLayers(t, allToggles)
	md := config.RenderSkillMD(l.Skill)
	fm := strings.SplitN(md, "\n---\n", 2)[0]
	for _, ln := range strings.Split(fm, "\n")[1:] {
		if strings.Count(ln, ": ") > 1 {
			t.Errorf("frontmatter line is not valid plain YAML (second `: `): %q", ln)
		}
	}
	for _, want := range []string{"disable-model-invocation: true", "name: sdd-orchestrator", "## Role", "You are a COORDINATOR", "NEVER block on phase gates", "| Action |", "modules/artifacts.md", "modules/tdd.md", "modules/delegation.md", "### How to Apply", "SESSION CLOSE PROTOCOL", "## Starting Work on a Project", "## Sub-Agent Launch Pattern"} {
		if !strings.Contains(md, want) {
			t.Errorf("SKILL.md missing %q", want)
		}
	}
	for _, bad := range []string{"## Artifact Lifecycle", "## Key CLI Commands Reference", "## Delegation Rules", "## Model Assignments", "Strict TDD Mode", "### OPSX wiring", "### GIT SYNC", "Bind this to the dedicated", "<!-- active-stack:sdd-", "### Governance Levels", "PROACTIVE SAVE TRIGGERS"} {
		if strings.Contains(md, bad) {
			t.Errorf("SKILL.md must not contain %q", bad)
		}
	}
	// D10: role, mandatory rules and the action table come before the long sections.
	idx := func(s string) int { return strings.Index(md, s) }
	if !(idx("## Role") < idx("NEVER block on phase gates") && idx("NEVER block on phase gates") < idx("| Action |") && idx("| Action |") < idx("## Starting Work on a Project")) {
		t.Errorf("SKILL.md order must be role, rules, table, rest; got %d %d %d %d", idx("## Role"), idx("NEVER block on phase gates"), idx("| Action |"), idx("## Starting Work on a Project"))
	}
}

// 4.2 (M): every module exists only with its toggle; artifacts always.
func TestComposeLayers_ModulesFollowToggles(t *testing.T) {
	if got := moduleNames(mustLayers(t, nil)); strings.Join(got, ",") != "artifacts" {
		t.Errorf("no toggles: modules = %v, want [artifacts]", got)
	}
	if got := moduleNames(mustLayers(t, allToggles)); strings.Join(got, ",") != "artifacts,delegation,engram-opsx,model-assignments,tdd" {
		t.Errorf("all toggles: modules = %v", got)
	}
	for toggle, module := range map[string]string{"tdd": "tdd", "delegation": "delegation", "model-routing": "model-assignments", "engram": "engram-opsx"} {
		got := moduleNames(mustLayers(t, []string{toggle}))
		if strings.Join(got, ",") != "artifacts,"+module {
			t.Errorf("toggle %s: modules = %v, want artifacts + %s", toggle, got, module)
		}
	}
	l := mustLayers(t, []string{"governance"})
	if !strings.Contains(l.Universal, "### Governance Levels") || !strings.Contains(config.RenderSkillMD(l.Skill), "### How to Apply") {
		t.Error("governance toggle must add levels to universal and How to Apply to SKILL.md")
	}
	l = mustLayers(t, nil)
	if strings.Contains(l.Universal, "### Governance Levels") || strings.Contains(l.Universal, "PROACTIVE SAVE") || strings.Contains(config.RenderSkillMD(l.Skill), "SESSION CLOSE PROTOCOL") {
		t.Error("governance and engram content must disappear without their toggles")
	}
}

func TestComposeLayers_ModuleContents(t *testing.T) {
	m := mustLayers(t, allToggles).Modules
	heads := map[string]string{
		"artifacts": "## Artifact Lifecycle", "delegation": "## Delegation Rules",
		"model-assignments": "## Model Assignments", "tdd": "Strict TDD Mode: enabled",
		"engram-opsx": "### OPSX wiring",
	}
	for name, head := range heads {
		if !strings.HasPrefix(m[name], head) {
			t.Errorf("modules/%s.md should start with %q, got %.40q", name, head, m[name])
		}
		if strings.Contains(m[name], "<!-- as:") || strings.Contains(m[name], "<!-- active-stack:") {
			t.Errorf("modules/%s.md leaks markers", name)
		}
	}
	if !strings.Contains(m["engram-opsx"], "### GIT SYNC") {
		t.Error("engram-opsx must include the git sync section")
	}
	if !strings.Contains(m["artifacts"], "## Key CLI Commands Reference") {
		t.Error("artifacts must include the CLI commands reference")
	}
}

// tbdLines are original lines the matrix leaves without a row (no layer yet).
var tbdLines = map[string]bool{
	"# OPSX Orchestrator Instructions": true,
	"## Core Principles":               true,
}

var ruleLineRe = regexp.MustCompile(`^(#{1,6} |\d+\. |- NEVER |.*\bNEVER\b)`)

// 4.2: "Ninguna regla sin capa": every heading, NEVER line and numbered rule of
// the original assets appears in exactly one layer (or is an explicit TBD).
func TestComposeLayers_NoRuleWithoutALayerNorInTwo(t *testing.T) {
	l := mustLayers(t, allToggles)
	layers := map[string]string{"universal": l.Universal, "skill": config.RenderSkillMD(l.Skill)}
	for n, c := range l.Modules {
		layers["module/"+n] = c
	}
	// The original text is the legacy composition (golden of 35a89f8).
	original := loadGolden(t, "compose_all_toggles_claude")
	count := func(text, line string) int {
		n := 0
		for _, ln := range strings.Split(text, "\n") {
			if ln == line {
				n++
			}
		}
		return n
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(original, "\n") {
		if !ruleLineRe.MatchString(line) || seen[line] {
			continue
		}
		seen[line] = true
		total, in := 0, []string{}
		for name, c := range layers {
			if n := count(c, line); n > 0 {
				total += n
				in = append(in, name)
			}
		}
		if tbdLines[line] {
			if total != 0 {
				t.Errorf("TBD line %q must not be delivered yet, found in %v", line, in)
			}
			continue
		}
		want := count(original, line)
		if total != want {
			t.Errorf("line %q appears %d times in the original and %d in the layers %v", line, want, total, in)
		}
		// A line that occurs once in the original (the same heading text can
		// legitimately occur in two source files) lives in exactly one layer.
		if want == 1 && len(in) != 1 {
			t.Errorf("line %q is in %d layers %v, want exactly 1", line, len(in), in)
		}
	}
	// The line added by the composition lives only in the tdd module.
	for name, c := range layers {
		if has := strings.Contains(c, "Strict TDD Mode: enabled"); has != (name == "module/tdd") {
			t.Errorf("%q: Strict TDD Mode line present=%v", name, has)
		}
	}
}

// 4.4: goldens for universal, SKILL.md and every module.
func TestComposeLayers_Goldens(t *testing.T) {
	l := mustLayers(t, allToggles)
	files := map[string]string{
		"compose_layers_universal": l.Universal,
		"compose_layers_skill":     config.RenderSkillMD(l.Skill),
	}
	for n, c := range l.Modules {
		files["compose_layers_module_"+n] = c
	}
	for name, got := range files {
		if got == "" {
			t.Errorf("%s is empty", name)
			continue
		}
		if *update {
			writeGolden(t, name, got)
			continue
		}
		if want := loadGolden(t, name); got != want {
			t.Errorf("%s differs from its golden: got len=%d want len=%d", name, len(got), len(want))
		}
	}
}
