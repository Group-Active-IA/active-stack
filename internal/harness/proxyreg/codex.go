package proxyreg

import (
	"fmt"
	"reflect"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/Group-Active-IA/active-stack/internal/filemerge"
)

// Codex: ~/.codex/config.toml, table [mcp_servers.<server>]. TOML does allow
// comments, so the original table is kept commented out inside a marked block
// right where it was, followed by the proxied table:
//
//	# active-stack:memory-proxy begin
//	# Managed by active-stack: ...
//	#orig [mcp_servers.engram]
//	#orig command = "engram"
//	#orig args = ["mcp"]
//	[mcp_servers.engram]
//	command = 'active-stack'
//	args = ['memory', 'proxy', '--', 'engram', 'mcp']
//	# active-stack:memory-proxy end
//
// Unregister turns the "#orig " lines back into the original lines, so the
// file returns to its original bytes.

const (
	MarkerBegin = "# active-stack:memory-proxy begin"
	MarkerEnd   = "# active-stack:memory-proxy end"

	origPrefix      = "#orig "
	origNoEOLPrefix = "#orig-noeol " // the original table was the last line of the file, with no line break
	managedLine     = "# Managed by active-stack: removing the proxy restores the commented original below."
)

type tomlLine struct {
	text string // without terminator
	eol  string // "\n", "\r\n" or "" (last line)
}

func splitLines(content string) []tomlLine {
	var lines []tomlLine
	for len(content) > 0 {
		i := strings.IndexByte(content, '\n')
		if i < 0 {
			lines = append(lines, tomlLine{text: content})
			break
		}
		text, eol := content[:i], "\n"
		if strings.HasSuffix(text, "\r") {
			text, eol = text[:len(text)-1], "\r\n"
		}
		lines = append(lines, tomlLine{text, eol})
		content = content[i+1:]
	}
	return lines
}

func joinLines(lines []tomlLine) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.text)
		b.WriteString(l.eol)
	}
	return b.String()
}

func isHeader(text string) bool { return strings.HasPrefix(strings.TrimSpace(text), "[") }

// isServerHeader reports whether the line is exactly [mcp_servers.<server>].
func isServerHeader(text, server string) bool {
	t := strings.TrimSpace(text)
	if i := strings.IndexByte(t, ']'); i >= 0 {
		if rest := strings.TrimSpace(t[i+1:]); rest == "" || strings.HasPrefix(rest, "#") {
			t = t[:i+1] // a trailing comment is not part of the header
		}
	}
	return t == "[mcp_servers."+server+"]" || t == `[mcp_servers."`+server+`"]` || t == `[mcp_servers.'`+server+`']`
}

// findBlock returns the index of the begin and end marker lines.
func findBlock(lines []tomlLine) (begin, end int, found bool, err error) {
	begin, end = -1, -1
	nb, ne := 0, 0
	for i, l := range lines {
		switch strings.TrimSpace(l.text) {
		case MarkerBegin:
			begin, nb = i, nb+1
		case MarkerEnd:
			end, ne = i, ne+1
		}
	}
	switch {
	case nb == 0 && ne == 0:
		return 0, 0, false, nil
	case nb != 1 || ne != 1:
		return 0, 0, false, fmt.Errorf("found %d begin and %d end marker lines, want exactly one of each", nb, ne)
	case end < begin:
		return 0, 0, false, fmt.Errorf("the end marker comes before the begin marker")
	}
	return begin, end, true, nil
}

func lineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func tomlString(s string) string {
	if !strings.Contains(s, "'") && !strings.ContainsAny(s, "\r\n") {
		return "'" + s + "'"
	}
	b, _ := toml.Marshal(map[string]string{"v": s})
	return strings.TrimSpace(strings.TrimPrefix(string(b), "v = "))
}

// serverTable reads mcp_servers.<server> from the TOML text.
func serverTable(content, server string) (map[string]any, bool, error) {
	var doc map[string]any
	if err := toml.Unmarshal([]byte(content), &doc); err != nil {
		return nil, false, err
	}
	servers, _ := doc["mcp_servers"].(map[string]any)
	t, ok := servers[server].(map[string]any)
	return t, ok, nil
}

// dropCommandArgs returns the raw lines of the table body without the
// assignments of command and args (an args array may span several lines).
func dropCommandArgs(body []tomlLine) []tomlLine {
	var out []tomlLine
	skipping := false
	depth := 0
	for _, l := range body {
		if skipping {
			depth += strings.Count(l.text, "[") - strings.Count(l.text, "]")
			if depth <= 0 {
				skipping = false
			}
			continue
		}
		t := strings.TrimSpace(l.text)
		key := ""
		if i := strings.IndexByte(t, '='); i > 0 {
			key = strings.TrimSpace(t[:i])
		}
		if key == "command" {
			continue
		}
		if key == "args" {
			depth = strings.Count(t, "[") - strings.Count(t, "]")
			skipping = depth > 0
			continue
		}
		out = append(out, l)
	}
	return out
}

