package config

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

// Layers is the layered Claude composition (claude-native-opsx 4.1): the
// universal block for CLAUDE.md, the coordinator skill and its modules.
type Layers struct {
	// Universal is the content of the active-stack:universal-rules section.
	Universal string
	// Skill is the coordinator skill (frontmatter is added by RenderSkillMD).
	Skill SkillSpec
	// Modules maps module name to content (modules/<name>.md).
	Modules map[string]string
}

// coordinatorSkillDescription is the SKILL.md frontmatter description.
const coordinatorSkillDescription = "OPSX coordinator role. Decides per action whether to work directly or delegate, and synthesizes results. Invoke it with /sdd-orchestrator."

// nestedMarkerRe matches the legacy nested toggle markers kept inside the
// orchestrator asset (sdd-delegation, sdd-model-assignments); modules drop them.
var nestedMarkerRe = regexp.MustCompile(`(?m)^<!-- /?active-stack:sdd-[a-z-]+ -->\n?`)

// ComposeLayers splits the Claude assets according to openspec/matriz-reglas-opsx.md.
// Text is only relocated, never rewritten; the legacy single block is still
// produced by Compose from the same annotated assets.
func ComposeLayers(toggles []string) (Layers, error) {
	for _, t := range toggles {
		if !knownToggles[t] {
			return Layers{}, fmt.Errorf("config compose layers: unknown toggle %q", t)
		}
	}
	active := make(map[string]bool, len(toggles))
	for _, t := range toggles {
		active[t] = true
	}

	orch, err := loadLayerFragments("claude/sdd-orchestrator.md")
	if err != nil {
		return Layers{}, err
	}
	gov, err := loadLayerFragments("governance.md")
	if err != nil {
		return Layers{}, err
	}
	eng, err := loadLayerFragments("engram-protocol.md")
	if err != nil {
		return Layers{}, err
	}
	byName := map[string]string{}
	for _, f := range append(append(append([]layerFragment{}, orch...), gov...), eng...) {
		key := f.layer + "/" + f.name
		if _, dup := byName[key]; dup {
			return Layers{}, fmt.Errorf("config compose layers: duplicate fragment %q", key)
		}
		byName[key] = f.text
	}
	frag := func(key string) (string, error) {
		s, ok := byName[key]
		if !ok {
			return "", fmt.Errorf("config compose layers: missing fragment %q", key)
		}
		return s, nil
	}
	pick := func(keys ...string) ([]string, error) {
		var out []string
		for _, k := range keys {
			s, err := frag(k)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	}

	intro, err := fs.ReadFile(embeddedAssets, "assets/claude/universal-intro.md")
	if err != nil {
		return Layers{}, fmt.Errorf("config compose layers: %w", err)
	}

	// Structural headings (matrix "Reglas de la tabla"): repeated in every layer
	// that has content below them; numbering is rebuilt inside each layer.
	coreHeading, err := frag("structure/core-principles")
	if err != nil {
		return Layers{}, err
	}
	rulesText, err := frag("universal/rules")
	if err != nil {
		return Layers{}, err
	}
	rulesHeading, _, _ := strings.Cut(rulesText, "\n")

	// Universal: matrix rows 3, 6, 8, 15, 18, 20, 22.
	sentence, err := frag("universal/sentence-cli-source-of-truth")
	if err != nil {
		return Layers{}, err
	}
	var principles []string
	p1, err := frag("universal/principle-cli-owns-state")
	if err != nil {
		return Layers{}, err
	}
	principles = append(principles, p1)
	if active["engram"] {
		p3, err := frag("universal/principle-engram")
		if err != nil {
			return Layers{}, err
		}
		principles = append(principles, p3)
	}
	uBlocks := []string{string(intro), sentence, renumbered(strings.TrimSpace(coreHeading), principles), rulesText}
	if active["governance"] {
		g, err := frag("universal/governance-levels")
		if err != nil {
			return Layers{}, err
		}
		uBlocks = append(uBlocks, g)
	}
	if active["engram"] {
		e, err := pick("universal/engram-header", "universal/engram-save-search")
		if err != nil {
			return Layers{}, err
		}
		uBlocks = append(uBlocks, e...)
	}
	universal := joinBlocks(uBlocks...)

	// Coordinator body (D10): role, mandatory rules and the action table first.
	roleParts, err := pick("coordinator/role", "coordinator/role-fluid")
	if err != nil {
		return Layers{}, err
	}
	// opsx-selective-delegation (07): the coordinator decides per action instead
	// of carrying the obligation to delegate. coordinator/principle-delegate and
	// the "ALWAYS delegate" rule are replaced by the shared delegation rule; the
	// legacy single block (Compose) keeps the original text.
	if roleParts[0], err = replaceOnce(roleParts[0], roleDelegateOld, roleDelegateNew); err != nil {
		return Layers{}, err
	}
	// "not an executor" contradicts the rule's "work directly" (PR #13 review).
	if roleParts[0], err = replaceOnce(roleParts[0], roleExecutorOld, roleExecutorNew); err != nil {
		return Layers{}, err
	}
	rule, err := delegationRule(active["delegation"])
	if err != nil {
		return Layers{}, err
	}
	cRules, err := frag("coordinator/rules")
	if err != nil {
		return Layers{}, err
	}
	if cRules, err = replaceOnce(cRules, alwaysDelegateRule, ""); err != nil {
		return Layers{}, err
	}
	head := append(roleParts,
		rule,
		strings.TrimSpace(rulesHeading)+"\n\n"+strings.TrimSpace(cRules),
		actionTable(active))
	restKeys := []string{"coordinator/starting-work", "coordinator/workflow", "coordinator/launch-pattern"}
	if active["governance"] {
		restKeys = append(restKeys, "coordinator/governance-apply")
	}
	if active["engram"] {
		restKeys = append(restKeys, "coordinator/engram-session-close")
	}
	rest, err := pick(restKeys...)
	if err != nil {
		return Layers{}, err
	}
	if active["model-routing"] {
		for i, r := range rest {
			rest[i] = strings.Replace(r, launchModelOld, launchModelNew, 1)
		}
	}
	envelopeRef := "the envelope of `${CLAUDE_SKILL_DIR}/modules/delegation.md`\n"
	if !active["delegation"] {
		envelopeRef = "the return envelope "
	}
	for i, r := range rest {
		for _, c := range layeredDelegationRewrites {
			r = strings.Replace(r, c[0], strings.Replace(c[1], "{envelope}", envelopeRef, 1), 1)
		}
		rest[i] = r
	}
	body := joinBlocks(append(head, rest...)...)

	// Modules.
	modules := map[string]string{}
	art, err := frag("module/artifacts")
	if err != nil {
		return Layers{}, err
	}
	modules["artifacts"] = joinBlocks(art)
	if active["delegation"] {
		d, err := delegationModule()
		if err != nil {
			return Layers{}, err
		}
		modules["delegation"] = d
	}
	if active["model-routing"] {
		m, err := frag("module/model-assignments")
		if err != nil {
			return Layers{}, err
		}
		modules["model-assignments"] = joinBlocks(nestedMarkerRe.ReplaceAllString(m, ""))
	}
	if active["engram"] {
		e, err := pick("module/engram-opsx-wiring", "module/engram-opsx-git-sync")
		if err != nil {
			return Layers{}, err
		}
		modules["engram-opsx"] = joinBlocks(e...)
	}
	if active["tdd"] {
		strict, err := loadFragment("strict-tdd.md")
		if err != nil {
			return Layers{}, err
		}
		// Matrix row 25: the flag line added by Compose belongs to the tdd module.
		modules["tdd"] = joinBlocks("Strict TDD Mode: enabled\n\n" + strict)
	}

	return Layers{
		Universal: universal,
		Skill: SkillSpec{
			ID:          "sdd-orchestrator",
			Description: coordinatorSkillDescription,
			Body:        body,
			Modules:     modules,
		},
		Modules: modules,
	}, nil
}

// launchModelOld/New: the one authorized text change of the layered coordinator
// (the "table above" no longer exists once model-assignments is a module).
const (
	launchModelOld = `model: "<model from table above>",`
	launchModelNew = "model: \"<model from ${CLAUDE_SKILL_DIR}/modules/model-assignments.md>\","
)

// Text changes of opsx-selective-delegation (07) on the layered coordinator.
const (
	roleExecutorOld    = "You are a COORDINATOR, not an executor."
	roleExecutorNew    = "You are a COORDINATOR."
	roleDelegateOld    = "delegate real work to sub-agents via the **Agent tool**"
	roleDelegateNew    = "decide per action whether to work directly or delegate to sub-agents via the **Agent tool**"
	alwaysDelegateRule = "- NEVER do apply or propose work inline — ALWAYS delegate via Agent tool\n"
)

// layeredDelegationRewrites: the workflow and launch-pattern lines that assumed
// every action is delegated, and the prose summary replaced by the envelope.
var layeredDelegationRewrites = [][2]string{
	{"### For each action, delegate to a sub-agent:", "### For each action, apply the delegation rule:"},
	{"**You delegate the skill's work to a sub-agent. You don't replicate skill logic inline.**",
		"**When you delegate, the sub-agent runs the skill. You don't replicate skill logic inline.**"},
	{"Follow the skill's instructions completely. When done, return a summary of:\n- What was accomplished\n- Files created or changed\n- Domain skills loaded and applied\n- Any issues or blockers found\n- Recommended next action",
		"Follow the skill's instructions completely. Do not launch sub-agents.\nWhen done, return {envelope}(status, start, modified, findings, receipts, pending) and the domain skills loaded."},
}

// replaceOnce replaces old by new and fails if old is not present exactly once,
// so an edited asset cannot silently keep an obligation that 07 removes.
func replaceOnce(s, old, new string) (string, error) {
	if n := strings.Count(s, old); n != 1 {
		return "", fmt.Errorf("config compose layers: expected one %q, found %d", strings.TrimSpace(old), n)
	}
	return strings.Replace(s, old, new, 1), nil
}

// delegationRule is the coordinator's delegation rule (07 D1, D8), the same
// text for Claude and Codex.
func delegationRule(enabled bool) (string, error) {
	name := "assets/opsx/delegation-rule-disabled.md"
	if enabled {
		name = "assets/opsx/delegation-rule.md"
	}
	b, err := fs.ReadFile(embeddedAssets, name)
	if err != nil {
		return "", fmt.Errorf("config delegation rule: %w", err)
	}
	return joinBlocks(string(b)), nil
}

// delegationModule is the shared delegation module (07 D3-D7, D10).
func delegationModule() (string, error) {
	b, err := fs.ReadFile(embeddedAssets, "assets/opsx/delegation-module.md")
	if err != nil {
		return "", fmt.Errorf("config delegation module: %w", err)
	}
	return joinBlocks(string(b)), nil
}

var leadingNumberRe = regexp.MustCompile(`^\d+\. `)

// renumbered returns heading + the items with their numbering rebuilt (1., 2., ...).
func renumbered(heading string, items []string) string {
	out := []string{heading}
	for i, it := range items {
		out = append(out, fmt.Sprintf("%d. ", i+1)+leadingNumberRe.ReplaceAllString(strings.TrimSpace(it), ""))
	}
	return strings.Join(out, "\n\n")
}

// actionTable is the action -> modules table of the coordinator body (matrix,
// "Reglas de la tabla"). A module whose toggle is inactive is not listed, and a
// row left without modules is dropped.
func actionTable(active map[string]bool) string {
	exists := func(n string) bool {
		switch n {
		case "artifacts":
			return true
		case "tdd":
			return active["tdd"]
		case "delegation":
			return active["delegation"]
		case "model-assignments":
			return active["model-routing"]
		case "engram-opsx":
			return active["engram"]
		}
		return false
	}
	rows := []struct {
		action  string
		modules []string
	}{
		{"propose", []string{"artifacts"}},
		{"apply", []string{"artifacts", "tdd"}},
		{"archive", []string{"artifacts"}},
		{"delegate", []string{"delegation", "model-assignments"}},
		{"session start and close", []string{"engram-opsx"}},
	}
	var b strings.Builder
	b.WriteString("## Action to Modules\n\n")
	b.WriteString("Before an action, read ONLY the modules listed for it. When you delegate, paste their content into the sub-agent prompt: the sub-agent does not see this skill.\n\n")
	b.WriteString("| Action | Modules to read |\n|--------|-----------------|\n")
	for _, r := range rows {
		var refs []string
		for _, m := range r.modules {
			if exists(m) {
				refs = append(refs, fmt.Sprintf("`${CLAUDE_SKILL_DIR}/modules/%s.md`", m))
			}
		}
		if len(refs) > 0 {
			b.WriteString("| " + r.action + " | " + strings.Join(refs, ", ") + " |\n")
		}
	}
	b.WriteString("\nIf a module listed in the table is missing or empty, do not run that action: name the module and its path, tell the user to reinstall with `active-stack install`, and continue with the actions that do not depend on it.\n\n")
	b.WriteString("NEVER delegate the coordinator role to a sub-agent or ask one to act as `sdd-orchestrator`.\n")
	return b.String()
}

// joinBlocks trims each block, drops empty ones, joins with a blank line and
// ends with a single newline.
func joinBlocks(blocks ...string) string {
	var out []string
	for _, b := range blocks {
		if t := strings.Trim(b, "\n \t"); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n\n") + "\n"
}

// ExpectedModules lists the module names ComposeLayers produces for the given
// toggles (artifacts always), so verify can check that none is missing.
func ExpectedModules(toggles []string) []string {
	names := []string{"artifacts"}
	for _, t := range toggles {
		switch t {
		case "delegation":
			names = append(names, "delegation")
		case "model-routing":
			names = append(names, "model-assignments")
		case "engram":
			names = append(names, "engram-opsx")
		case "tdd":
			names = append(names, "tdd")
		}
	}
	return names
}
