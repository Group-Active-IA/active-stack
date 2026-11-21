// Package hooks installs the memory read hooks of R-16 ([EN] Fase 6) into the
// user-level configuration of Claude Code (~/.claude/settings.json) and Codex
// (~/.codex/config.toml).
//
// Scope: only the read hooks of R-15 (session-start and pre-read). No save hook
// is ever installed here: how Claude tags saves (hook or proxy) is a pending
// owner decision.
//
// Governance ALTO: it edits files of the user's agents. Every write is preceded
// by a snapshot (if it fails, the file is not touched), foreign content is
// never rewritten, invalid files are refused and left intact, and Remove undoes
// Install byte for byte.
package hooks

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Command is the executable the hooks call. It is resolved through PATH, with
// no absolute path and no version, so updating the binary does not change the
// hook definition (Codex records trust against the hook's hash).
const Command = "active-stack"

// Marker lines that delimit our block in config.toml.
const (
	EntryMarkerBegin = "# active-stack:memory-hooks begin"
	EntryMarkerEnd   = "# active-stack:memory-hooks end"
)

// Adapter is the minimal view of an agent adapter this installer needs.
type Adapter interface {
	Agent() model.Agent
	// SettingsPath returns the agent's user-level settings file for home. An
	// empty string means the agent has none: it is skipped.
	SettingsPath(homeDir string) string
}

// Spec is one read hook to install.
type Spec struct {
	Event          string
	Matcher        string // empty: all sources
	Subcommand     string
	TimeoutSeconds int
}

// AgentResult is the outcome for one agent.
type AgentResult struct {
	Agent    model.Agent
	File     string
	Changed  bool
	Skipped  bool
	Warnings []string
}

// Result is the outcome of Install or Remove.
type Result struct {
	Agents  []AgentResult
	Notices []string
}

// AgentStatus is what is installed for one agent.
type AgentStatus struct {
	Agent        model.Agent
	File         string
	Installed    bool
	Missing      []string
	BinaryOnPath bool
}

// readMatcher covers the two read tools of Engram, under both names the tools
// can have (the standalone server and the plugin), anchored so that no other
// tool matches. In Claude Code and Codex a matcher with these characters is a
// regular expression. No save tool is in it.
const readMatcher = "^mcp__(plugin_engram_engram|engram)__(mem_search|mem_context)$"

// Specs returns the hooks this package installs: only the two read hooks.
func Specs() []Spec {
	return []Spec{
		{Event: "SessionStart", Subcommand: "session-start", TimeoutSeconds: 10},
		{Event: "PreToolUse", Matcher: readMatcher, Subcommand: "pre-read", TimeoutSeconds: 10},
	}
}

var (
	snapshotCreate = func(snapshotDir string, paths []string) error {
		_, err := backup.NewSnapshotter().Create(snapshotDir, paths)
		return err
	}
	lookPath = exec.LookPath
)

// SetSnapshotterCreate replaces the snapshot function (tests) and returns a
// function that restores the previous one.
func SetSnapshotterCreate(fn func(snapshotDir string, paths []string) error) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

// SetLookPath replaces the PATH lookup (tests) and returns a restore function.
func SetLookPath(fn func(file string) (string, error)) (restore func()) {
	old := lookPath
	lookPath = fn
	return func() { lookPath = old }
}

func supported(a model.Agent) bool { return a == model.AgentClaude || a == model.AgentCodex }

const (
	noticeCodexTrust = "Codex: open /hooks and review and trust the active-stack hook. Codex does not run it until you do, and it asks again whenever the hook definition changes."
	warnNotOnPath    = "active-stack is not in PATH: the hooks are installed but cannot run until it is (they never block Claude or Codex)"
)

// Install adds the read hooks to Claude and Codex. Any other agent is skipped.
// A missing active-stack in PATH is a warning, not a failure.
func Install(home string, adapters []Adapter) (Result, error) {
	var res Result
	codexSeen := false
	for _, a := range adapters {
		r, err := apply(home, a, true)
		res.Agents = append(res.Agents, r)
		if err != nil {
			return res, err
		}
		if a.Agent() == model.AgentCodex && (!r.Skipped || len(r.Warnings) > 0) {
			codexSeen = true
		}
	}
	if codexSeen {
		res.Notices = append(res.Notices, noticeCodexTrust)
	}
	return res, nil
}

