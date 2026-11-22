package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/memory"
)

// Exit codes of "active-stack memory".
const (
	memoryExitOK    = 0 // decided (tagged or not)
	memoryExitError = 1 // could not decide: missing directory, no git, empty slug
	memoryExitUsage = 2 // bad subcommand, flag, argument or language
)

// memoryBranchJSON is the stable --json schema. Absent values are null so the
// keys never change.
type memoryBranchJSON struct {
	Mode   string  `json:"mode"`
	Tagged bool    `json:"tagged"`
	Branch *string `json:"branch"`
	Slug   *string `json:"slug"`
}

// runMemoryDispatch handles "active-stack memory <subcommand>". args is
// os.Args[2:]. It returns the process exit code.
func runMemoryDispatch(args []string, stdout, stderr io.Writer) int {
	lang := preferredLang()
	if len(args) == 0 {
		fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
		return memoryExitUsage
	}
	if args[0] == "hook" {
		return runMemoryHook(args[1:], lang, stdout, stderr)
	}
	if args[0] == "proxy" {
		return runMemoryProxy(args[1:], lang, stdout, stderr)
	}
	if args[0] == "init" {
		return runMemoryInit(args[1:], lang, stdout, stderr)
	}
	if args[0] != "branch" {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.error.unknown_subcommand", args[0]))
		fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
		return memoryExitUsage
	}
	return runMemoryBranch(args[1:], lang, stdout, stderr)
}

func runMemoryBranch(args []string, lang i18n.Lang, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("memory branch", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	project := fs.String("project", ".", "")
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

	res, err := memory.Detect(*project)
	if err != nil {
		fmt.Fprintln(stderr, memoryErrorText(lang, err, *project, res))
		return memoryExitError
	}
	for _, w := range res.Warnings {
		switch w.Code {
		case memory.WarnMarkerUnknownPolicy:
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.warn.marker_unknown_policy", w.Path, w.Policy))
		default:
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.warn.marker_invalid", w.Path))
		}
	}

	if *asJSON {
		out := memoryBranchJSON{Mode: res.Mode, Tagged: res.Tagged}
		if res.Branch != "" {
			out.Branch = &res.Branch
		}
		if res.Tagged {
			out.Slug = &res.Slug
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.error.failed", err))
			return memoryExitError
		}
		return memoryExitOK
	}

	tagged, branch, slug := i18n.T(lang, "memory.value.no"), res.Branch, i18n.T(lang, "memory.value.none")
	if res.Tagged {
		tagged, slug = i18n.T(lang, "memory.value.yes"), res.Slug
	}
	if branch == "" {
		branch = i18n.T(lang, "memory.value.no_repo")
	}
	fmt.Fprintf(stdout, "%s: %s\n%s: %s\n%s: %s\n%s: %s\n",
		i18n.T(lang, "memory.label.mode"), res.Mode,
		i18n.T(lang, "memory.label.tagged"), tagged,
		i18n.T(lang, "memory.label.branch"), branch,
		i18n.T(lang, "memory.label.slug"), slug)
	return memoryExitOK
}

func memoryErrorText(lang i18n.Lang, err error, project string, res memory.Result) string {
	switch {
	case errors.Is(err, memory.ErrNoDirectory):
		return i18n.Tf(lang, "memory.error.no_directory", project)
	case errors.Is(err, memory.ErrGitMissing):
		return i18n.T(lang, "memory.error.git_missing")
	case errors.Is(err, memory.ErrEmptySlug):
		return i18n.Tf(lang, "memory.error.empty_slug", res.Branch)
	default:
		return i18n.Tf(lang, "memory.error.failed", err)
	}
}

// preferredLang is the language saved in ~/.active-stack/config.json, or the
// default when there is none.
func preferredLang() i18n.Lang {
	home, err := os.UserHomeDir()
	if err != nil {
		return i18n.Default
	}
	return i18n.LoadPreferred(home)
}
