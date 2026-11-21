// Package jsonedit is a minimal JSON editor that changes only the bytes it has
// to. It exists because filemerge.MergeJSONObjects replaces arrays, treats
// invalid JSON as {} and rewrites the whole file: none of that is acceptable for
// a user's settings.json or .engram/config.json. Everything outside the
// inserted or deleted range stays byte for byte.
//
// It was extracted from internal/harness/hooks (R-16) so that R-17 can share it.
package jsonedit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Span is the byte range [Start, End) of a JSON value inside the source.
type Span struct{ Start, End int }

// Member is one key of an object, with where its key and its value start.
type Member struct {
	Key      string
	KeyStart int // first byte of the key token
	Val      Span
}

// Style is the formatting of the file, so that what is inserted looks like it.
type Style struct {
	Eol  string
	Unit string
}

func DetectStyle(src []byte) Style {
	st := Style{Eol: "\n", Unit: "  "}
	if bytes.Contains(src, []byte("\r\n")) {
		st.Eol = "\r\n"
	}
	// the first indented line sets the unit
	for _, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimRight(line, "\r")
		body := strings.TrimLeft(trimmed, " \t")
		if body == "" || len(body) == len(trimmed) {
			continue
		}
		st.Unit = trimmed[:len(trimmed)-len(body)]
		break
	}
	return st
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

// RootSpan returns the Span of the top-level value (which must be an object).
func RootSpan(src []byte) (Span, error) {
	if !json.Valid(src) {
		return Span{}, errors.New("not valid JSON (comments, trailing commas and a byte order mark are not accepted)")
	}
	start, end := 0, len(src)
	for start < end && isSpace(src[start]) {
		start++
	}
	for end > start && isSpace(src[end-1]) {
		end--
	}
	if start == end || src[start] != '{' {
		return Span{}, errors.New("the top level is not a JSON object")
	}
	return Span{start, end}, nil
}

func ParseObject(src []byte, v Span) ([]Member, error) {
	if v.End-v.Start < 2 || src[v.Start] != '{' {
		return nil, errors.New("not an object")
	}
	dec := json.NewDecoder(bytes.NewReader(src[v.Start:v.End]))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []Member
	prevEnd := v.Start + 1
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
		end := v.Start + int(dec.InputOffset())
		ks := prevEnd
		for ks < len(src) && (isSpace(src[ks]) || src[ks] == ',') {
			ks++
		}
		out = append(out, Member{Key: key, KeyStart: ks, Val: Span{end - len(raw), end}})
		prevEnd = end
	}
	return out, nil
}

func ParseArray(src []byte, v Span) ([]Span, error) {
	if v.End-v.Start < 2 || src[v.Start] != '[' {
		return nil, errors.New("not an array")
	}
	dec := json.NewDecoder(bytes.NewReader(src[v.Start:v.End]))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []Span
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := v.Start + int(dec.InputOffset())
		out = append(out, Span{end - len(raw), end})
	}
	return out, nil
}

// LastMember returns the last Member with that key (the one a JSON reader keeps).
func LastMember(ms []Member, key string) (int, bool) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Key == key {
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

func render(v any, childIndent string, st Style) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(childIndent, st.Unit)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	text := strings.TrimRight(buf.String(), "\n")
	return strings.ReplaceAll(text, "\n", st.Eol), nil
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

// AddMember appends "key": value to the object, formatted like its neighbours.
func AddMember(src []byte, obj Span, key string, value any, st Style) ([]byte, error) {
	ms, err := ParseObject(src, obj)
	if err != nil {
		return nil, err
	}
	parent := lineIndent(src, obj.Start)
	if len(ms) == 0 {
		child := parent + st.Unit
		text, err := render(value, child, st)
		if err != nil {
			return nil, err
		}
		return splice(src, obj.Start, obj.End,
			"{"+st.Eol+child+quote(key)+": "+text+st.Eol+parent+"}"), nil
	}
	child := lineIndent(src, ms[0].KeyStart)
	if sameLine(src, obj.Start, ms[0].KeyStart) {
		child = parent + st.Unit
	}
	text, err := render(value, child, st)
	if err != nil {
		return nil, err
	}
	last := ms[len(ms)-1]
	return splice(src, last.Val.End, last.Val.End,
		","+st.Eol+child+quote(key)+": "+text), nil
}

// RemoveMember deletes Member idx; the inverse of AddMember.
func RemoveMember(src []byte, obj Span, idx int) ([]byte, error) {
	ms, err := ParseObject(src, obj)
	if err != nil {
		return nil, err
	}
	switch {
	case len(ms) == 1:
		return splice(src, obj.Start+1, obj.End-1, ""), nil
	case idx > 0:
		return splice(src, ms[idx-1].Val.End, ms[idx].Val.End, ""), nil
	default:
		return splice(src, ms[0].KeyStart, ms[1].KeyStart, ""), nil
	}
}

// AddElem appends a value to the array, formatted like its neighbours.
func AddElem(src []byte, arr Span, value any, st Style) ([]byte, error) {
	es, err := ParseArray(src, arr)
	if err != nil {
		return nil, err
	}
	parent := lineIndent(src, arr.Start)
	if len(es) == 0 {
		child := parent + st.Unit
		text, err := render(value, child, st)
		if err != nil {
			return nil, err
		}
		return splice(src, arr.Start, arr.End,
			"["+st.Eol+child+text+st.Eol+parent+"]"), nil
	}
	child := lineIndent(src, es[0].Start)
	if sameLine(src, arr.Start, es[0].Start) {
		child = parent + st.Unit
	}
	text, err := render(value, child, st)
	if err != nil {
		return nil, err
	}
	return splice(src, es[len(es)-1].End, es[len(es)-1].End, ","+st.Eol+child+text), nil
}

// RemoveElem deletes element idx; the inverse of AddElem.
func RemoveElem(src []byte, arr Span, idx int) ([]byte, error) {
	es, err := ParseArray(src, arr)
	if err != nil {
		return nil, err
	}
	switch {
	case len(es) == 1:
		return splice(src, arr.Start+1, arr.End-1, ""), nil
	case idx > 0:
		return splice(src, es[idx-1].End, es[idx].End, ""), nil
	default:
		return splice(src, es[0].Start, es[1].Start, ""), nil
	}
}

var ErrShape = errors.New("unexpected structure")

func ShapeErr(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrShape, fmt.Sprintf(format, a...))
}
