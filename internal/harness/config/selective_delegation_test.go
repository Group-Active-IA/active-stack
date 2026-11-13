package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/harness/config"
)

// opsx-selective-delegation (07). The legacy single block (Compose) stays
// byte-identical (TestComposeLegacy_ByteIdenticalToGolden); 07 changes the
// layered Claude delivery and the native Codex package.

// Obligations to delegate that 07 removes (design D8, tasks 2.1).
var delegationObligations = []string{
	"ALWAYS delegate",
	"Delegate, don't inflate",
	"delegate real work",
	"Delegates real work",
	"For each action, delegate to a sub-agent",
	"You delegate the skill's work to a sub-agent",
	"Read 4+ files",
	"1-3 files",
}

func claudeDelivery(t *testing.T, toggles []string) string {
	t.Helper()
	l := mustLayers(t, toggles)
	var b strings.Builder
	b.WriteString(config.RenderSkillMD(l.Skill))
	for _, n := range moduleNames(l) {
		b.WriteString("\n" + l.Modules[n])
	}
	return b.String()
}

func codexDelivery(t *testing.T, toggles []string) config.CodexPackage {
	t.Helper()
	pkg, err := config.ComposeCodexPackage(toggles)
	if err != nil {
		t.Fatalf("ComposeCodexPackage: %v", err)
	}
	return pkg
}

// 2.1: with or without the toggle, no layer keeps an obligation to delegate.
func TestSelectiveDelegation_NoObligationToDelegate(t *testing.T) {
	for _, toggles := range [][]string{nil, allToggles} {
		claude := claudeDelivery(t, toggles)
		var codex strings.Builder
		for _, c := range codexDelivery(t, toggles).Files {
			codex.WriteString(c)
		}
		for _, bad := range delegationObligations {
			if strings.Contains(claude, bad) {
				t.Errorf("claude delivery (toggles=%v) still contains %q", toggles, bad)
			}
			if strings.Contains(codex.String(), bad) {
				t.Errorf("codex package (toggles=%v) still contains %q", toggles, bad)
			}
		}
	}
}

// 2.2: with delegation the rule is in the coordinator, independence first, and
// the old 1-3 / 4+ threshold is gone.
func TestSelectiveDelegation_RuleOrderWithToggle(t *testing.T) {
	bodies := map[string]string{
		"claude": mustLayers(t, allToggles).Skill.Body,
		"codex":  codexDelivery(t, allToggles).Files["SKILL.md"],
	}
	steps := []string{
		"## Direct or delegated",
		"1. **Independent review required**",
		"2. **File conflict**",
		"3. **Work directly**",
		"4. **Delegate**",
		"5. Otherwise, work directly.",
		"`{action, step, signal, direct|delegated|stopped}`",
	}
	for client, body := range bodies {
		last := -1
		for _, s := range steps {
			i := strings.Index(body, s)
			if i < 0 {
				t.Errorf("%s coordinator lacks %q", client, s)
				continue
			}
			if i < last {
				t.Errorf("%s coordinator: %q is out of order", client, s)
			}
			last = i
		}
		if strings.Contains(body, "Delegation is disabled") {
			t.Errorf("%s coordinator with the toggle must not say delegation is disabled", client)
		}
	}
}

// 2.2 / D8: without delegation the coordinator works directly and stops only
// an action that requires independent review.
func TestSelectiveDelegation_RuleWithoutToggle(t *testing.T) {
	bodies := map[string]string{
		"claude": mustLayers(t, nil).Skill.Body,
		"codex":  codexDelivery(t, nil).Files["SKILL.md"],
	}
	for client, body := range bodies {
		for _, want := range []string{
			"## Direct or delegated",
			"1. **Independent review required**",
			"Delegation is disabled in this installation: work directly.",
			"`missing module: delegation — <action> stopped`",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("%s coordinator without delegation lacks %q", client, want)
			}
		}
		for _, bad := range []string{"2. **File conflict**", "4. **Delegate**"} {
			if strings.Contains(body, bad) {
				t.Errorf("%s coordinator without delegation must not contain %q", client, bad)
			}
		}
	}
	if _, ok := codexDelivery(t, nil).Files["modules/delegation.md"]; ok {
		t.Error("codex package without delegation must not ship modules/delegation.md")
	}
}

