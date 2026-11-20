package config

import (
	"regexp"
	"strings"
	"testing"
)

// [EN] Fase 2 (R-13): contract of the engram-cloud section text.

// engramCloudMaxBytes is the size cap of the section (T-F2-9). The plan left the
// number open (TBD); it was confirmed on 2026-11-20 (R-14 card, decision on the
// R-13 cap). It changes only by an explicit decision.
const engramCloudMaxBytes = 2500

func engramCloudText(t *testing.T) string {
	t.Helper()
	got, err := loadFragment("engram-cloud.md")
	if err != nil {
		t.Fatalf("loadFragment(engram-cloud.md): %v", err)
	}
	return got
}

// T-F2-9: invariants of the content. Each case is a distinct instruction the
// agent must receive; deleting that line from the asset fails exactly that case.
func TestEngramCloudFragment_Invariants(t *testing.T) {
	norm := strings.Join(strings.Fields(engramCloudText(t)), " ")
	cases := []struct{ name, want string }{
		{"heading", "## Cloud Team Memory"},
		{"extends the engram protocol block", "extends the Engram Protocol"},
		{"applies only with the project marker", ".engram/active-stack.json"},
		{"marker policy value", "cloud-branch"},
		{"ignored otherwise", "ignore this whole section"},
		{"overrides proactive-save routing", "proactive-save"},
		{"overrides opsx keys", "opsx/"},
		{"overrides git sync section", "GIT SYNC"},
		{"never git sync chunks", ".engram/chunks"},
		{"author is added for you", "Author:"},
		{"do not tag by hand", "do NOT add tags"},
		{"untagged is merged team knowledge", "merged team knowledge"},
		{"current branch memories", "branch/<current slug>"},
		{"other branch is unverified", "UNVERIFIED"},
		{"untagged wins on contradiction", "untagged one wins"},
		{"slug may be unavailable", "treat every `branch/"},
		{"closing only after merge", "merged into the default branch"},
		{"only own memories", "YOUR memories"},
		{"promote only durable", "durable"},
		{"search target first", "search the target"},
		{"do not touch ephemeral", "Do NOT delete"},
		{"forbidden: secrets", "tokens or secrets"},
		{"forbidden: sync all", "engram sync --all"},
		{"forbidden: other project", "enrolling a project other"},
		{"forbidden: promoting before merge", "promoting before the merge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(norm, tc.want) {
				t.Errorf("section missing %q", tc.want)
			}
		})
	}
}

// T-F2-9 (secrets): no token, URL of a server or secret-shaped value.
func TestEngramCloudFragment_NoSecrets(t *testing.T) {
	got := engramCloudText(t)
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`https?://`),
		regexp.MustCompile(`(?i)(token|bearer|password|api[_-]?key)\s*[:=]`),
		regexp.MustCompile(`\b[A-Za-z0-9_\-]{32,}\b`),
		regexp.MustCompile(`(?i)gentle-`),
	} {
		if m := re.FindString(got); m != "" {
			t.Errorf("section contains %q, matched by %s", m, re)
		}
	}
}

// T-F2-9 (size): at most the cap.
func TestEngramCloudFragment_SizeCap(t *testing.T) {
	if n := len(engramCloudText(t)); n > engramCloudMaxBytes {
		t.Errorf("section is %d bytes, cap is %d", n, engramCloudMaxBytes)
	}
}

// T-F2-11: nothing tied to OpenSpec or archive: it works in ODD, SDD and
// without OpenSpec.
func TestEngramCloudFragment_NotTiedToOpenSpec(t *testing.T) {
	low := strings.ToLower(engramCloudText(t))
	for _, banned := range []string{"openspec", "archive", "/opsx:", "propose", "apply"} {
		if strings.Contains(low, banned) {
			t.Errorf("section is tied to OpenSpec: contains %q", banned)
		}
	}
}

// T-F2-12 (content): same text for Claude and Codex, no hardcoded agent paths,
// and no hook named as the mechanism (DA-05: in Claude a hook cannot be the
// delivery mechanism).
func TestEngramCloudFragment_AgentNeutral(t *testing.T) {
	got := engramCloudText(t)
	for _, banned := range []string{"CLAUDE.md", "AGENTS.md", ".claude", ".codex", "~/", "Claude", "Codex", "hook"} {
		if strings.Contains(got, banned) {
			t.Errorf("section is not agent-neutral: contains %q", banned)
		}
	}
}

// The registry: the harness ID maps to its asset, and the section is owned.
func TestEngramCloud_RegisteredAsOwnedSection(t *testing.T) {
	if sectionFragments["engram-cloud"] != "engram-cloud.md" {
		t.Errorf("sectionFragments[engram-cloud] = %q", sectionFragments["engram-cloud"])
	}
	if !isOwnedSection("engram-cloud") || !HasSectionFragment("engram-cloud") {
		t.Error("engram-cloud must be an owned section with a registered fragment")
	}
}
