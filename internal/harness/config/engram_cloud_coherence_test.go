package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// [EN] Fase 3 (R-14): coherence between the engram-cloud section, the Engram
// protocol and the orchestrator. The orchestrator is NOT edited: these tests
// pin that it stays complete without the harness and that the section names
// exactly what it overrides.

var allToggles = []string{"delegation", "model-routing", "engram", "tdd", "governance"}

// presentVariants lists the variant directories that really exist in the
// embedded assets (those with their own sdd-orchestrator.md).
func presentVariants(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(embeddedAssets, "assets")
	if err != nil {
		t.Fatalf("read assets: %v", err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := fs.Stat(embeddedAssets, "assets/"+e.Name()+"/sdd-orchestrator.md"); err == nil {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// T-F3-1: byte-for-byte regression. The eight variants each have a golden of
// Compose with every toggle (taken before the engram-cloud work, commit
// a899d2a and earlier), none is missing and the output carries no trace of the
// section: registering engram-cloud must not change the orchestrator.
func TestCoherence_EightVariantsHaveGoldenAndNoCloudTrace(t *testing.T) {
	variants := presentVariants(t)
	if len(variants) != 8 {
		t.Fatalf("variants = %v, want eight", variants)
	}
	if !equalStrings(variants, sortedKeys(knownVariants)) {
		t.Fatalf("knownVariants %v differs from the asset directories %v", sortedKeys(knownVariants), variants)
	}
	for _, v := range variants {
		t.Run(v, func(t *testing.T) {
			path := filepath.Join("testdata", "compose_all_toggles_"+v+".golden")
			golden, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("variant %q has no golden: %v", v, err)
			}
			got, err := Compose(allToggles, v)
			if err != nil {
				t.Fatalf("Compose(%s): %v", v, err)
			}
			if got != string(golden) {
				t.Errorf("Compose(%s) differs from its golden (len %d vs %d)", v, len(got), len(golden))
			}
			for _, trace := range []string{"active-stack:engram-cloud", "Cloud Team Memory", "cloud-branch"} {
				if strings.Contains(got, trace) {
					t.Errorf("Compose(%s) carries %q: the orchestrator must stay unedited", v, trace)
				}
			}
		})
	}
}

// gitSyncParts are the pieces that make the GIT SYNC section complete.
var gitSyncParts = []string{
	"### GIT SYNC — version project memory in its repo",
	"`engram sync`",
	"`engram sync --import`",
	"**NEVER `engram sync --all`**",
	"Before versioning `.engram/` in a PUBLIC repo",
	"git config alias.sync",
}

// T-F3-2: without the harness, GIT SYNC stays complete in every variant that
// delivers the Engram protocol, and in the layered Claude/Codex module.
func TestCoherence_GitSyncCompleteWithoutHarness(t *testing.T) {
	for _, v := range presentVariants(t) {
		t.Run(v, func(t *testing.T) {
			got, err := Compose(allToggles, v)
			if err != nil {
				t.Fatal(err)
			}
			for _, part := range gitSyncParts {
				if !strings.Contains(got, part) {
					t.Errorf("Compose(%s) lost GIT SYNC part %q", v, part)
				}
			}
		})
	}
	t.Run("layered engram-opsx module", func(t *testing.T) {
		layers, err := ComposeLayers(allToggles)
		if err != nil {
			t.Fatal(err)
		}
		mod := layers.Modules["engram-opsx"]
		for _, part := range gitSyncParts {
			if !strings.Contains(mod, part) {
				t.Errorf("layered engram-opsx module lost GIT SYNC part %q", part)
			}
		}
	})
	t.Run("without the engram toggle there is no GIT SYNC to override", func(t *testing.T) {
		got, err := Compose([]string{"delegation"}, "claude")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(got, "GIT SYNC") {
			t.Error("GIT SYNC must come only with the engram toggle")
		}
	})
}

// T-F3-3: the section declares what it overrides, and every one of those
// things really exists in the Engram protocol it extends.
func TestCoherence_SectionOverridesExistInProtocol(t *testing.T) {
	section, err := loadFragment("engram-cloud.md")
	if err != nil {
		t.Fatal(err)
	}
	protocol, err := loadFragment("engram-protocol.md")
	if err != nil {
		t.Fatal(err)
	}
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	section, protocol = norm(section), norm(protocol)

	if !strings.Contains(protocol, "## Engram Persistent Memory — Protocol") {
		t.Fatal("the protocol heading changed: the section refers to the Engram Protocol block")
	}
	if !strings.Contains(section, "extends the Engram Protocol block above and OVERRIDES:") {
		t.Fatal("the section must declare that it overrides parts of the protocol")
	}
	cases := []struct{ name, inSection, inProtocol string }{
		{"proactive saves", "proactive-save routing", "PROACTIVE SAVE TRIGGERS"},
		{"opsx topic keys", "`opsx/{change-name}/{phase}` keys", `topic_key: "opsx/{change-name}/{phase}"`},
		{"git sync", `the "GIT SYNC" section`, "### GIT SYNC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(section, tc.inSection) {
				t.Errorf("section no longer declares it overrides %q", tc.inSection)
			}
			if !strings.Contains(protocol, tc.inProtocol) {
				t.Errorf("protocol no longer has %q: the override points at nothing", tc.inProtocol)
			}
		})
	}
}

// reviewedVariants is the table of H5: every orchestrator variant, reviewed
// for how it briefs sub-agents about Engram. The cloud section does not edit
// the orchestrator because proxy B also tags saves made from sub-agents
// (R-05 V5/Q6; R-01 Q6 for Codex), so the briefing line is left as is.
// A NEW variant, or a changed briefing, fails this test until it is reviewed
// and the table is updated on purpose.
var reviewedVariants = map[string]bool{
	// variant -> carries the line that tells sub-agents to save progress to Engram.
	"antigravity": false,
	"claude":      true,
	"codex":       false,
	"cursor":      false,
	"gemini":      false,
	"generic":     false,
	"opencode":    true,
	"windsurf":    false,
}

const subagentBriefingLine = "**Engram instruction**: Tell the sub-agent to save progress to engram"

// T-F3-4: table test over the variants (H5).
func TestCoherence_ReviewedVariantsTable(t *testing.T) {
	present := presentVariants(t)
	if !equalStrings(present, sortedKeys(reviewedVariants)) {
		t.Fatalf("variants present %v differ from the reviewed table %v: review the new or removed variant (briefing of sub-agents about Engram) and update reviewedVariants", present, sortedKeys(reviewedVariants))
	}
	for _, v := range present {
		t.Run(v, func(t *testing.T) {
			data, err := fs.ReadFile(embeddedAssets, "assets/"+v+"/sdd-orchestrator.md")
			if err != nil {
				t.Fatal(err)
			}
			has := strings.Contains(string(data), subagentBriefingLine)
			if has != reviewedVariants[v] {
				t.Errorf("variant %q: briefing line present=%v, reviewed as %v: review it against the cloud section", v, has, reviewedVariants[v])
			}
		})
	}
}