// Remove deletes only the hooks Install added.
func Remove(home string, adapters []Adapter) (Result, error) {
	var res Result
	for _, a := range adapters {
		r, err := apply(home, a, false)
		res.Agents = append(res.Agents, r)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

// Status reports what is installed, without writing anything.
func Status(home string, adapters []Adapter) ([]AgentStatus, error) {
	_, pathErr := lookPath(Command)
	var out []AgentStatus
	for _, a := range adapters {
		path := a.SettingsPath(home)
		if path == "" || !supported(a.Agent()) {
			continue
		}
		raw, existed, err := readSettings(path)
		if err != nil {
			return out, wrap(path, a.Agent(), err)
		}
		var missing []string
		if !existed || len(bytes.TrimSpace(raw)) == 0 {
			for _, s := range Specs() {
				missing = append(missing, s.Subcommand)
			}
		} else if a.Agent() == model.AgentClaude {
			missing, err = claudeMissing(raw, true)
		} else {
			missing, err = codexMissing(string(raw))
		}
		if err != nil {
			return out, wrap(path, a.Agent(), err)
		}
		out = append(out, AgentStatus{
			Agent: a.Agent(), File: path, Installed: len(missing) == 0,
			Missing: missing, BinaryOnPath: pathErr == nil,
		})
	}
	return out, nil
}

func wrap(path string, agent model.Agent, err error) error {
	return fmt.Errorf("hooks: %s (agent %q): %w", path, agent, err)
}

func readSettings(path string) (raw []byte, existed bool, err error) {
	raw, err = os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

// apply installs or removes the hooks of one agent. Order matters: compute the
// new content first (an invalid file stops here, untouched), skip when nothing
// changes, take the snapshot, and only then write.
func apply(home string, a Adapter, install bool) (AgentResult, error) {
	res := AgentResult{Agent: a.Agent(), File: a.SettingsPath(home)}
	if res.File == "" || !supported(a.Agent()) {
		res.Skipped = true
		return res, nil
	}
	raw, existed, err := readSettings(res.File)
	if err != nil {
		return res, wrap(res.File, a.Agent(), err)
	}
	blank := !existed || len(bytes.TrimSpace(raw)) == 0
	if !install && blank {
		return res, nil
	}

	var next []byte
	switch a.Agent() {
	case model.AgentClaude:
		if install {
			next, err = claudeInstall(raw, existed)
		} else {
			next, err = claudeRemove(raw)
		}
	case model.AgentCodex:
		if install {
			if dup := hooksJSONDuplicate(res.File); dup != "" {
				res.Skipped = true
				res.Warnings = append(res.Warnings, dup)
				return res, nil
			}
			var s string
			s, err = codexInstall(string(raw))
			next = []byte(s)
		} else {
			var s string
			s, err = codexRemove(string(raw))
			next = []byte(s)
		}
	}
	if err != nil {
		return res, wrap(res.File, a.Agent(), err)
	}
	if install {
		if _, perr := lookPath(Command); perr != nil {
			res.Warnings = append(res.Warnings, warnNotOnPath)
		}
	}
	if existed && bytes.Equal(next, raw) {
		return res, nil
	}
	if existed {
		dir := filepath.Join(home, ".active-stack", "backups", "memory-hooks-"+string(a.Agent()))
		if err := snapshotCreate(dir, []string{res.File}); err != nil {
			return res, wrap(res.File, a.Agent(), fmt.Errorf("backup failed, the file was not touched: %w", err))
		}
	}
	perm := fs.FileMode(0o644)
	if fi, err := os.Stat(res.File); err == nil {
		perm = fi.Mode().Perm()
	}
	w, err := filemerge.WriteFileAtomicPreservingParent(res.File, next, perm)
	if err != nil {
		return res, wrap(res.File, a.Agent(), err)
	}
	res.Changed = w.Changed
	return res, nil
}

// hooksJSONDuplicate returns a warning when hooks.json next to config.toml
// already declares our hook: declaring it in both makes Codex run it twice.
func hooksJSONDuplicate(configPath string) string {
	path := filepath.Join(filepath.Dir(configPath), "hooks.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, s := range Specs() {
		if strings.Contains(string(b), Command+" memory hook "+s.Subcommand) {
			return fmt.Sprintf("%s already declares the active-stack hook: config.toml was left alone so Codex does not run it twice", path)
		}
	}
	return ""
}
