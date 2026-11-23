package hooks

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Group-Active-IA/active-stack/internal/filemerge"
	toml "github.com/pelletier/go-toml/v2"
)

// Codex: ~/.codex/config.toml, inline [[hooks.<Event>]] with nested
// [[hooks.<Event>.hooks]]. Our entries live in one block between two marker
// lines, appended at the end of the file. The generic TOML helpers of
// filemerge find tables by exact name and do not understand arrays of tables,
// so the block is located by its markers and nothing outside it is rewritten.
//
// The hook is declared only here, never in hooks.json too: Codex loads both and
// would run it twice. Codex also asks the person to review and trust a hook
// (/hooks) before it runs; the command carries no path or version so that
// updating the binary does not invalidate the trust that was given.

func codexBlock(eol string) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s + eol) }
	line(EntryMarkerBegin)
	line("# Managed by active-stack: reinstalling rewrites this block, removing it removes the hooks.")
	for i, s := range Specs() {
		if i > 0 {
			line("")
		}
		line("[[hooks." + s.Event + "]]")
		if s.Matcher != "" {
			line("matcher = '" + s.Matcher + "'")
		}
		line("")
		line("[[hooks." + s.Event + ".hooks]]")
		line(`type = "command"`)
		line("command = '" + codexCommand(s) + "'")
		line(fmt.Sprintf("timeout = %d", s.TimeoutSeconds))
	}
	line(EntryMarkerEnd)
	return b.String()
}

func codexCommand(s Spec) string { return Command + " memory hook " + s.Subcommand }

// markerLines locates the block: byte offsets of the start of the begin line
// and the end of the end line (including its line break).
type markerLines struct {
	found      bool
	start, end int
}

func findBlock(content string) (markerLines, error) {
	var begins, ends []int // line start offsets
	var endLineEnds []int
	pos := 0
	for pos < len(content) {
		nl := strings.IndexByte(content[pos:], '\n')
		lineEnd := len(content)
		next := len(content)
		if nl >= 0 {
			lineEnd = pos + nl
			next = pos + nl + 1
		}
		text := strings.TrimSpace(content[pos:lineEnd])
		switch text {
		case EntryMarkerBegin:
			begins = append(begins, pos)
		case EntryMarkerEnd:
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

func lineEnding(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func validateTOML(content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	return filemerge.ValidateTOML(content)
}

// codexInstall returns the content with our block added or refreshed.
func codexInstall(content string) (string, error) {
	if err := validateTOML(content); err != nil {
		return "", err
	}
	eol := lineEnding(content)
	block := codexBlock(eol)
	loc, err := findBlock(content)
	if err != nil {
		return "", err
	}
	var out string
	if loc.found {
		// Codex may write runtime trust tables inside our marker block. When
		// definitions are unchanged, preserve every byte, including trust.
		if sameCodexDefinitions(content[loc.start:loc.end], block) {
			return content, nil
		}
		out = content[:loc.start] + block + content[loc.end:]
	} else {
		sep := ""
		switch {
		case content == "":
		case strings.HasSuffix(content, "\n"):
			sep = eol // a blank line before the block
		default:
			sep = eol + eol
		}
		out = content + sep + block
	}
	if err := validateTOML(out); err != nil {
		return "", fmt.Errorf("the hooks would conflict with what is already declared: %w", err)
	}
	return out, nil
}

// codexRemove deletes the block, and the blank line Install put before it.
func codexRemove(content string) (string, error) {
	if err := validateTOML(content); err != nil {
		return "", err
	}
	loc, err := findBlock(content)
	if err != nil || !loc.found {
		return content, err
	}
	start := loc.start
	// one blank line directly before the block was ours
	if start > 0 {
		prev := start - 1 // the '\n' that ends the previous line
		if content[prev] == '\n' {
			ls := strings.LastIndexByte(content[:prev], '\n') + 1
			if strings.TrimRight(content[ls:prev], "\r") == "" {
				start = ls
			}
		}
	}
	return content[:start] + content[loc.end:], nil
}

// codexMissing lists the subcommands whose entry is not in our block.
func codexMissing(content string) ([]string, error) {
	if err := validateTOML(content); err != nil {
		return nil, err
	}
	loc, err := findBlock(content)
	if err != nil {
		return nil, err
	}
	block := ""
	if loc.found {
		block = content[loc.start:loc.end]
	}
	var missing []string
	for _, s := range Specs() {
		if !strings.Contains(block, "command = '"+codexCommand(s)+"'") {
			missing = append(missing, s.Subcommand)
		}
	}
	return missing, nil
}

// sameCodexDefinitions ignores only Codex's runtime trust state. A changed
// command, matcher, timeout or other definition still refreshes the block.
func sameCodexDefinitions(current, expected string) bool {
	var got, want map[string]any
	if toml.Unmarshal([]byte(current), &got) != nil || toml.Unmarshal([]byte(expected), &want) != nil {
		return false
	}
	if hooks, ok := got["hooks"].(map[string]any); ok {
		delete(hooks, "state")
	}
	return reflect.DeepEqual(got, want)
}
