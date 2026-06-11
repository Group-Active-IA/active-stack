package main

import (
	"fmt"
	"io"

	"github.com/Group-Active-IA/active-stack/assets"
	"github.com/Group-Active-IA/active-stack/cmd/active-stack/headless"
	"github.com/Group-Active-IA/active-stack/internal/install"
	"github.com/Group-Active-IA/active-stack/internal/verify"
)

// runStarterDispatch handles the "starter" top-level subcommand by dispatching
// on the first argument (the starter sub-subcommand: currently only "add").
//
// args is os.Args[2:] (the arguments after "starter").
//
// Returns the process exit code (0 = success, 1 = failure/usage error).
func runStarterDispatch(
	args []string,
	cat starterCatalog,
	reg install.Registry,
	w io.Writer,
) int {
	if len(args) == 0 {
		fmt.Fprintf(w, "Usage: active-stack starter <subcommand>\n")
		fmt.Fprintf(w, "Subcommands:\n")
		fmt.Fprintf(w, "  add <starter-id>   Apply a starter to a project\n")
		return 1
	}

	switch args[0] {
	case "add":
		return runStarterAddDispatch(args[1:], cat, reg, w)
	default:
		fmt.Fprintf(w, "error: unknown starter subcommand %q\n", args[0])
		fmt.Fprintf(w, "Usage: active-stack starter <subcommand>\n")
		fmt.Fprintf(w, "Subcommands:\n")
		fmt.Fprintf(w, "  add <starter-id>   Apply a starter to a project\n")
		return 1
	}
}

// runStarterAddDispatch parses the flags for "starter add" and delegates to
// runStarterAdd. Separated from runStarterDispatch so each level stays flat and
// testable.
func runStarterAddDispatch(
	args []string,
	cat starterCatalog,
	reg install.Registry,
	w io.Writer,
) int {
	flags, err := headless.ParseStarterAddFlags(args)
	if err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		fmt.Fprintf(w, "Usage: active-stack starter add <starter-id> [--project <path>] [--dry-run] [--yes] [--agent <csv>]\n")
		return 1
	}

	// Validate and absolutize the project root (D2: fail if doesn't exist; never create).
	projectRoot, err := headless.ResolveProjectRoot(flags.ProjectPath)
	if err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		return 1
	}
	flags.ProjectPath = projectRoot

	// Wire the embedded SkillsFS and a no-op verify hook (real hook wired via adapters).
	buildPlanFn := func(c install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error) {
		opts = install.WithEmbeddedSkillsFS(opts, assets.SkillsFS)
		if opts.VerifyHook == nil {
			opts.VerifyHook = verify.BuildHook(nil, nil, opts.HomeDir)
		}
		return install.BuildPlan(c, intent, opts)
	}

	return runStarterAdd(flags, cat, reg, buildPlanFn, w)
}
