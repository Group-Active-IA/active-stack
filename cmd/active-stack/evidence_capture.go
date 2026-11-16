package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/eval/evidence"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/capture"
	"github.com/Group-Active-IA/active-stack/eval/taskcontext/redact"
	"github.com/Group-Active-IA/active-stack/internal/catalog"
)

const evidenceCaptureUsage = "usage: active-stack evidence-capture -client CLIENT -interface SURFACE -version VERSION -change CHANGE -requirement ID -- COMMAND [ARG...]"

func runEvidenceCapture(ctx context.Context, args []string, cacheDir string, out io.Writer) error {
	flags := flag.NewFlagSet("evidence-capture", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	client := flags.String("client", "", "assistant client")
	surface := flags.String("interface", "", "client surface")
	version := flags.String("version", "", "client version")
	change := flags.String("change", "", "OpenSpec change")
	requirement := flags.String("requirement", "", "requirement or scenario ID")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			_, err = fmt.Fprintln(out, evidenceCaptureUsage)
		}
		return err
	}
	command := flags.Args()
	if len(command) == 0 || (*client != "claude" && *client != "codex") || *surface == "" || *version == "" || *requirement == "" || *change == "" {
		return fmt.Errorf("%s", evidenceCaptureUsage)
	}
	before, err := checkoutEvidenceState(ctx)
	if err != nil {
		return err
	}
	dir := before.Dir
	cat, err := catalog.Load()
	if err != nil {
		return err
	}
	redactor, err := redact.New(cat.TaskContext.SensitivePatterns)
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp("", "active-stack-capture-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	var id [16]byte
	rand.Read(id[:])
	run := hex.EncodeToString(id[:])
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = dir
	argv, err := json.Marshal(command)
	if err != nil {
		return err
	}
	rec, err := capture.Run(capture.Spec{
		Command: string(argv), Cmd: cmd, RunDir: stage, RunID: run, Redactor: redactor,
		Meta: capture.Meta{Client: *client, Interface: *surface, System: runtime.GOOS, Version: *version},
	})
	if err != nil {
		return err
	}
	after, err := checkoutEvidenceState(ctx)
	if err != nil {
		return err
	}
	if before.Dir != after.Dir || before.Commit != after.Commit || *before.Fingerprint != *after.Fingerprint {
		return fmt.Errorf("checkout changed during capture: execution remains unverified")
	}
	copy, err := os.ReadFile(rec.CopyPath)
	if err != nil {
		return err
	}
	receipt, err := evidence.ReceiptFromCapture(rec, before, time.Now(), []string{*requirement})
	if err != nil {
		return err
	}
	store := capture.NewStore(cacheDir, time.Now)
	rec.CopyPath = filepath.Join(store.RunDir(*change, run), capture.CopyFile)
	path, err := store.Save(*change, run, rec, copy)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Receipt  evidence.Receipt `json:"receipt"`
		CopyPath string           `json:"copy_path"`
	}{Receipt: receipt, CopyPath: path})
}

func checkoutEvidenceState(ctx context.Context) (evidence.Tree, error) {
	root, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return evidence.Tree{}, fmt.Errorf("resolve checkout root: %w", err)
	}
	dir := strings.TrimSpace(string(root))
	before, err := evidence.ReceiptStateFingerprint(dir)
	if err != nil {
		return evidence.Tree{}, err
	}
	head := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	head.Dir = dir
	commit, err := head.Output()
	if err != nil {
		return evidence.Tree{}, err
	}
	return evidence.Tree{Dir: dir, Commit: strings.TrimSpace(string(commit)), Fingerprint: &before}, nil
}
