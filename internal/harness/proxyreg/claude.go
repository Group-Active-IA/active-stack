package proxyreg

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/Group-Active-IA/active-stack/internal/jsonedit"
)

// Claude Code: the MCP servers live under the top-level "mcpServers" key of
// ~/.claude.json (or a project .mcp.json). The proxy replaces the server's
// command and args in place:
//
//	"engram": { "command": "engram", "args": ["mcp", "--tools=agent"] }
//	"engram": { "command": "active-stack",
//	            "args": ["memory", "proxy", "--", "engram", "mcp", "--tools=agent"] }
//
// Nothing else in the file is rewritten. JSON has no comments, so the original
// command and args travel inside the new args, as raw JSON text, which is what
// lets Unregister give back the original bytes (whitespace inside the args
// array included).

// prefix is what the proxy's args start with; the original command follows it.
var prefix = []string{"memory", "proxy", "--"}

const prefixText = `"memory", "proxy", "--", `

type claudeEntry struct {
	cmd  jsonedit.Span // value of "command"
	args jsonedit.Span // value of "args"
}

// locateClaude finds mcpServers.<server>.command and .args.
func locateClaude(src []byte, server string) (claudeEntry, error) {
	var e claudeEntry
	root, err := jsonedit.RootSpan(src)
	if err != nil {
		return e, err
	}
	rootMs, err := jsonedit.ParseObject(src, root)
	if err != nil {
		return e, err
	}
	i, ok := jsonedit.LastMember(rootMs, "mcpServers")
	if !ok {
		return e, ErrEntryNotFound
	}
	servers := rootMs[i].Val
	if src[servers.Start] != '{' {
		return e, jsonedit.ShapeErr(`"mcpServers" is not an object`)
	}
	serverMs, err := jsonedit.ParseObject(src, servers)
	if err != nil {
		return e, err
	}
	j, ok := jsonedit.LastMember(serverMs, server)
	if !ok {
		return e, ErrEntryNotFound
	}
	entry := serverMs[j].Val
	if src[entry.Start] != '{' {
		return e, jsonedit.ShapeErr(`"mcpServers.%s" is not an object`, server)
	}
	ms, err := jsonedit.ParseObject(src, entry)
	if err != nil {
		return e, err
	}
	c, okc := jsonedit.LastMember(ms, "command")
	a, oka := jsonedit.LastMember(ms, "args")
	if !okc || src[ms[c].Val.Start] != '"' {
		return e, jsonedit.ShapeErr(`"mcpServers.%s" has no string "command" (an HTTP or SSE server cannot be proxied)`, server)
	}
	if !oka || src[ms[a].Val.Start] != '[' {
		return e, jsonedit.ShapeErr(`"mcpServers.%s" has no "args" array`, server)
	}
	return claudeEntry{cmd: ms[c].Val, args: ms[a].Val}, nil
}

// isProxied reports whether the entry already runs through the proxy.
func isProxied(src []byte, e claudeEntry) bool {
	var cmd string
	if json.Unmarshal(src[e.cmd.Start:e.cmd.End], &cmd) != nil || cmd != Command {
		return false
	}
	elems, err := jsonedit.ParseArray(src, e.args)
	if err != nil || len(elems) < len(prefix)+1 {
		return false
	}
	for i, want := range prefix {
		var got string
		if json.Unmarshal(src[elems[i].Start:elems[i].End], &got) != nil || got != want {
			return false
		}
	}
	return true
}

// splice replaces the later span first so the earlier offsets stay valid.
func spliceTwo(src []byte, a jsonedit.Span, aText string, b jsonedit.Span, bText string) []byte {
	if a.Start < b.Start {
		a, b, aText, bText = b, a, bText, aText
	}
	out := make([]byte, 0, len(src)+len(aText)+len(bText))
	out = append(out, src[:b.Start]...)
	out = append(out, bText...)
	out = append(out, src[b.End:a.Start]...)
	out = append(out, aText...)
	out = append(out, src[a.End:]...)
	return out
}

// RegisterClaude points the server entry at the proxy. It returns the source
// unchanged (changed=false) when the entry already runs through it, and
// ErrEntryNotFound when the file has no such server.
func RegisterClaude(src []byte, server string) ([]byte, bool, error) {
	e, err := locateClaude(src, server)
	if err != nil {
		return nil, false, err
	}
	if isProxied(src, e) {
		return src, false, nil
	}
	cmdRaw := src[e.cmd.Start:e.cmd.End]
	inner := src[e.args.Start+1 : e.args.End-1]
	elems, err := jsonedit.ParseArray(src, e.args)
	if err != nil {
		return nil, false, err
	}
	var args []byte
	args = append(args, '[')
	if len(elems) == 0 {
		args = append(args, inner...) // whitespace of the empty array, kept
		args = append(args, prefixText...)
		args = append(args, cmdRaw...)
	} else {
		args = append(args, src[e.args.Start+1:elems[0].Start]...) // leading whitespace
		args = append(args, prefixText...)
		args = append(args, cmdRaw...)
		args = append(args, ", "...)
		args = append(args, src[elems[0].Start:elems[len(elems)-1].End]...)
		args = append(args, src[elems[len(elems)-1].End:e.args.End-1]...) // trailing whitespace
	}
	args = append(args, ']')
	out := spliceTwo(src, e.cmd, `"`+Command+`"`, e.args, string(args))
	if !json.Valid(out) {
		return nil, false, errors.New("proxyreg: the edit would leave invalid JSON; the file was not changed")
	}
	return out, true, nil
}

// UnregisterClaude gives the server entry back its original command and args.
// An entry that does not run through the proxy is left as is.
func UnregisterClaude(src []byte, server string) ([]byte, bool, error) {
	e, err := locateClaude(src, server)
	if errors.Is(err, ErrEntryNotFound) {
		return src, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !isProxied(src, e) {
		return src, false, nil
	}
	elems, err := jsonedit.ParseArray(src, e.args)
	if err != nil {
		return nil, false, err
	}
	n := len(prefix) // index of the original command
	cmdRaw := src[elems[n].Start:elems[n].End]
	var args []byte
	args = append(args, '[')
	if len(elems) == n+1 { // the original array was empty
		args = append(args, src[e.args.Start+1:elems[0].Start]...)
	} else {
		args = append(args, src[e.args.Start+1:elems[0].Start]...)
		args = append(args, src[elems[n+1].Start:elems[len(elems)-1].End]...)
		args = append(args, src[elems[len(elems)-1].End:e.args.End-1]...)
	}
	args = append(args, ']')
	out := spliceTwo(src, e.cmd, string(bytes.Clone(cmdRaw)), e.args, string(args))
	if !json.Valid(out) {
		return nil, false, errors.New("proxyreg: the edit would leave invalid JSON; the file was not changed")
	}
	return out, true, nil
}
