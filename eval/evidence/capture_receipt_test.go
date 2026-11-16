package evidence

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

func TestReceiptFromCapture(t *testing.T) {
	at := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	output := []byte("redacted output\n")
	fingerprint := Fingerprint256(output)
	for _, source := range []string{capture.SourceHarness, capture.SourceAgent} {
		record := capture.Record{
			RunID: "run-1", Command: "test command", Source: source,
			Redactor: capture.SourceHarness, Client: "codex", Interface: "cli",
			System: "windows", Version: "test", ExitCode: 3,
			Coverage: capture.CoverageDemonstrated, SHA256: fingerprint.Value,
		}
		tree := Tree{Dir: "repo", Commit: "base"}
		receipt, err := ReceiptFromCapture(record, tree, at, []string{"R1"})
		if err != nil {
			t.Fatal(err)
		}
		if len(receipt.Missing()) != 0 || receipt.Output.Fingerprint == nil || *receipt.Output.Fingerprint != fingerprint ||
			receipt.Output.Capture.Source != source || *receipt.Result.ExitCode != 3 {
			t.Fatalf("incomplete capture mapping: %+v", receipt)
		}
		env := Env{Dir: tree.Dir, Commit: tree.Commit, Copies: map[string][]byte{record.RunID: output}}
		if got := EvaluateRequirement("R1", []Receipt{receipt}, env); got.Status != Unverified {
			t.Fatalf("receipt without separate corroboration: %+v", got)
		}
	}
}

func TestCapturedCommandReceiptReview(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	redactor, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		t.Fatal(err)
	}
	dir := gitRepo(t, "false")
	fingerprint, err := StateFingerprint(dir)
	if err != nil {
		t.Fatal(err)
	}
	head := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	rawHead, err := head.Output()
	if err != nil {
		t.Fatal(err)
	}
	tree := Tree{Dir: dir, Commit: strings.TrimSpace(string(rawHead)), Fingerprint: &fingerprint}
	for _, client := range []string{"claude", "codex"} {
		t.Run(client, func(t *testing.T) {
			cmd := exec.Command("git", "status", "--porcelain")
			cmd.Dir = dir
			record, err := capture.Run(capture.Spec{
				Command: "git status --porcelain", Cmd: cmd, RunDir: t.TempDir(), RunID: client,
				Redactor: redactor, Meta: capture.Meta{Client: client, Interface: "cli", System: runtime.GOOS, Version: "fixture"},
			})
			if err != nil {
				t.Fatal(err)
			}
			receipt, err := ReceiptFromCapture(record, tree, time.Now(), []string{"clean-tree"})
			if err != nil {
				t.Fatal(err)
			}
			output, err := os.ReadFile(record.CopyPath)
			if err != nil {
				t.Fatal(err)
			}
			if record.ExitCode != 0 || len(output) != 0 {
				t.Fatalf("expected clean checkout, got exit %d and %q", record.ExitCode, output)
			}
			env := Env{
				Dir: dir, Commit: tree.Commit, StateFP: &fingerprint, Copies: map[string][]byte{record.RunID: output},
				Corroboration: []CorroborationRecord{{Run: record.RunID, Command: record.Command, Dir: dir, Commit: tree.Commit, StateFP: &fingerprint, ExitCode: &record.ExitCode, Source: capture.SourceHarness, Verdict: Confirms}},
				Sufficiency:   []SufficiencyRecord{{Run: record.RunID, Requirement: "clean-tree", Command: record.Command, Verdict: Sufficient}},
			}
			if got := EvaluateRequirement("clean-tree", []Receipt{receipt}, env); got.Status != Verified {
				t.Fatalf("valid execution: %+v", got)
			}
			changed := Fingerprint256([]byte("different staged or unstaged bytes"))
			env.StateFP = &changed
			if got := EvaluateRequirement("clean-tree", []Receipt{receipt}, env); got.Status != Rejected {
				t.Fatalf("receipt for stale tree: %+v", got)
			}
			env.StateFP = &fingerprint
			env.Corroboration = nil
			if got := EvaluateRequirement("clean-tree", []Receipt{receipt}, env); got.Status != Unverified {
				t.Fatalf("uncorroborated execution: %+v", got)
			}
		})
	}
}
