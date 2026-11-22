package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/memoryproxy"
)

// runMemoryProxy is "active-stack memory proxy [--project <dir>] -- <cmd> [args...]":
// the stdio MCP proxy that tags Engram saves by branch (DA-13). Everything
// after "--" is the real Engram MCP server command. The exit code is the
// server's own; usage errors give memoryExitUsage.
func runMemoryProxy(args []string, lang i18n.Lang, stdout, stderr io.Writer) int {
	var flags, child []string
	for i, a := range args {
		if a == "--" {
			flags, child = args[:i], args[i+1:]
			break
		}
		flags = args
	}
	fs := flag.NewFlagSet("memory proxy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	project := fs.String("project", "", "")
	if err := fs.Parse(flags); err != nil || fs.NArg() != 0 {
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
		fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
		return memoryExitUsage
	}
	if len(child) == 0 {
		fmt.Fprintln(stderr, i18n.T(lang, "memory.proxy.no_command"))
		return memoryExitUsage
	}
	dir := *project
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.error.failed", err))
			return memoryExitError
		}
		dir = wd
	}
	return memoryproxy.Run(context.Background(), memoryproxy.Options{
		Child: child, Dir: dir, Stdin: memoryStdin, Stdout: stdout, Stderr: stderr,
	})
}
