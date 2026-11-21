package projectinit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/jsonedit"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// planJSON returns the new content of a JSON file so that it holds key = want,
// changing only the bytes it has to. next equal to raw means nothing to do.
//
//   - missing or blank file: a new two-line object;
//   - the key is absent: it is appended, formatted like its neighbours;
//   - the key has the same value: unchanged;
//   - the key has another value (or not a string): a ConflictError, never a write;
//   - anything that is not a JSON object (comments and trailing commas included):
//     an InvalidFileError.
//
// A UTF-8 BOM (Notepad adds one) is accepted and kept.
func planJSON(rel string, raw []byte, existed bool, key, want string) ([]byte, error) {
	body, hasBOM := raw, false
	if bytes.HasPrefix(raw, utf8BOM) {
		body, hasBOM = raw[len(utf8BOM):], true
	}
	withBOM := func(b []byte) []byte {
		if hasBOM {
			return append(append([]byte{}, utf8BOM...), b...)
		}
		return b
	}
	if !existed || len(bytes.TrimSpace(body)) == 0 {
		return withBOM([]byte(fmt.Sprintf("{\n  %s: %s\n}\n", quote(key), quote(want)))), nil
	}
	root, err := jsonedit.RootSpan(body)
	if err != nil {
		return nil, &InvalidFileError{File: rel, Err: err}
	}
	ms, err := jsonedit.ParseObject(body, root)
	if err != nil {
		return nil, &InvalidFileError{File: rel, Err: err}
	}
	if i, ok := jsonedit.LastMember(ms, key); ok {
		val := body[ms[i].Val.Start:ms[i].Val.End]
		var have string
		if err := json.Unmarshal(val, &have); err != nil {
			return nil, &ConflictError{File: rel, Reason: ReasonValue, Key: key, Have: string(val), Want: want}
		}
		if have != want {
			return nil, &ConflictError{File: rel, Reason: ReasonValue, Key: key, Have: have, Want: want}
		}
		return raw, nil
	}
	out, err := jsonedit.AddMember(body, root, key, want, jsonedit.DetectStyle(body))
	if err != nil {
		return nil, &InvalidFileError{File: rel, Err: err}
	}
	return withBOM(out), nil
}

func quote(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimRight(buf.String(), "\n")
}

// gitignoreBlock is the rules, between markers, in the file's line ending.
func gitignoreBlock(eol string) string {
	return strings.Join([]string{
		BlockBegin,
		".engram/*",
		"!" + ConfigPath,
		"!" + MarkerPath,
		BlockEnd,
	}, eol) + eol
}

// planGitignore returns the .gitignore with our block added or refreshed. The
// person's lines are never edited or reordered. If one of them ignores the
// whole .engram/ folder, Git cannot re-include the two files, so it is a
// conflict and nothing is written.
func planGitignore(raw []byte, existed bool) ([]byte, error) {
	content := string(raw)
	eol := "\n"
	if strings.Contains(content, "\r\n") {
		eol = "\r\n"
	}
	loc, err := findBlock(content)
	if err != nil {
		return nil, &InvalidFileError{File: GitignorePath, Err: err}
	}
	outside := content
	if loc.found {
		outside = content[:loc.start] + content[loc.end:]
	}
	for _, line := range strings.Split(outside, "\n") {
		if ignoresWholeDir(line) {
			return nil, &ConflictError{File: GitignorePath, Reason: ReasonIgnoredDir, Have: strings.TrimSpace(line)}
		}
	}
	block := gitignoreBlock(eol)
	if loc.found {
		return []byte(content[:loc.start] + block + content[loc.end:]), nil
	}
	sep := ""
	switch {
	case content == "":
	case strings.HasSuffix(content, "\n"):
		sep = eol
	default:
		sep = eol + eol
	}
	return []byte(content + sep + block), nil
}

// ignoresWholeDir reports whether a .gitignore line ignores the .engram folder
// itself (not just what is inside it).
func ignoresWholeDir(line string) bool {
	l := strings.TrimSpace(line)
	if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "!") {
		return false
	}
	l = strings.TrimPrefix(l, "**/")
	l = strings.TrimPrefix(l, "/")
	l = strings.TrimSuffix(l, "/")
	return l == ".engram"
}

type markerLines struct {
	found      bool
	start, end int
}

// findBlock locates our block by its marker lines: the offset of the start of
// the begin line and the end of the end line (with its line break).
func findBlock(content string) (markerLines, error) {
	var begins, ends, endLineEnds []int
	pos := 0
	for pos < len(content) {
		next := len(content)
		lineEnd := len(content)
		if nl := strings.IndexByte(content[pos:], '\n'); nl >= 0 {
			lineEnd, next = pos+nl, pos+nl+1
		}
		switch strings.TrimSpace(content[pos:lineEnd]) {
		case BlockBegin:
			begins = append(begins, pos)
		case BlockEnd:
			ends = append(ends, pos)
			endLineEnds = append(endLineEnds, next)
		}
		pos = next
	}
	switch {
	case len(begins) == 0 && len(ends) == 0:
		return markerLines{}, nil
	case len(begins) != 1 || len(ends) != 1:
		return markerLines{}, fmt.Errorf("found %d begin and %d end marker lines, want exactly one of each", len(begins), len(ends))
	case ends[0] < begins[0]:
		return markerLines{}, errors.New("the end marker comes before the begin marker")
	}
	return markerLines{found: true, start: begins[0], end: endLineEnds[0]}, nil
}
