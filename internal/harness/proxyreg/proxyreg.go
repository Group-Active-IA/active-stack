// Package proxyreg registers the memory proxy (R-26, DA-13) in the MCP
// configuration of Claude Code (~/.claude.json) and Codex
// (~/.codex/config.toml): the existing Engram MCP entry is rewritten so the
// agent starts "active-stack memory proxy -- <original command and args>"
// instead of the Engram server.
//
// It is a library: nothing calls it yet. R-20 wires it into install, update and
// uninstall. The byte-level functions (RegisterClaude, RegisterCodex and their
// inverses) are pure; Register and Unregister add the file handling, in the
// same way as internal/harness/hooks (R-16).
//
// Governance ALTO: it edits files of the user's agents. Every write is preceded
// by a snapshot (if it fails, the file is not touched), only the Engram entry is
// edited, other MCP servers are never touched, invalid files are refused and
// left intact, and Unregister gives the original bytes back.
package proxyreg

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

// Command is the executable the proxy entry runs. It is resolved through PATH,
// with no absolute path and no version, like the hooks of R-16.
const Command = "active-stack"

// ServerName is the name of the Engram MCP entry the proxy replaces.
const ServerName = "engram"

// ErrEntryNotFound means the file has no Engram MCP entry to proxy. It is not
// a failure: there is nothing to register.
var ErrEntryNotFound = errors.New("no Engram MCP entry")

// Adapter is the minimal view of an agent adapter this package needs.
type Adapter interface {
	Agent() model.Agent
	// MCPConfigPath returns the file with the agent's MCP servers for home. An
	// empty string means the agent has none: it is skipped.
	MCPConfigPath(homeDir, serverName string) string
}

// AgentResult is the outcome for one agent.
type AgentResult struct {
	Agent   model.Agent
	File    string
	Changed bool
	Skipped bool // not Claude or Codex, no config path, or no Engram entry
	Reason  string
}

var snapshotCreate = func(snapshotDir string, paths []string) error {
	_, err := backup.NewSnapshotter().Create(snapshotDir, paths)
	return err
}

// SetSnapshotterCreate replaces the snapshot function (tests) and returns a
// function that restores the previous one.
func SetSnapshotterCreate(fn func(snapshotDir string, paths []string) error) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

// Register proxies the Engram entry of every Claude and Codex adapter.
func Register(home string, adapters []Adapter) ([]AgentResult, error) {
	return apply(home, adapters, true)
}

// Unregister gives each Engram entry its original command back.
func Unregister(home string, adapters []Adapter) ([]AgentResult, error) {
	return apply(home, adapters, false)
}

func apply(home string, adapters []Adapter, register bool) ([]AgentResult, error) {
	var out []AgentResult
	for _, a := range adapters {
		r, err := applyOne(home, a, register)
		out = append(out, r)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func applyOne(home string, a Adapter, register bool) (AgentResult, error) {
	res := AgentResult{Agent: a.Agent(), File: a.MCPConfigPath(home, ServerName)}
	if res.File == "" || (a.Agent() != model.AgentClaude && a.Agent() != model.AgentCodex) {
		res.Skipped = true
		return res, nil
	}
	wrap := func(err error) error { return fmt.Errorf("proxyreg: %s (agent %q): %w", res.File, a.Agent(), err) }
	raw, err := os.ReadFile(res.File)
	if os.IsNotExist(err) {
		res.Skipped, res.Reason = true, "no MCP configuration file"
		return res, nil
	}
	if err != nil {
		return res, wrap(err)
	}

	var next []byte
	var changed bool
	switch {
	case a.Agent() == model.AgentClaude && register:
		next, changed, err = RegisterClaude(raw, ServerName)
	case a.Agent() == model.AgentClaude:
		next, changed, err = UnregisterClaude(raw, ServerName)
	case register:
		var s string
		s, changed, err = RegisterCodex(string(raw), ServerName)
		next = []byte(s)
	default:
		var s string
		s, changed, err = UnregisterCodex(string(raw), ServerName)
		next = []byte(s)
	}
	if errors.Is(err, ErrEntryNotFound) {
		res.Skipped, res.Reason = true, "no Engram MCP entry"
		return res, nil
	}
	if err != nil {
		return res, wrap(err)
	}
	if !changed || bytes.Equal(next, raw) {
		return res, nil
	}
	dir := filepath.Join(home, ".active-stack", "backups", "memory-proxy-"+string(a.Agent()))
	if err := snapshotCreate(dir, []string{res.File}); err != nil {
		return res, wrap(fmt.Errorf("backup failed, the file was not touched: %w", err))
	}
	perm := fs.FileMode(0o644)
	if fi, err := os.Stat(res.File); err == nil {
		perm = fi.Mode().Perm()
	}
	w, err := filemerge.WriteFileAtomicPreservingParent(res.File, next, perm)
	if err != nil {
		return res, wrap(err)
	}
	res.Changed = w.Changed
	return res, nil
}
