package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/selfupdate"
)

// R-27. The startup check only warns; the binary changes only with an explicit
// `active-stack self-update` that the person confirms.

// updateCheckDisabled is true for CI and when the person turned the check off.
func updateCheckDisabled() bool {
	return os.Getenv("ACTIVE_STACK_NO_UPDATE_CHECK") != "" || os.Getenv("CI") != ""
}

// maybeNotifyUpdate warns on stderr when a newer release exists. It runs only
// for the interactive entry points (no arguments, install, update): never for
// hooks or the proxy, which agents start on every event.
func maybeNotifyUpdate(args []string, stderr io.Writer) {
	if updateCheckDisabled() {
		return
	}
	if len(args) > 0 && args[0] != "install" && args[0] != "update" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	tag := selfupdate.Check(context.Background(), home, buildVersion)
	if tag == "" {
		return
	}
	fmt.Fprintln(stderr, i18n.Tf(preferredLang(), "selfupdate.available", buildVersion, tag))
}

func runSelfUpdate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	lang := preferredLang()
	yes := false
	for _, a := range args {
		switch a {
		case "--yes", "-y":
			yes = true
		default:
			fmt.Fprintln(stderr, i18n.Tf(lang, "selfupdate.usage"))
			return 2
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	tag := selfupdate.Check(ctx, home, buildVersion)
	if tag == "" {
		fmt.Fprintln(stdout, i18n.Tf(lang, "selfupdate.current", buildVersion))
		return 0
	}
	if !yes {
		fmt.Fprint(stdout, i18n.Tf(lang, "selfupdate.confirm", buildVersion, tag))
		line, _ := bufio.NewReader(stdin).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(line)); a != "y" && a != "s" && a != "yes" && a != "si" && a != "sí" {
			fmt.Fprintln(stdout, i18n.Tf(lang, "selfupdate.cancelled"))
			return 0
		}
	}
	got, err := selfupdate.Apply(ctx, home, exe, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		fmt.Fprintln(stderr, i18n.Tf(lang, "selfupdate.failed", err))
		return 1
	}
	fmt.Fprintln(stdout, i18n.Tf(lang, "selfupdate.done", got))
	return 0
}
