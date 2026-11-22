// Package autosync sets ENGRAM_CLOUD_AUTOSYNC=1 in the environment of the
// Engram MCP entry of Claude Code (~/.claude.json) and Codex
// (~/.codex/config.toml), and removes it again ([EN] Fase 9, T-F9-9, decision
// D9 of the Engram Cloud plan).
//
// What the stack configures is only that one variable. The server and the
// token are each person's own and are never written here.
//
// The strategy allows it only where Engram is registered as a stdio MCP with an
// environment table (Claude and Codex). The entry must already exist: this
// package never creates an MCP server, it only adds a variable to the one the
// stack (or the person) registered. OpenCode and Antigravity are skipped.
//
// Governance ALTO: it edits files of the user's agents. Every write is preceded
// by a snapshot (if it fails, the file is not touched), invalid files are
// refused and left intact, a value the person set to something other than "1"
// is never changed or removed, and Remove undoes Set byte for byte.
package autosync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/backup"
	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	"github.com/Group-Active-IA/active-stack/internal/jsonedit"
	"github.com/Group-Active-IA/active-stack/internal/model"
)

const (
	// EnvKey is the only variable this package writes.
	EnvKey = "ENGRAM_CLOUD_AUTOSYNC"
	// EnvValue turns the continuous background replication on.
	EnvValue = "1"
	// ServerName is the MCP entry the variable goes in.
	ServerName = "engram"
)

// Adapter is the minimal view of an agent adapter this package needs.
type Adapter interface {
	Agent() model.Agent
	// MCPConfigPath returns the file that holds the agent's MCP servers.
	MCPConfigPath(homeDir, serverName string) string
}

// AgentResult is the outcome for one agent.
type AgentResult struct {
	Agent    model.Agent
	File     string
	Changed  bool
	Skipped  bool
	Reason   string
	Warnings []string
}

// Result is the outcome of Set or Remove.
type Result struct{ Agents []AgentResult }

// AgentStatus reports whether the variable is set for one agent.
type AgentStatus struct {
	Agent   model.Agent
	File    string
	Enabled bool
}

var snapshotCreate = func(dir string, paths []string) error {
	_, err := backup.NewSnapshotter().Create(dir, paths)
	return err
}

// SetSnapshotterCreate replaces the snapshot function (tests) and returns a
// function that restores the previous one.
func SetSnapshotterCreate(fn func(dir string, paths []string) error) (restore func()) {
	old := snapshotCreate
	snapshotCreate = fn
	return func() { snapshotCreate = old }
}

func supported(a model.Agent) bool { return a == model.AgentClaude || a == model.AgentCodex }

// Set adds ENGRAM_CLOUD_AUTOSYNC=1 to the engram MCP entry of each supported
// agent whose file already declares one.
func Set(home string, adapters []Adapter) (Result, error) { return apply(home, adapters, true) }

// Remove deletes the variable again, only where its value is the one Set wrote.
func Remove(home string, adapters []Adapter) (Result, error) { return apply(home, adapters, false) }

// Status reports, without writing, whether the variable is set.
func Status(home string, adapters []Adapter) ([]AgentStatus, error) {
	var out []AgentStatus
	for _, a := range adapters {
		path := a.MCPConfigPath(home, ServerName)
		if path == "" || !supported(a.Agent()) {
			continue
		}
		raw, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			out = append(out, AgentStatus{Agent: a.Agent(), File: path})
			continue
		}
		if err != nil {
			return out, wrap(path, a.Agent(), err)
		}
		var val string
		var found bool
		if a.Agent() == model.AgentClaude {
			val, found, err = claudeValue(raw)
		} else {
			val, found, err = codexValue(string(raw))
		}
		if err != nil {
			return out, wrap(path, a.Agent(), err)
		}
		out = append(out, AgentStatus{Agent: a.Agent(), File: path, Enabled: found && val == EnvValue})
	}
	return out, nil
}

func wrap(path string, agent model.Agent, err error) error {
	return fmt.Errorf("autosync: %s (agent %q): %w", path, agent, err)
}

