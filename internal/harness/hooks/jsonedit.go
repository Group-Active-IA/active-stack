package hooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// A minimal JSON editor that changes only the bytes it has to. It exists
// because filemerge.MergeJSONObjects replaces arrays, treats invalid JSON as
// {} and rewrites the whole file: none of that is acceptable for a user's
// settings.json. Everything outside the inserted or deleted range stays
// byte for byte.

// span is the byte range [start, end) of a JSON value inside the source.
type span struct{ start, end int }

type member struct {
	key      string
	keyStart int // first byte of the key token
	val      span
}

// style is the formatting of the file, so that what is inserted looks like it.
type style struct {
	eol  string
	unit string
}

func detectStyle(src []byte) style {
	st := style{eol: "\n", unit: "  "}
	if bytes.Contains(src, []byte("\r\n")) {
		st.eol = "\r\n"
	}
	// the first indented line sets the unit
	for _, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimRight(line, "\r")
		body := strings.TrimLeft(trimmed, " \t")
		if body == "" || len(body) == len(trimmed) {
			continue
		}
		st.unit = trimmed[:len(trimmed)-len(body)]
		break
	}
	return st
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

// rootSpan returns the span of the top-level value (which must be an object).
func rootSpan(src []byte) (span, error) {
	if !json.Valid(src) {
		return span{}, errors.New("not valid JSON (comments, trailing commas and a byte order mark are not accepted)")
	}
	start, end := 0, len(src)
	for start < end && isSpace(src[start]) {
		start++
	}
	for end > start && isSpace(src[end-1]) {
		end--
	}
	if start == end || src[start] != '{' {
		return span{}, errors.New("the top level is not a JSON object")
	}
	return span{start, end}, nil
}

func parseObject(src []byte, v span) ([]member, error) {
	if v.end-v.start < 2 || src[v.start] != '{' {
		return nil, errors.New("not an object")
	}
	dec := json.NewDecoder(bytes.NewReader(src[v.start:v.end]))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []member
	prevEnd := v.start + 1
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := kt.(string)
		if !ok {
			return nil, errors.New("object key is not a string")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := v.start + int(dec.InputOffset())
		ks := prevEnd
		for ks < len(src) && (isSpace(src[ks]) || src[ks] == ',') {
			ks++
		}
		out = append(out, member{key: key, keyStart: ks, val: span{end - len(raw), end}})
		prevEnd = end
	}
	return out, nil
}

func parseArray(src []byte, v span) ([]span, error) {
	if v.end-v.start < 2 || src[v.start] != '[' {
		return nil, errors.New("not an array")
	}
	dec := json.NewDecoder(bytes.NewReader(src[v.start:v.end]))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []span
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := v.start + int(dec.InputOffset())
		out = append(out, span{end - len(raw), end})
	}
	return out, nil
}

// lastMember returns the last member with that key (the one a JSON reader keeps).
func lastMember(ms []member, key string) (int, bool) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].key == key {
			return i, true
		}
	}
	return 0, false
}

func lineIndent(src []byte, pos int) string {
	ls := bytes.LastIndexByte(src[:pos], '\n') + 1
	i := ls
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	return string(src[ls:i])
}

func sameLine(src []byte, a, b int) bool { return !bytes.ContainsAny(src[a:b], "\r\n") }

func render(v any, childIndent string, st style) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(childIndent, st.unit)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	text := strings.TrimRight(buf.String(), "\n")
	return strings.ReplaceAll(text, "\n", st.eol), nil
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func splice(src []byte, start, end int, repl string) []byte {
	out := make([]byte, 0, len(src)-(end-start)+len(repl))
	out = append(out, src[:start]...)
	out = append(out, repl...)
	out = append(out, src[end:]...)
	return out
}

// addMember appends "key": value to the object, formatted like its neighbours.
func addMember(src []byte, obj span, key string, value any, st style) ([]byte, error) {
	ms, err := parseObject(src, obj)
	if err != nil {
		return nil, err
	}
	parent := lineIndent(src, obj.start)
	if len(ms) == 0 {
		child := parent + st.unit
		text, err := render(value, child, st)
		if err != nil {
			return nil, err
		}
		return splice(src, obj.start, obj.end,
			"{"+st.eol+child+quote(key)+": "+text+st.eol+parent+"}"), nil
	}
	child := lineIndent(src, ms[0].keyStart)
	if sameLine(src, obj.start, ms[0].keyStart) {
		child = parent + st.unit
	}
	text, err := render(value, child, st)
	if err != nil {
		return nil, err
	}
	last := ms[len(ms)-1]
	return splice(src, last.val.end, last.val.end,
		","+st.eol+child+quote(key)+": "+text), nil
}

// removeMember deletes member idx; the inverse of addMember.
func removeMember(src []byte, obj span, idx int) ([]byte, error) {
	ms, err := parseObject(src, obj)
	if err != nil {
		return nil, err
	}
	switch {
	case len(ms) == 1:
		return splice(src, obj.start+1, obj.end-1, ""), nil
	case idx > 0:
		return splice(src, ms[idx-1].val.end, ms[idx].val.end, ""), nil
	default:
		return splice(src, ms[0].keyStart, ms[1].keyStart, ""), nil
	}
}

// addElem appends a value to the array, formatted like its neighbours.
func addElem(src []byte, arr span, value any, st style) ([]byte, error) {
	es, err := parseArray(src, arr)
	if err != nil {
		return nil, err
	}
	parent := lineIndent(src, arr.start)
	if len(es) == 0 {
		child := parent + st.unit
		text, err := render(value, child, st)
		if err != nil {
			return nil, err
		}
		return splice(src, arr.start, arr.end,
			"["+st.eol+child+text+st.eol+parent+"]"), nil
	}
	child := lineIndent(src, es[0].start)
	if sameLine(src, arr.start, es[0].start) {
		child = parent + st.unit
	}
	text, err := render(value, child, st)
	if err != nil {
		return nil, err
	}
	return splice(src, es[len(es)-1].end, es[len(es)-1].end, ","+st.eol+child+text), nil
}

// removeElem deletes element idx; the inverse of addElem.
func removeElem(src []byte, arr span, idx int) ([]byte, error) {
	es, err := parseArray(src, arr)
	if err != nil {
		return nil, err
	}
	switch {
	case len(es) == 1:
		return splice(src, arr.start+1, arr.end-1, ""), nil
	case idx > 0:
		return splice(src, es[idx-1].end, es[idx].end, ""), nil
	default:
		return splice(src, es[0].start, es[1].start, ""), nil
	}
}

var errShape = errors.New("unexpected structure")

func shapeErr(format string, a ...any) error {
	return fmt.Errorf("%w: %s", errShape, fmt.Sprintf(format, a...))
}
