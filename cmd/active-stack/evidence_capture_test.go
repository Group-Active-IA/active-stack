package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/evidence"
)

func TestEvidenceCaptureRequiresCommandAndMetadata(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"--", "git", "status"},
		{"-client", "codex", "-interface", "cli", "-version", "fixture", "-requirement", "R1", "--"},
	} {
		var out bytes.Buffer
		if err := runEvidenceCapture(context.Background(), args, t.TempDir(), &out); err == nil || !strings.Contains(err.Error(), "usage:") {
			t.Errorf("args %v: want usage error, got %v", args, err)
		}
		if out.Len() != 0 {
			t.Errorf("args %v: unexpected output %q", args, &out)
		}
	}
}

func TestEvidenceCapturePersistsReceiptOutput(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is unavailable")
	}
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"init", "-q"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "initial"}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cache := t.TempDir()
	cmd := exec.Command(exe, "-test.run=^TestEvidenceCaptureCLIHelper$", "--", "evidence-capture",
		"-client", "codex", "-interface", "cli", "-version", "fixture",
		"-change", "fixture", "-requirement", "R1", "--", "git", "rev-parse", "HEAD")
	cmd.Env = append(os.Environ(), "ACTIVE_STACK_EVIDENCE_CLI_TEST=1", "LOCALAPPDATA="+cache, "XDG_CACHE_HOME="+cache, "HOME="+cache)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI returned %v: %s", err, out)
	}
	var result struct {
		Receipt  evidence.Receipt `json:"receipt"`
		CopyPath string           `json:"copy_path"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(result.CopyPath)
	if err != nil {
		t.Fatal(err)
	}
	r := result.Receipt
	if len(r.Missing()) > 0 || *r.Result.ExitCode != 0 || evidence.Fingerprint256(body) != *r.Output.Fingerprint || strings.TrimSpace(string(body)) != r.Tree.Commit {
		t.Fatalf("unexpected capture: %+v", r)
	}

	stateCmd := exec.Command(exe, "-test.run=^TestEvidenceCaptureCLIHelper$", "--", "evidence-state")
	stateCmd.Env = cmd.Env
	stateOut, err := stateCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("evidence-state: %v: %s", err, stateOut)
	}
	var current evidence.Tree
	if err := json.Unmarshal(stateOut, &current); err != nil {
		t.Fatal(err)
	}
	if current.Dir != r.Tree.Dir || current.Commit != r.Tree.Commit || current.Fingerprint == nil || *current.Fingerprint != *r.Tree.Fingerprint {
		t.Fatalf("state does not match captured tree: %+v", current)
	}

	if err := os.WriteFile("unstaged.txt", []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	var staged bytes.Buffer
	err = runEvidenceCapture(context.Background(), []string{
		"-client", "claude", "-interface", "cli", "-version", "fixture",
		"-change", "fixture", "-requirement", "R1", "--", "git", "add", "unstaged.txt",
	}, t.TempDir(), &staged)
	if err == nil || !strings.Contains(err.Error(), "checkout changed during capture") || staged.Len() != 0 {
		t.Fatalf("index mutation must not produce a receipt: err=%v output=%q", err, &staged)
	}

	var rejected bytes.Buffer
	err = runEvidenceCapture(context.Background(), []string{
		"-client", "claude", "-interface", "cli", "-version", "fixture",
		"-change", "fixture", "-requirement", "R1", "--",
		"git", "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
		"-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "changed",
	}, t.TempDir(), &rejected)
	if err == nil || !strings.Contains(err.Error(), "checkout changed during capture") || rejected.Len() != 0 {
		t.Fatalf("mutation must not produce a receipt: err=%v output=%q", err, &rejected)
	}
}

func TestEvidenceCaptureCLIHelper(t *testing.T) {
	if os.Getenv("ACTIVE_STACK_EVIDENCE_CLI_TEST") != "1" {
		return
	}
	os.Args = append(os.Args[:1], os.Args[3:]...)
	main()
	os.Exit(0)
}

func TestEvidenceCaptureHelp(t *testing.T) {
	var out bytes.Buffer
	err := runEvidenceCapture(context.Background(), []string{"--help"}, "", &out)
	if err != nil || !strings.Contains(out.String(), "active-stack evidence-capture") {
		t.Fatalf("help: err=%v output=%q", err, &out)
	}
}
