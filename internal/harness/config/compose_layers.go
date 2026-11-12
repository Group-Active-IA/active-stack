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
const coordinatorSkillDescription = "OPSX coordinator role. Delegates real work to sub-agents and synthesizes results. Invoke it with /sdd-orchestrator."

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

	// Universal: matrix rows 3, 6, 8, 15, 18, 20, 22.
	uKeys := []string{"universal/sentence-cli-source-of-truth", "universal/principle-cli-owns-state"}
	if active["engram"] {
		uKeys = append(uKeys, "universal/principle-engram")
	}
	uKeys = append(uKeys, "universal/rules")
	if active["governance"] {
		uKeys = append(uKeys, "universal/governance-levels")
	}
	if active["engram"] {
		uKeys = append(uKeys, "universal/engram-header", "universal/engram-save-search")
	}
	uParts, err := pick(uKeys...)
	if err != nil {
		return Layers{}, err
	}
	universal := joinBlocks(append([]string{string(intro)}, uParts...)...)

	// Coordinator body (D10): role, mandatory rules and the action table first.
	table := actionTable(active)
	cKeys := []string{"coordinator/role", "coordinator/role-fluid", "coordinator/principle-delegate", "coordinator/rules"}
	head, err := pick(cKeys...)
	if err != nil {
		return Layers{}, err
	}
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
	body := joinBlocks(append(append(head, table), rest...)...)

	// Modules.
	modules := map[string]string{}
	art, err := frag("module/artifacts")
	if err != nil {
		return Layers{}, err
	}
	modules["artifacts"] = joinBlocks(art)
	if active["delegation"] {
		d, err := frag("module/delegation")
		if err != nil {
			return Layers{}, err
		}
		modules["delegation"] = joinBlocks(nestedMarkerRe.ReplaceAllString(d, ""))
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

// actionTable is the short action -> modules table of the coordinator body. It
// lists only modules that exist for the active toggles.
func actionTable(active map[string]bool) string {
	ref := func(names ...string) string {
		var out []string
		for _, n := range names {
			out = append(out, fmt.Sprintf("`${CLAUDE_SKILL_DIR}/modules/%s.md`", n))
		}
		return strings.Join(out, ", ")
	}
	var b strings.Builder
	b.WriteString("## Action to Modules\n\n")
	b.WriteString("Before an action, read ONLY the modules listed for it. When you delegate, paste their content into the sub-agent prompt: the sub-agent does not see this skill.\n\n")
	b.WriteString("| Action | Modules to read |\n|--------|-----------------|\n")
	b.WriteString("| propose | " + ref("artifacts") + " |\n")
	apply := []string{"artifacts"}
	if active["tdd"] {
		apply = append(apply, "tdd")
	}
	b.WriteString("| apply | " + ref(apply...) + " |\n")
	b.WriteString("| archive | " + ref("artifacts") + " |\n")
	if active["delegation"] {
		b.WriteString("| delegate | " + ref("delegation") + " |\n")
	}
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
