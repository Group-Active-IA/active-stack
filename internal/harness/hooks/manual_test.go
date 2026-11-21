package hooks_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/agents/claude"
	"github.com/Group-Active-IA/active-stack/internal/agents/codex"
	"github.com/Group-Active-IA/active-stack/internal/harness/hooks"
)

// TestManual drives the installer with the REAL adapters, the REAL snapshotter
// and the REAL PATH lookup, against a directory the person chooses. It exists
// for the manual check with a real Claude Code and a real Codex (ficha R-16,
// «Verificación con agentes reales»), because the installer is not wired to the
// `active-stack install` command yet (R-20).
//
// It does nothing unless AS_HOOKS_HOME is set, so `go test ./...` skips it.
//
//	AS_HOOKS_HOME    directory used as HOME (it must exist)
//	AS_HOOKS_ACTION  install | remove | status
//	AS_HOOKS_AGENTS  claude | codex | claude,codex (default: both)
func TestManual(t *testing.T) {
	home := os.Getenv("AS_HOOKS_HOME")
	if home == "" {
		t.Skip("manual check: set AS_HOOKS_HOME and AS_HOOKS_ACTION to run it")
	}
	if fi, err := os.Stat(home); err != nil || !fi.IsDir() {
		t.Fatalf("AS_HOOKS_HOME %q is not an existing directory", home)
	}
	var adapters []hooks.Adapter
	agents := os.Getenv("AS_HOOKS_AGENTS")
	if agents == "" {
		agents = "claude,codex"
	}
	for _, a := range strings.Split(agents, ",") {
		switch strings.TrimSpace(a) {
		case "claude":
			adapters = append(adapters, claude.NewAdapter())
		case "codex":
			adapters = append(adapters, codex.NewAdapter())
		default:
			t.Fatalf("unknown agent %q in AS_HOOKS_AGENTS", a)
		}
	}

	switch action := os.Getenv("AS_HOOKS_ACTION"); action {
	case "install", "remove":
		var (
			res hooks.Result
			err error
		)
		if action == "install" {
			res, err = hooks.Install(home, adapters)
		} else {
			res, err = hooks.Remove(home, adapters)
		}
		for _, r := range res.Agents {
			t.Logf("%s: file=%s changed=%v skipped=%v", r.Agent, r.File, r.Changed, r.Skipped)
			for _, w := range r.Warnings {
				t.Logf("  warning: %s", w)
			}
		}
		for _, n := range res.Notices {
			t.Logf("notice: %s", n)
		}
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	case "status":
		st, err := hooks.Status(home, adapters)
		for _, s := range st {
			t.Logf("%s: file=%s installed=%v missing=%v active-stack-in-PATH=%v", s.Agent, s.File, s.Installed, s.Missing, s.BinaryOnPath)
		}
		if err != nil {
			t.Fatalf("status: %v", err)
		}
	default:
		t.Fatalf("AS_HOOKS_ACTION must be install, remove or status, got %q", action)
	}
}
