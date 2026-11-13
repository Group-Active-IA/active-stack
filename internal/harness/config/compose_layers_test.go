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

// droppedLines are original lines the layered delivery eliminates (matrix: the
// asset title, like row 4).
var droppedLines = map[string]bool{"# OPSX Orchestrator Instructions": true}

// structuralHeadings group content and repeat in every layer that has content
// below them ("Estructura no es regla").
var structuralHeadings = map[string]bool{"## Core Principles": true, "## Rules": true}

var ruleLineRe = regexp.MustCompile(`^(#{1,6} |\d+\. |.*\bNEVER\b)`)
var numberRe = regexp.MustCompile(`^\d+\. `)

// 4.2: "Ninguna regla sin capa": every heading, NEVER line and rule of the
// original assets appears in exactly one layer, with the same count. Structural
// headings may repeat across layers and numbering is not part of a rule.
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
			if numberRe.ReplaceAllString(ln, "") == line {
				n++
			}
		}
		return n
	}
	seen := map[string]bool{}
	for _, raw := range strings.Split(original, "\n") {
		if !ruleLineRe.MatchString(raw) {
			continue
		}
		line := numberRe.ReplaceAllString(raw, "")
		if seen[line] {
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
		switch {
		case droppedLines[line]:
			if total != 0 {
				t.Errorf("dropped line %q must not be delivered, found in %v", line, in)
			}
		case structuralHeadings[line]:
			if total == 0 {
				t.Errorf("structural heading %q is in no layer", line)
			}
		default:
			want := count(original, line)
			if total != want {
				t.Errorf("line %q appears %d times in the original and %d in the layers %v", line, want, total, in)
			}
			// A line that occurs once in the original lives in exactly one layer.
			if want == 1 && len(in) != 1 {
				t.Errorf("line %q is in %d layers %v, want exactly 1", line, len(in), in)
			}
		}
	}
	// The line added by the composition lives only in the tdd module.
	for name, c := range layers {
		if has := strings.Contains(c, "Strict TDD Mode: enabled"); has != (name == "module/tdd") {
			t.Errorf("%q: Strict TDD Mode line present=%v", name, has)
		}
	}
}

// Structure: grouping headings repeat per layer and numbering restarts in each.
func TestComposeLayers_StructuralHeadingsAndNumbering(t *testing.T) {
	l := mustLayers(t, allToggles)
	md := config.RenderSkillMD(l.Skill)
	if !strings.Contains(l.Universal, "## Core Principles\n\n1. **The `openspec` CLI owns all state.**") ||
		!strings.Contains(l.Universal, "\n\n2. **Engram persists context.**") || strings.Contains(l.Universal, "3. **Engram") {
		t.Errorf("universal must list Core Principles 1 and 2:\n%.600s", l.Universal)
	}
	if !strings.Contains(md, "## Core Principles\n\n1. **Delegate, don't inflate.**") || strings.Contains(md, "2. **Delegate") {
		t.Error("SKILL.md must list Core Principles with the Delegate item as 1")
	}
	if !strings.Contains(md, "## Rules\n\n- NEVER block on phase gates") {
		t.Error("SKILL.md NEVER bullets must sit under their own ## Rules")
	}
	if strings.Contains(l.Universal, "# OPSX Orchestrator Instructions") || strings.Contains(md, "# OPSX Orchestrator Instructions") {
		t.Error("the asset title must be dropped")
	}
	// Without engram the universal keeps only item 1.
	if u := mustLayers(t, nil).Universal; strings.Contains(u, "2. **") {
		t.Errorf("without engram the universal has a single principle: %.400s", u)
	}
}

func tableOf(t *testing.T, toggles []string) string {
	t.Helper()
	md := config.RenderSkillMD(mustLayers(t, toggles).Skill)
	i := strings.Index(md, "## Action to Modules")
	j := strings.Index(md, "## Starting Work on a Project")
	if i < 0 || j < i {
		t.Fatalf("action table not found in SKILL.md")
	}
	return md[i:j]
}

// Action -> modules table: fixed rows, and a module of an inactive toggle is not listed.
func TestComposeLayers_ActionTableFollowsToggles(t *testing.T) {
	all := tableOf(t, allToggles)
	for _, want := range []string{
		"| propose | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |",
		"| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md`, `${CLAUDE_SKILL_DIR}/modules/tdd.md` |",
		"| archive | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |",
		"| delegate | `${CLAUDE_SKILL_DIR}/modules/delegation.md`, `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |",
		"| session start and close | `${CLAUDE_SKILL_DIR}/modules/engram-opsx.md` |",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("table missing row %q:\n%s", want, all)
		}
	}
	none := tableOf(t, nil)
	for _, bad := range []string{"tdd.md", "delegation.md", "model-assignments.md", "engram-opsx.md", "| delegate", "| session"} {
		if strings.Contains(none, bad) {
			t.Errorf("no toggles: table must not list %q:\n%s", bad, none)
		}
	}
	if !strings.Contains(none, "| apply | `${CLAUDE_SKILL_DIR}/modules/artifacts.md` |") {
		t.Errorf("no toggles: apply keeps artifacts only:\n%s", none)
	}
	for toggle, absent := range map[string][]string{
		"tdd":           {"delegation.md", "model-assignments.md", "engram-opsx.md"},
		"delegation":    {"tdd.md", "model-assignments.md", "engram-opsx.md"},
		"model-routing": {"tdd.md", "delegation.md", "engram-opsx.md"},
		"engram":        {"tdd.md", "delegation.md", "model-assignments.md"},
	} {
		tb := tableOf(t, []string{toggle})
		for _, bad := range absent {
			if strings.Contains(tb, bad) {
				t.Errorf("only %s active: table must not list %s:\n%s", toggle, bad, tb)
			}
		}
	}
	if tb := tableOf(t, []string{"model-routing"}); !strings.Contains(tb, "| delegate | `${CLAUDE_SKILL_DIR}/modules/model-assignments.md` |") {
		t.Errorf("model-routing alone keeps a delegate row with model-assignments:\n%s", tb)
	}
}

// The launch template points at the model-assignments module in the layered
// coordinator only; the legacy block keeps the original text.
func TestComposeLayers_LaunchTemplateReferencesModelAssignments(t *testing.T) {
	md := config.RenderSkillMD(mustLayers(t, allToggles).Skill)
	if strings.Contains(md, "<model from table above>") || !strings.Contains(md, `model: "<model from ${CLAUDE_SKILL_DIR}/modules/model-assignments.md>",`) {
		t.Error("SKILL.md launch template must reference modules/model-assignments.md")
	}
	legacy, err := config.Compose(allToggles, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(legacy, `model: "<model from table above>",`) {
		t.Error("legacy composition must keep the original launch template text")
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

// skillTokenBudget is the 4000-token threshold Juani fixed on 2026-11-13 for the
// composed SKILL.md (design D10). 02 left no token counter in main, so tokens
// are estimated as bytes/4.
const skillTokenBudget = 4000

// 4.3: the SKILL.md composed with the five toggles fits the re-attachment budget.
func TestComposeLayers_SkillMDFitsReattachBudget(t *testing.T) {
	md := config.RenderSkillMD(mustLayers(t, allToggles).Skill)
	if tokens := (len(md) + 3) / 4; tokens > skillTokenBudget {
		t.Errorf("SKILL.md is ~%d tokens (%d bytes), budget %d", tokens, len(md), skillTokenBudget)
	}
}
