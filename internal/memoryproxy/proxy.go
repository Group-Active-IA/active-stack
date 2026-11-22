// Package memoryproxy is the stdio MCP proxy that tags Engram saves with the
// current Git branch (DA-05, DA-13). It sits between the agent and the real
// Engram MCP server: it starts the server as a child process, copies what the
// server writes straight back to the agent, and forwards what the agent sends,
// rewriting only the arguments of the Engram write tools through
// memory.PrepareSave. Every other message goes through byte for byte, ids and
// line endings included.
//
// The proxy never blocks a save: if tagging fails or takes longer than the
// budget, the original message is forwarded and a line goes to stderr.
package memoryproxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/Group-Active-IA/active-stack/internal/memory"
)

// PrepareFunc tags the arguments of an Engram tool call made from dir. It has
// the signature of memory.PrepareSave.
type PrepareFunc func(dir, tool string, input map[string]any) (map[string]any, bool, error)

// DefaultBudget is how long tagging may take before the save goes untagged.
const DefaultBudget = 3 * time.Second

// Options configures Run.
type Options struct {
	Child   []string // the real Engram MCP server command and arguments
	Dir     string   // project directory used to detect the branch
	Prepare PrepareFunc
	Stdin   io.Reader // from the agent
	Stdout  io.Writer // to the agent
	Stderr  io.Writer
	Budget  time.Duration
}

// taggedTools are the Engram tools whose arguments are tagged: the ones the
// R-05 prototype covered (save and the narrative paths).
var taggedTools = map[string]bool{
	"mem_save":             true,
	"mem_session_summary":  true,
	"mem_session_finalize": true,
	"mem_capture_passive":  true,
	"mem_save_prompt":      true,
}

// Exit codes besides the child's own.
const (
	ExitStartFailed = 1 // the child could not be started
	ExitUsage       = 2 // no child command
)

type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

// Run starts the child and proxies until the agent closes its input or the
// child exits. It returns the child's exit code.
func Run(ctx context.Context, o Options) int {
	if o.Stderr == nil {
		o.Stderr = io.Discard
	}
	errw := &lockedWriter{w: o.Stderr}
	if len(o.Child) == 0 {
		fmt.Fprintln(errw, "memory proxy: no server command given (use: active-stack memory proxy -- <command> [args...])")
		return ExitUsage
	}
	if o.Prepare == nil {
		o.Prepare = memory.PrepareSave
	}
	if o.Budget <= 0 {
		o.Budget = DefaultBudget
	}

	cmd := exec.CommandContext(ctx, o.Child[0], o.Child[1:]...)
	cmd.Stdout = o.Stdout
	cmd.Stderr = errw
	childIn, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(errw, "memory proxy: %v\n", err)
		return ExitStartFailed
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(errw, "memory proxy: cannot start %q: %v\n", o.Child[0], err)
		return ExitStartFailed
	}

	go func() {
		defer childIn.Close() // the agent closed its side: tell the server
		r := bufio.NewReader(o.Stdin)
		for {
			line, rerr := r.ReadBytes('\n')
			if len(line) > 0 {
				if _, werr := childIn.Write(transform(line, &o, errw)); werr != nil {
					return // the child is gone; Wait reports its code
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		fmt.Fprintf(errw, "memory proxy: %v\n", err)
		return ExitStartFailed
	}
	return 0
}

// transform returns the bytes to forward for one line from the agent: the line
// itself unless it is a tools/call of a tagged tool that PrepareSave changes.
func transform(line []byte, o *Options, errw io.Writer) []byte {
	body, eol := splitEOL(line)
	if len(bytes.TrimSpace(body)) == 0 || bytes.TrimSpace(body)[0] != '{' {
		return line // blank, batch array, not JSON: untouched
	}
	var msg map[string]json.RawMessage
	if json.Unmarshal(body, &msg) != nil {
		return line
	}
	var method string
	if json.Unmarshal(msg["method"], &method) != nil || method != "tools/call" {
		return line
	}
	var params map[string]json.RawMessage
	if json.Unmarshal(msg["params"], &params) != nil {
		return line
	}
	var name string
	if json.Unmarshal(params["name"], &name) != nil || !taggedTools[memory.ToolBase(name)] {
		return line
	}
	var args map[string]any
	dec := json.NewDecoder(bytes.NewReader(params["arguments"]))
	dec.UseNumber() // keep large numbers exactly as they came
	if len(params["arguments"]) == 0 || dec.Decode(&args) != nil || args == nil {
		return line
	}

	tagged, changed, err := prepareWithin(o, name, args)
	if err != nil {
		fmt.Fprintf(errw, "memory proxy: tagging skipped for %s (%v). The save goes through untagged.\n", name, err)
		return line
	}
	if !changed {
		return line
	}
	rewritten, err := rebuild(msg, params, tagged)
	if err != nil {
		fmt.Fprintf(errw, "memory proxy: tagging skipped for %s (%v). The save goes through untagged.\n", name, err)
		return line
	}
	return append(rewritten, eol...)
}

func splitEOL(line []byte) (body, eol []byte) {
	switch {
	case bytes.HasSuffix(line, []byte("\r\n")):
		return line[:len(line)-2], line[len(line)-2:]
	case bytes.HasSuffix(line, []byte("\n")):
		return line[:len(line)-1], line[len(line)-1:]
	}
	return line, nil
}

var errBudget = errors.New("timed out")

// prepareWithin runs Prepare but gives up after the budget, so a slow Git never
// holds the agent's save. The abandoned work ends on its own timeouts.
func prepareWithin(o *Options, tool string, args map[string]any) (map[string]any, bool, error) {
	type result struct {
		out     map[string]any
		changed bool
		err     error
	}
	done := make(chan result, 1)
	go func() {
		out, changed, err := o.Prepare(o.Dir, tool, args)
		done <- result{out, changed, err}
	}()
	select {
	case r := <-done:
		return r.out, r.changed, r.err
	case <-time.After(o.Budget):
		return nil, false, errBudget
	}
}

// rebuild puts the tagged arguments back into the message. Only the arguments
// value changes; the rest of params and of the envelope (id included) keep
// their raw JSON.
func rebuild(msg, params map[string]json.RawMessage, args map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(args); err != nil {
		return nil, err
	}
	params["arguments"] = bytes.TrimRight(buf.Bytes(), "\n")
	rawParams, err := marshalRaw(params)
	if err != nil {
		return nil, err
	}
	msg["params"] = rawParams
	return marshalRaw(msg)
}

func marshalRaw(m map[string]json.RawMessage) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
