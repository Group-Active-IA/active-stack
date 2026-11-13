package config

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

// Digests come from the literal module blocks of change 05 at e4ec0f5,
// entrega-03-04.md section 2, including the terminating newline.
func TestCodexContextModulesMatchSource05(t *testing.T) {
	want := map[string]string{
		"directed-code-reading":   "db1edbe36ee928b1dc70c095b2df7d55b35b6d19fcd2ccb557580aa821321cbf",
		"bounded-command-output":  "fcc5b089881d48ad6e466d250bb9825aa6e5bc81f06b8d98f9e9e5290494c0e9",
		"task-tool-selection":     "052215169c55cadc2310c1561118b8b6c0f7f818f3213b30ae82ff6181c715bd",
		"concise-agent-reporting": "713cf2c124658624c31788cd8e06d9e2b3aae40d2a527be058b7c34e7bdf71f7",
	}
	for _, toggles := range [][]string{nil, {"delegation", "tdd", "engram", "governance", "model-routing"}} {
		pkg, err := ComposeCodexPackage(toggles)
		if err != nil {
			t.Fatal(err)
		}
		for id, hash := range want {
			t.Run(id+"/"+strings.Join(toggles, "+"), func(t *testing.T) {
				body := pkg.Files["modules/"+id+".md"]
				if !strings.HasPrefix(body, "## ") {
					t.Fatalf("module %s missing or exposed as a standalone skill", id)
				}
				if got := fmt.Sprintf("%x", sha256.Sum256([]byte(body))); got != hash {
					t.Fatalf("module %s differs from pinned 05 source: %s", id, got)
				}
			})
		}
	}
}

func TestCodexActionModulesLoad05AndHandleUnavailableModules(t *testing.T) {
	for _, toggles := range [][]string{nil, {"delegation", "tdd", "engram"}} {
		pkg, err := ComposeCodexPackage(toggles)
		if err != nil {
			t.Fatal(err)
		}
		skill := pkg.Files["SKILL.md"]
		for _, text := range []string{"## Action to Modules", "| explore |", "| propose |", "| apply |", "| verify |", "| archive |", "| delegate (any action) |", "| session start and close |", "modules/bounded-command-output.md", "modules/directed-code-reading.md", "modules/task-tool-selection.md", "modules/concise-agent-reporting.md", "zero bytes or only", "missing module: <id> — <action> stopped", "missing optional module: <id> — continuing without it", "After compaction or resume", "Invoke the skill again", "e4ec0f54536760f7fca6692e27bafcb8cd380df1"} {
			if !strings.Contains(skill, text) {
				t.Errorf("missing action-module contract %q", text)
			}
		}
		if strings.Contains(skill, "model-assignments.md") || strings.Contains(skill, "${CLAUDE_SKILL_DIR}") {
			t.Fatal("Codex action table must not reference Claude-only module paths")
		}
	}
}

func TestCodexPartialEvidenceRequirements06StayExplicitlyIncomplete(t *testing.T) {
	for _, toggles := range [][]string{nil, {"delegation"}} {
		pkg, err := ComposeCodexPackage(toggles)
		if err != nil {
			t.Fatal(err)
		}
		skill := pkg.Files["SKILL.md"]
		for _, text := range []string{"Partial evidence requirements", "f62cf00c09dc8a3aa96ddb4648801250e1a67a8c", "not a complete receipt contract", "direct and delegated", "Integrity, provenance and sufficiency", "same HEAD with different diffs", "staged, unstaged and new untracked files", "matching hash alone", "LF and CRLF", "unknown or unassessable", "all three checks pass for every declared requirement", "at least one execution receipt", "stale and rejected", "stop only the action requiring evidence"} {
			if !strings.Contains(skill, text) {
				t.Errorf("missing partial evidence rule %q", text)
			}
		}
		if _, ok := pkg.Files["modules/evidence.md"]; ok {
			t.Fatal("undelivered 06 evidence module must not be fabricated")
		}
	}
}
