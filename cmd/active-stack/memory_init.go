package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/memory/projectinit"
)

// memoryExitNeedsDecision is "active-stack memory init" stopping so as not to
// overwrite something of the person, or waiting for the project name to be
// confirmed. It is different from an error so that whoever invokes it (the
// foundation, R-19) can tell "failed" from "needs a human decision".
const memoryExitNeedsDecision = 3

var (
	// memoryInitStdin is where the confirmation is read from (tests replace it).
	memoryInitStdin io.Reader = os.Stdin
	// memoryInitInteractive says whether there is someone to ask (tests replace it).
	memoryInitInteractive = defaultInteractive
)

// defaultInteractive is true when stdin is a terminal.
func defaultInteractive() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// memoryInitJSON is the stable --json schema. Absent values are null so the
// keys never change.
type memoryInitJSON struct {
	Status         string               `json:"status"` // ok | needs_decision | error
	Reason         *string              `json:"reason"`
	Root           *string              `json:"root"`
	ProjectName    *string              `json:"project_name"`
	NameSource     *string              `json:"name_source"`
	Files          []memoryInitFileJSON `json:"files"`
	BackupDir      *string              `json:"backup_dir"`
	Warnings       []memoryInitWarnJSON `json:"warnings"`
	ProposedCommit *memoryInitCommit    `json:"proposed_commit"`
	Detail         map[string]string    `json:"detail"`
}

type memoryInitFileJSON struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

type memoryInitWarnJSON struct {
	Code  string   `json:"code"`
	Files []string `json:"files"`
}

type memoryInitCommit struct {
	Message  string   `json:"message"`
	Files    []string `json:"files"`
	Commands []string `json:"commands"`
}

func runMemoryInit(args []string, lang i18n.Lang, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("memory init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	project := fs.String("project", ".", "")
	name := fs.String("name", "", "")
	yes := fs.Bool("yes", false, "")
	asJSON := fs.Bool("json", false, "")
	langFlag := fs.String("lang", "", "")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
		fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
		return memoryExitUsage
	}
	if *langFlag != "" {
		parsed, err := i18n.Parse(*langFlag)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
			return memoryExitUsage
		}
		lang = parsed
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.init.error.failed", errors.New("cannot find the home directory")))
		return memoryExitError
	}
	opts := projectinit.Options{Dir: *project, Name: *name, Yes: *yes, Home: home}
	if !*yes && *name == "" && memoryInitInteractive() {
		opts.Confirm = func(derived string) (bool, error) {
			fmt.Fprint(stderr, i18n.Tf(lang, "memory.init.prompt", derived))
			line, _ := bufio.NewReader(memoryInitStdin).ReadString('\n')
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "y", "yes", "s", "si", "sí":
				return true, nil
			}
			return false, nil
		}
	}

	res, err := projectinit.Init(opts)
	status, reason, detail, code := classifyInitError(err)
	if err != nil {
		fmt.Fprintln(stderr, initErrorText(lang, err, reason, res))
	}
	if *asJSON {
		if encErr := json.NewEncoder(stdout).Encode(initJSON(res, status, reason, detail)); encErr != nil {
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.init.error.failed", encErr))
			return memoryExitError
		}
		return code
	}
	if err == nil {
		printInitResult(lang, stdout, res)
	}
	return code
}

// classifyInitError maps an error of projectinit to the JSON status, a stable
// reason, its detail and the exit code.
func classifyInitError(err error) (status, reason string, detail map[string]string, code int) {
	if err == nil {
		return "ok", "", nil, memoryExitOK
	}
	var ce *projectinit.ConflictError
	var nc *projectinit.NeedsConfirmationError
	var ie *projectinit.InvalidFileError
	var be *projectinit.BackupError
	switch {
	case errors.As(err, &ce):
		detail = map[string]string{"file": ce.File, "key": ce.Key, "have": ce.Have, "want": ce.Want}
		return "needs_decision", "conflict_" + ce.Reason, detail, memoryExitNeedsDecision
	case errors.As(err, &nc):
		reason = "name_not_confirmed"
		if nc.Declined {
			reason = "name_declined"
		}
		return "needs_decision", reason, map[string]string{"derived_name": nc.Derived, "source": nc.Source}, memoryExitNeedsDecision
	case errors.As(err, &ie):
		return "error", "invalid_file", map[string]string{"file": ie.File, "error": ie.Err.Error()}, memoryExitError
	case errors.As(err, &be):
		return "error", "backup_failed", map[string]string{"error": be.Err.Error()}, memoryExitError
	case errors.Is(err, projectinit.ErrHomeDirectory):
		return "error", "home_directory", nil, memoryExitError
	case errors.Is(err, projectinit.ErrNotGitRepo):
		return "error", "not_git_repo", nil, memoryExitError
	case errors.Is(err, projectinit.ErrEmptyName):
		return "error", "empty_name", nil, memoryExitError
	case errors.Is(err, projectinit.ErrGitMissing):
		return "error", "git_missing", nil, memoryExitError
	}
	return "error", "failed", map[string]string{"error": err.Error()}, memoryExitError
}

