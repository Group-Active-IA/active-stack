package main

import (
	"fmt"
	"io"

	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// runUninstallDispatch handles the "uninstall" top-level subcommand.
//
// It parses the flags, validates them, and delegates to RunHeadlessUninstall.
// It is flat — there are no sub-subcommands under "uninstall" (D7): targeted/restore
// are flags, not subcommands.
//
// args is os.Args[2:] (the arguments after "uninstall").
//
// Returns the process exit code (0 = success, 1 = failure/usage error).
func runUninstallDispatch(args []string, cat uninstall.Catalog, reg uninstall.Registry, w io.Writer) int {
	flags, err := headless.ParseUninstallFlags(args)
	if err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		fmt.Fprintln(w, uninstallUsage)
		return 1
	}

	// Early validation: --strategy restore requires --restore-manifest.
	// Fail here with a clear error before touching any filesystem.
	if flags.Intent.Strategy == uninstall.StrategyRestore && flags.RestoreManifestPath == "" {
		fmt.Fprintln(w, "error: --strategy restore requires --restore-manifest <path>")
		fmt.Fprintln(w, uninstallUsage)
		return 1
	}

	return headless.RunHeadlessUninstall(flags, cat, reg, w)
}

const uninstallUsage = `Usage: jr-stack uninstall [flags]

Flags:
  --mode lite|full|custom      harness bundle to uninstall
  --agent claude,opencode,...  agents to target (default: all)
  --custom a,b,...             harness IDs (requires --mode custom)
  --strategy targeted|restore  reversal strategy (default: targeted)
  --restore-manifest <path>    backup manifest (required with --strategy restore)
  --dry-run                    print plan steps; do not execute
  --yes / -y                   confirm without prompt
  --home <path>                override home directory`