func apply(home string, adapters []Adapter, set bool) (Result, error) {
	var res Result
	for _, a := range adapters {
		r, err := applyOne(home, a, set)
		res.Agents = append(res.Agents, r)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func applyOne(home string, a Adapter, set bool) (AgentResult, error) {
	res := AgentResult{Agent: a.Agent(), File: a.MCPConfigPath(home, ServerName)}
	if res.File == "" || !supported(a.Agent()) {
		res.Skipped, res.Reason = true, "agent out of scope"
		return res, nil
	}
	raw, err := os.ReadFile(res.File)
	if os.IsNotExist(err) {
		res.Skipped, res.Reason = true, "no MCP configuration file"
		return res, nil
	}
	if err != nil {
		return res, wrap(res.File, a.Agent(), err)
	}

	var next []byte
	var outcome edit
	if a.Agent() == model.AgentClaude {
		next, outcome, err = claudeEdit(raw, set)
	} else {
		var s string
		s, outcome, err = codexEdit(string(raw), set)
		next = []byte(s)
	}
	if err != nil {
		return res, wrap(res.File, a.Agent(), err)
	}
	res.Warnings = outcome.warnings
	if outcome.noEntry {
		res.Skipped, res.Reason = true, "no engram MCP entry in the file"
		return res, nil
	}
	if bytes.Equal(next, raw) {
		return res, nil
	}
	dir := filepath.Join(home, ".active-stack", "backups", "engram-cloud-autosync-"+string(a.Agent()))
	if err := snapshotCreate(dir, []string{res.File}); err != nil {
		return res, wrap(res.File, a.Agent(), fmt.Errorf("backup failed, the file was not touched: %w", err))
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

// edit carries what an editor learned besides the new bytes.
type edit struct {
	noEntry  bool
	warnings []string
}

func userValueWarning(val string) string {
	return fmt.Sprintf("%s is already set to %q in the engram MCP entry: left as it is", EnvKey, val)
}

// ── Claude: ~/.claude.json, mcpServers.engram.env ──────────────────────────

type claudePos struct {
	root, engram jsonedit.Span
	rootMs       []jsonedit.Member
	engramMs     []jsonedit.Member
	envIdx       int // index in engramMs, -1 when absent
	env          jsonedit.Span
	envMs        []jsonedit.Member
	keyIdx       int // index in envMs, -1 when absent
	found        bool
}

func claudeLocate(src []byte) (claudePos, error) {
	p := claudePos{envIdx: -1, keyIdx: -1}
	var err error
	if p.root, err = jsonedit.RootSpan(src); err != nil {
		return p, err
	}
	if p.rootMs, err = jsonedit.ParseObject(src, p.root); err != nil {
		return p, err
	}
	i, ok := jsonedit.LastMember(p.rootMs, "mcpServers")
	if !ok {
		return p, nil
	}
	servers := p.rootMs[i].Val
	if src[servers.Start] != '{' {
		return p, jsonedit.ShapeErr(`"mcpServers" is not an object`)
	}
	serversMs, err := jsonedit.ParseObject(src, servers)
	if err != nil {
		return p, err
	}
	j, ok := jsonedit.LastMember(serversMs, ServerName)
	if !ok {
		return p, nil
	}
	p.engram = serversMs[j].Val
	if src[p.engram.Start] != '{' {
		return p, jsonedit.ShapeErr(`"mcpServers.%s" is not an object`, ServerName)
	}
	p.found = true
	if p.engramMs, err = jsonedit.ParseObject(src, p.engram); err != nil {
		return p, err
	}
	k, ok := jsonedit.LastMember(p.engramMs, "env")
	if !ok {
		return p, nil
	}
	p.envIdx, p.env = k, p.engramMs[k].Val
	if src[p.env.Start] != '{' {
		return p, jsonedit.ShapeErr(`"mcpServers.%s.env" is not an object`, ServerName)
	}
	if p.envMs, err = jsonedit.ParseObject(src, p.env); err != nil {
		return p, err
	}
	if m, ok := jsonedit.LastMember(p.envMs, EnvKey); ok {
		p.keyIdx = m
	}
	return p, nil
}

func memberString(src []byte, v jsonedit.Span) string {
	var s string
	if json.Unmarshal(src[v.Start:v.End], &s) != nil {
		return strings.TrimSpace(string(src[v.Start:v.End]))
	}
	return s
}

func claudeValue(src []byte) (string, bool, error) {
	p, err := claudeLocate(src)
	if err != nil || p.keyIdx < 0 {
		return "", false, err
	}
	return memberString(src, p.envMs[p.keyIdx].Val), true, nil
}

func claudeEdit(src []byte, set bool) ([]byte, edit, error) {
	p, err := claudeLocate(src)
	if err != nil {
		return nil, edit{}, err
	}
	if !p.found {
		return src, edit{noEntry: true}, nil
	}
	st := jsonedit.DetectStyle(src)
	if set {
		switch {
		case p.envIdx < 0:
			out, err := jsonedit.AddMember(src, p.engram, "env", map[string]string{EnvKey: EnvValue}, st)
			return out, edit{}, err
		case p.keyIdx < 0:
			out, err := jsonedit.AddMember(src, p.env, EnvKey, EnvValue, st)
			return out, edit{}, err
		}
		if v := memberString(src, p.envMs[p.keyIdx].Val); v != EnvValue {
			return src, edit{warnings: []string{userValueWarning(v)}}, nil
		}
		return src, edit{}, nil
	}
	if p.keyIdx < 0 || memberString(src, p.envMs[p.keyIdx].Val) != EnvValue {
		return src, edit{}, nil
	}
	if len(p.envMs) == 1 {
		out, err := jsonedit.RemoveMember(src, p.engram, p.envIdx)
		return out, edit{}, err
	}
	out, err := jsonedit.RemoveMember(src, p.env, p.keyIdx)
	return out, edit{}, err
}

// ── Codex: ~/.codex/config.toml, [mcp_servers.engram.env] ──────────────────

const (
	codexServerTable = "mcp_servers." + ServerName
	codexEnvTable    = codexServerTable + ".env"
)

func tomlHeader(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") && !strings.HasPrefix(t, "[[") {
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(t, "["), "]")), true
	}
	return "", false
}

func isKeyLine(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, EnvKey+" ") || strings.HasPrefix(t, EnvKey+"=")
}

func tomlString(line string) string {
	_, v, _ := strings.Cut(line, "=")
	return strings.Trim(strings.TrimSpace(v), `"'`)
}

func codexValue(content string) (string, bool, error) {
	if err := filemerge.ValidateTOML(content); err != nil {
		return "", false, err
	}
	in := false
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if name, ok := tomlHeader(line); ok {
			in = name == codexEnvTable
			continue
		}
		if in && isKeyLine(line) {
			return tomlString(line), true, nil
		}
	}
	return "", false, nil
}

func codexHasEngram(content string) bool {
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if name, ok := tomlHeader(line); ok && name == codexServerTable {
			return true
		}
	}
	return false
}

