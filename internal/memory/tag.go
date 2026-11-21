package memory

import (
	"os/user"
	"regexp"
	"strings"
	"unicode"
)

// Engram MCP tool prefixes: the plain MCP server and the Claude plugin.
var toolPrefixes = []string{"mcp__engram__", "mcp__plugin_engram_engram__"}

// narrativeFields maps each narrative write tool to its text field (R15-2,
// from the R-05 prototype). They get the header but keep key and title.
var narrativeFields = map[string]string{
	"mem_session_summary":  "content",
	"mem_session_finalize": "summary",
	"mem_capture_passive":  "content",
	"mem_save_prompt":      "content",
}

// listItem matches the bullet of a Markdown list item ("- ", "* ", "1. ", "2) ").
var listItem = regexp.MustCompile(`^(\s*(?:[-*]|\d+[.)])\s+)(.*)$`)

const maxAuthor = 64

// SaveContext is what TagInput needs to tag a save: the Detect result for the
// project and the author name.
type SaveContext struct {
	Result Result
	Author string
}

// ToolBase strips the Engram MCP prefixes from a tool name. A bare name (as
// the MCP proxy sees it) is returned as is; another server's tool gives "".
func ToolBase(tool string) string {
	for _, prefix := range toolPrefixes {
		if strings.HasPrefix(tool, prefix) {
			return strings.TrimPrefix(tool, prefix)
		}
	}
	if strings.HasPrefix(tool, "mcp__") {
		return ""
	}
	return tool
}

// TagInput returns a copy of input tagged for ctx ([EN] §3.3) and whether
// anything changed. It never modifies input and keeps every field it does not
// tag, because the agents replace the whole tool input with the result.
//
//   - mode off: nothing changes;
//   - mem_save: on a branch, topic_key becomes branch/<slug>/K (only when K is
//     set), the title gets "[<slug>] " and the content (or its alias
//     observation) starts with "Branch: <branch> · Author: <name>"; otherwise
//     only the "Author: <name>" line is added;
//   - narrative tools: the same header in their text field; passive capture
//     instead tags each list item, and only on a branch.
//
// Tagging is idempotent: an already tagged field is left alone.
func TagInput(tool string, input map[string]any, ctx SaveContext) (map[string]any, bool) {
	if ctx.Result.Mode != ModeCloudBranch {
		return input, false
	}
	base := ToolBase(tool)
	out := make(map[string]any, len(input))
	for k, v := range input {
		out[k] = v
	}
	r := ctx.Result
	header := "Author: " + ctx.Author
	if r.Tagged {
		header = "Branch: " + r.Branch + " · " + header
	}

	changed := false
	switch field, narrative := narrativeFields[base]; {
	case base == "mem_save":
		if r.Tagged {
			if key, ok := out["topic_key"].(string); ok && key != "" && !strings.HasPrefix(key, "branch/"+r.Slug+"/") {
				out["topic_key"] = "branch/" + r.Slug + "/" + key
				changed = true
			}
			if title, ok := out["title"].(string); ok && !strings.HasPrefix(title, "["+r.Slug+"] ") {
				out["title"] = "[" + r.Slug + "] " + title
				changed = true
			}
		}
		field := "content"
		if _, ok := out[field].(string); !ok {
			field = "observation"
		}
		changed = addHeader(out, field, header) || changed
	case narrative && base == "mem_capture_passive":
		if text, ok := out[field].(string); ok && r.Tagged {
			if tagged := tagListItems(text, r.Slug); tagged != text {
				out[field] = tagged
				changed = true
			}
		}
	case narrative:
		changed = addHeader(out, field, header)
	}
	if !changed {
		return input, false
	}
	return out, true
}

// addHeader puts header as the first line of the string field, unless the
// field is missing or already starts with a Branch or Author line.
func addHeader(fields map[string]any, field, header string) bool {
	text, ok := fields[field].(string)
	if !ok || strings.HasPrefix(text, "Branch: ") || strings.HasPrefix(text, "Author: ") {
		return false
	}
	fields[field] = header + "\n" + text
	return true
}

func tagListItems(text, slug string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		m := listItem.FindStringSubmatch(line)
		if m != nil && !strings.HasPrefix(m[2], "["+slug+"] ") {
			lines[i] = m[1] + "[" + slug + "] " + m[2]
		}
	}
	return strings.Join(lines, "\n")
}

// SanitizeAuthor cleans a name for the Author line: one line, single spaces,
// no "·" (the header separator), at most 64 characters, and the part after a
// Windows domain ("CORP\ana" → "ana"). Anything that looks like an email
// gives "": the email is never written.
func SanitizeAuthor(name string) string {
	if strings.Contains(name, "@") {
		return ""
	}
	if i := strings.LastIndex(name, `\`); i >= 0 {
		name = name[i+1:]
	}
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '·' {
			return ' '
		}
		return r
	}, name)
	name = strings.Join(strings.Fields(name), " ")
	if runes := []rune(name); len(runes) > maxAuthor {
		name = strings.TrimSpace(string(runes[:maxAuthor]))
	}
	return name
}

// ResolveAuthor returns the author for saves made in dir: git user.name, or
// the system user when it is missing or unusable, or "unknown".
func ResolveAuthor(dir string) string {
	if name, err := (gitRunner{dir: dir}).run("config", "user.name"); err == nil {
		if clean := SanitizeAuthor(name); clean != "" {
			return clean
		}
	}
	if current, err := user.Current(); err == nil {
		if clean := SanitizeAuthor(current.Username); clean != "" {
			return clean
		}
	}
	return "unknown"
}

// PrepareSave detects the memory mode of dir and tags input for tool. It is
// the shared core of the save hook and of the MCP proxy. On error the input
// is returned unchanged.
func PrepareSave(dir, tool string, input map[string]any) (map[string]any, bool, error) {
	res, err := Detect(dir)
	if err != nil {
		return input, false, err
	}
	if res.Mode != ModeCloudBranch {
		return input, false, nil
	}
	out, changed := TagInput(tool, input, SaveContext{Result: res, Author: ResolveAuthor(dir)})
	return out, changed, nil
}

// SessionContext is the line the session-start hook adds in projects with
// the marker ([EN] §3.2), or "" when the mode is off.
func SessionContext(r Result) string {
	if r.Mode != ModeCloudBranch {
		return ""
	}
	return "Team memory: cloud-branch. Current branch: " + branchText(r)
}

// ReadContext is the line the pre-read hook adds before a search ([EN] §3.6):
// only with the marker and on a tagged branch, otherwise "".
func ReadContext(r Result) string {
	if r.Mode != ModeCloudBranch || !r.Tagged {
		return ""
	}
	return "Current branch: " + branchText(r)
}

func branchText(r Result) string {
	if r.Tagged {
		return r.Branch + " (slug: " + r.Slug + ")."
	}
	return r.Branch + " (untagged)."
}