func initErrorText(lang i18n.Lang, err error, reason string, res projectinit.Result) string {
	var ce *projectinit.ConflictError
	var nc *projectinit.NeedsConfirmationError
	var ie *projectinit.InvalidFileError
	switch {
	case errors.As(err, &ce):
		switch ce.Reason {
		case projectinit.ReasonIgnoredDir:
			return i18n.Tf(lang, "memory.init.error.ignored_dir", ce.Have)
		case projectinit.ReasonStillIgnored:
			return i18n.Tf(lang, "memory.init.error.still_ignored", ce.Have)
		case projectinit.ReasonNotIgnored:
			return i18n.Tf(lang, "memory.init.error.not_ignored", ce.Have)
		}
		if ce.Key == "project_name" {
			return i18n.Tf(lang, "memory.init.error.conflict_name", ce.File, ce.Have, ce.Want)
		}
		return i18n.Tf(lang, "memory.init.error.conflict_policy", ce.File, ce.Key, ce.Have, ce.Want)
	case errors.As(err, &nc):
		if nc.Declined {
			return i18n.T(lang, "memory.init.error.declined")
		}
		return i18n.Tf(lang, "memory.init.error.not_confirmed", nc.Derived, i18n.T(lang, "memory.init.source."+nc.Source))
	case errors.As(err, &ie):
		return i18n.Tf(lang, "memory.init.error.invalid_file", ie.File, ie.Err)
	case errors.Is(err, projectinit.ErrHomeDirectory):
		return i18n.Tf(lang, "memory.init.error.home", res.Root)
	case errors.Is(err, projectinit.ErrNotGitRepo):
		return i18n.T(lang, "memory.init.error.not_git_repo")
	case errors.Is(err, projectinit.ErrEmptyName):
		return i18n.T(lang, "memory.init.error.empty_name")
	case errors.Is(err, projectinit.ErrGitMissing):
		return i18n.T(lang, "memory.error.git_missing")
	}
	return i18n.Tf(lang, "memory.init.error.failed", err)
}

func initJSON(res projectinit.Result, status, reason string, detail map[string]string) memoryInitJSON {
	out := memoryInitJSON{Status: status, Files: []memoryInitFileJSON{}, Warnings: []memoryInitWarnJSON{}, Detail: detail}
	str := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	out.Reason = str(reason)
	out.Root = str(res.Root)
	out.ProjectName = str(res.ProjectName)
	out.NameSource = str(res.NameSource)
	out.BackupDir = str(res.BackupDir)
	for _, f := range res.Files {
		out.Files = append(out.Files, memoryInitFileJSON{Path: f.Path, Action: string(f.Action)})
	}
	for _, w := range res.Warnings {
		out.Warnings = append(out.Warnings, memoryInitWarnJSON{Code: w.Code, Files: w.Files})
	}
	if res.Proposal != nil {
		out.ProposedCommit = &memoryInitCommit{Message: res.Proposal.Message, Files: res.Proposal.Files, Commands: res.Proposal.Commands}
	}
	return out
}

func printInitResult(lang i18n.Lang, w io.Writer, res projectinit.Result) {
	fmt.Fprintln(w, i18n.Tf(lang, "memory.init.project", res.Root))
	fmt.Fprintln(w, i18n.Tf(lang, "memory.init.name", res.ProjectName, i18n.T(lang, "memory.init.source."+res.NameSource)))
	for _, f := range res.Files {
		fmt.Fprintln(w, i18n.Tf(lang, "memory.init.file", f.Path, i18n.T(lang, "memory.init.action."+string(f.Action))))
	}
	if res.BackupDir != "" {
		fmt.Fprintln(w, i18n.Tf(lang, "memory.init.backup", res.BackupDir))
	}
	for _, wn := range res.Warnings {
		switch wn.Code {
		case projectinit.WarnDirtyFiles:
			fmt.Fprintln(w, i18n.Tf(lang, "memory.init.warn.dirty", strings.Join(wn.Files, ", ")))
		case projectinit.WarnTrackedOther:
			fmt.Fprintln(w, i18n.Tf(lang, "memory.init.warn.tracked", strings.Join(wn.Files, ", ")))
		}
	}
	if res.Proposal == nil {
		fmt.Fprintln(w, i18n.T(lang, "memory.init.commit.none"))
		return
	}
	fmt.Fprintln(w, i18n.Tf(lang, "memory.init.commit.header", res.Root))
	for _, c := range res.Proposal.Commands {
		fmt.Fprintln(w, "  "+c)
	}
}