func codexEdit(content string, set bool) (string, edit, error) {
	if err := filemerge.ValidateTOML(content); err != nil {
		return content, edit{}, err
	}
	if !codexHasEngram(content) {
		return content, edit{noEntry: true}, nil
	}
	val, found, err := codexValue(content)
	if err != nil {
		return content, edit{}, err
	}
	if set {
		if found {
			if val != EnvValue {
				return content, edit{warnings: []string{userValueWarning(val)}}, nil
			}
			return content, edit{}, nil
		}
		out, err := filemerge.UpsertTOMLTableValues(content, codexEnvTable, map[string]string{EnvKey: `"` + EnvValue + `"`})
		return out, edit{}, err
	}
	if !found || val != EnvValue {
		return content, edit{}, nil
	}
	return codexRemoveKey(content), edit{}, nil
}

// codexRemoveKey drops our key line and, when that leaves the env table empty,
// the table header (and the blank line before it) too.
func codexRemoveKey(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var out []string
	in := false
	headerAt := -1
	for _, line := range lines {
		if name, ok := tomlHeader(line); ok {
			in = name == codexEnvTable
			if in {
				headerAt = len(out)
			}
			out = append(out, line)
			continue
		}
		if in && isKeyLine(line) {
			continue
		}
		out = append(out, line)
	}
	if headerAt >= 0 {
		empty := true
		for _, line := range out[headerAt+1:] {
			if _, ok := tomlHeader(line); ok {
				break
			}
			if t := strings.TrimSpace(line); t != "" {
				empty = false
				break
			}
		}
		if empty {
			end := headerAt + 1
			for end < len(out) {
				if _, ok := tomlHeader(out[end]); ok {
					break
				}
				end++
			}
			start := headerAt
			for start > 0 && strings.TrimSpace(out[start-1]) == "" {
				start--
			}
			out = append(out[:start], out[end:]...)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}