// 3.1: the delegation module carries the nine assignment fields, ownership,
// parallelism, depth (2.5), the return envelope and attempts.
func TestSelectiveDelegation_ModuleContent(t *testing.T) {
	mod := mustLayers(t, allToggles).Modules["delegation"]
	for _, want := range []string{
		"Load this module only when the delegation rule decided to delegate.",
		"- goal:", "- action:", "- start:", "- write set:", "- read pointers:",
		"- limits:", "- exit criteria:", "- return contract:", "- attempts: at most 3",
		"`go.mod`, `go.sum`",
		"own worktree created from the recorded start",
		"`out_of_scope`",
		"At most two workers with a write set at the same time",
		"Only you launch workers. An assignment never asks a worker to delegate.",
		"status:", "start:", "modified:", "findings:", "receipts:", "pending:",
		"a re-run without its own independent\ncapture does not corroborate anything",
		"Up to 3 attempts per assignment.",
		// 06 final (d3ab8b3): corroboration is bound to the concrete execution.
		"Corroboration must name the command, tree and result it observed",
		"a capture the worker declares about itself is not independent",
	} {
		if !strings.Contains(mod, want) {
			t.Errorf("delegation module lacks %q", want)
		}
	}
}

// 2.5: the sub-agent launch template asks for the return envelope and tells
// the worker not to launch sub-agents.
func TestSelectiveDelegation_LaunchTemplate(t *testing.T) {
	body := mustLayers(t, allToggles).Skill.Body
	if strings.Contains(body, "return a summary of:") {
		t.Error("launch template still asks for a prose summary")
	}
	for _, want := range []string{
		"return the envelope of `${CLAUDE_SKILL_DIR}/modules/delegation.md`",
		"Do not launch sub-agents.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("launch template lacks %q", want)
		}
	}
}

// 3.3: Claude and Codex receive the same rule and the same module text.
func TestSelectiveDelegation_ClaudeCodexParity(t *testing.T) {
	for _, toggles := range [][]string{nil, allToggles} {
		claude := mustLayers(t, toggles)
		codex := codexDelivery(t, toggles)
		rule := func(body string) string {
			_, after, ok := strings.Cut(body, "## Direct or delegated")
			if !ok {
				return ""
			}
			if i := strings.Index(after, "\n## "); i >= 0 {
				after = after[:i]
			}
			return strings.TrimSpace(after)
		}
		if a, b := rule(claude.Skill.Body), rule(codex.Files["SKILL.md"]); a == "" || a != b {
			t.Errorf("toggles=%v: delegation rule differs between Claude and Codex:\nclaude:\n%s\ncodex:\n%s", toggles, a, b)
		}
		if claude.Modules["delegation"] != codex.Files["modules/delegation.md"] {
			t.Errorf("toggles=%v: delegation module differs between Claude and Codex", toggles)
		}
	}
}

// Without delegation the launch template must not point at a module that is
// not installed.
func TestSelectiveDelegation_LaunchTemplateWithoutToggle(t *testing.T) {
	body := mustLayers(t, nil).Skill.Body
	if strings.Contains(body, "modules/delegation.md`\n(status") {
		t.Error("launch template references modules/delegation.md, which is not installed without the toggle")
	}
	for _, want := range []string{"Do not launch sub-agents.", "status, start, modified, findings, receipts, pending"} {
		if !strings.Contains(body, want) {
			t.Errorf("launch template without delegation lacks %q", want)
		}
	}
}

// Hot files (design D5) must name real paths of this repository: a pattern
// that matches nothing protects nothing. The repository root is three levels up.
func TestSelectiveDelegation_HotFilesExist(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("assets", "opsx", "delegation-module.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	const lead = "Never split these across concurrent workers:"
	i := strings.Index(text, lead)
	if i < 0 {
		t.Fatalf("module without %q", lead)
	}
	block, _, _ := strings.Cut(text[i+len(lead):], "\n\n")
	parts := strings.Split(block, "`")
	var hot []string
	for j := 1; j < len(parts); j += 2 {
		hot = append(hot, parts[j])
	}
	if len(hot) == 0 {
		t.Fatal("no hot files listed")
	}
	root := filepath.Join("..", "..", "..")
	for _, h := range hot {
		m, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(h)))
		if err != nil || len(m) == 0 {
			t.Errorf("hot file %q matches no file in the repository", h)
		}
	}
}
