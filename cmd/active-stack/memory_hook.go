package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/i18n"
	"github.com/Group-Active-IA/active-stack/internal/memory"
)

// Hooks of "active-stack memory hook <name>" ([EN] Fase 5, R15-1). They only
// add context; pre-save waits for the owner's decision between hook and proxy.
// Both Claude Code and Codex read the same output shape for these events.
var memoryHooks = map[string]struct {
	event   string
	context func(memory.Result) string
}{
	"session-start": {event: "SessionStart", context: memory.SessionContext},
	"pre-read":      {event: "PreToolUse", context: memory.ReadContext},
}

// Replaceable in tests.
var (
	memoryStdin  io.Reader = os.Stdin
	memoryDetect           = memory.Detect
	hookBudget             = 3 * time.Second
)

var errHookTimeout = errors.New("timed out")

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// runMemoryHook never blocks the agent (Q15): whatever happens it exits 0;
// problems go to stderr and nothing is printed on stdout.
func runMemoryHook(args []string, lang i18n.Lang, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, i18n.T(lang, "memory.usage"))
		return memoryExitOK
	}
	hook, ok := memoryHooks[args[0]]
	if !ok {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.hook.unknown", args[0]))
		return memoryExitOK
	}

	var event struct {
		HookEventName string `json:"hook_event_name"`
		Cwd           string `json:"cwd"`
	}
	if err := json.NewDecoder(memoryStdin).Decode(&event); err != nil {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.hook.skipped", args[0], err))
		return memoryExitOK
	}
	if event.HookEventName != hook.event {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.hook.skipped", args[0],
			fmt.Sprintf("event %q, expected %q", event.HookEventName, hook.event)))
		return memoryExitOK
	}
	dir := event.Cwd
	if dir == "" {
		if wd, err := os.Getwd(); err == nil {
			dir = wd
		}
	}

	res, err := detectWithin(dir, hookBudget)
	if err != nil {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.hook.skipped", args[0], err))
		return memoryExitOK
	}
	for _, w := range res.Warnings {
		switch w.Code {
		case memory.WarnMarkerUnknownPolicy:
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.warn.marker_unknown_policy", w.Path, w.Policy))
		default:
			fmt.Fprintln(stderr, i18n.Tf(lang, "memory.warn.marker_invalid", w.Path))
		}
	}
	line := hook.context(res)
	if line == "" {
		return memoryExitOK
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = hook.event
	out.HookSpecificOutput.AdditionalContext = line
	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		fmt.Fprintln(stderr, i18n.Tf(lang, "memory.hook.skipped", args[0], err))
	}
	return memoryExitOK
}

// detectWithin runs Detect but gives up after budget, so a slow Git never
// holds the agent. The abandoned Git commands end on their own timeout.
func detectWithin(dir string, budget time.Duration) (memory.Result, error) {
	type result struct {
		res memory.Result
		err error
	}
	done := make(chan result, 1)
	go func() {
		res, err := memoryDetect(dir)
		done <- result{res, err}
	}()
	select {
	case r := <-done:
		return r.res, r.err
	case <-time.After(budget):
		return memory.Result{}, errHookTimeout
	}
}