// RegisterCodex points [mcp_servers.<server>] at the proxy. It returns the
// content unchanged (changed=false) when the block is already there, and
// ErrEntryNotFound when the table does not exist.
func RegisterCodex(content, server string) (string, bool, error) {
	if err := filemerge.ValidateTOML(content); err != nil {
		return "", false, err
	}
	lines := splitLines(content)
	if _, _, found, err := findBlock(lines); err != nil {
		return "", false, err
	} else if found {
		return content, false, nil
	}
	start := -1
	for i, l := range lines {
		if isServerHeader(l.text, server) {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false, ErrEntryNotFound
	}
	end := start + 1
	for end < len(lines) && !isHeader(lines[end].text) {
		end++
	}
	orig, _, err := serverTable(content, server)
	if err != nil {
		return "", false, err
	}
	cmd, ok := orig["command"].(string)
	if !ok {
		return "", false, fmt.Errorf(`[mcp_servers.%s] has no string "command" (a remote server cannot be proxied)`, server)
	}
	var origArgs []string
	if raw, present := orig["args"]; present {
		list, ok := raw.([]any)
		if !ok {
			return "", false, fmt.Errorf(`[mcp_servers.%s] "args" is not an array`, server)
		}
		for _, v := range list {
			s, ok := v.(string)
			if !ok {
				return "", false, fmt.Errorf(`[mcp_servers.%s] "args" holds a non-string value`, server)
			}
			origArgs = append(origArgs, s)
		}
	}
	proxied := append(append([]string{}, prefix...), cmd)
	proxied = append(proxied, origArgs...)

	eol := lineEnding(content)
	var block []tomlLine
	add := func(text string) { block = append(block, tomlLine{text, eol}) }
	add(MarkerBegin)
	add(managedLine)
	for i := start; i < end; i++ {
		l := lines[i]
		if l.eol == "" {
			block = append(block, tomlLine{origNoEOLPrefix + l.text, eol})
		} else {
			block = append(block, tomlLine{origPrefix + l.text, l.eol})
		}
	}
	block = append(block, lines[start]) // the original header, reused as is
	if block[len(block)-1].eol == "" {
		block[len(block)-1].eol = eol
	}
	quoted := make([]string, len(proxied))
	for i, a := range proxied {
		quoted[i] = tomlString(a)
	}
	add("command = " + tomlString(Command))
	add("args = [" + strings.Join(quoted, ", ") + "]")
	for _, l := range dropCommandArgs(lines[start+1 : end]) {
		if l.eol == "" {
			l.eol = eol
		}
		block = append(block, l)
	}
	add(MarkerEnd)

	out := joinLines(lines[:start]) + joinLines(block) + joinLines(lines[end:])
	if err := filemerge.ValidateTOML(out); err != nil {
		return "", false, fmt.Errorf("the proxy entry would not be valid TOML: %w", err)
	}
	// safety net: command and args changed, every other key is as it was
	got, ok, err := serverTable(out, server)
	if err != nil || !ok || got["command"] != Command {
		return "", false, fmt.Errorf("the proxy entry could not be verified after the edit; the file was not changed")
	}
	want := map[string]any{}
	for k, v := range orig {
		if k != "command" && k != "args" {
			want[k] = v
		}
	}
	delete(got, "command")
	delete(got, "args")
	if !reflect.DeepEqual(got, want) {
		return "", false, fmt.Errorf("the other keys of [mcp_servers.%s] would change; the file was not changed", server)
	}
	return out, true, nil
}

// UnregisterCodex restores the original table. Content without the block is
// returned unchanged.
func UnregisterCodex(content, server string) (string, bool, error) {
	if err := filemerge.ValidateTOML(content); err != nil {
		return "", false, err
	}
	lines := splitLines(content)
	begin, end, found, err := findBlock(lines)
	if err != nil || !found {
		return content, false, err
	}
	var restored []tomlLine
	for _, l := range lines[begin+1 : end] {
		switch {
		case strings.HasPrefix(l.text, origNoEOLPrefix):
			restored = append(restored, tomlLine{strings.TrimPrefix(l.text, origNoEOLPrefix), ""})
		case strings.HasPrefix(l.text, origPrefix):
			restored = append(restored, tomlLine{strings.TrimPrefix(l.text, origPrefix), l.eol})
		}
	}
	if len(restored) == 0 {
		return "", false, fmt.Errorf("the proxy block has no original entry to restore; remove it by hand")
	}
	out := joinLines(lines[:begin]) + joinLines(restored) + joinLines(lines[end+1:])
	if err := filemerge.ValidateTOML(out); err != nil {
		return "", false, fmt.Errorf("the restored entry would not be valid TOML: %w", err)
	}
	return out, true, nil
}
